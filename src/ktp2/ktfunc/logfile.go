package ktfunc

// File logging + log bundling.
//
// Operators run the node headless and, when something misbehaves, have no easy
// way to send us what happened. SetupFileLogging mirrors everything logged to
// stdout into a timestamped file under a log directory (rotating once a file
// grows past maxLogSize), and ZipLogs bundles the recent files into a single
// zip an operator can attach to a bug report.

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

// maxLogSize is the size a single log file may reach before the writer rolls
// over to a fresh file. 50 MB keeps individual files small enough to send while
// rare enough to roll that the overhead is negligible.
const maxLogSize int64 = 50 * 1024 * 1024

// logZipMaxAge is how far back ZipLogs reaches when bundling files: logs older
// than this are left out so a report stays small and relevant.
const logZipMaxAge = 7 * 24 * time.Hour

// rotatingFileWriter is an io.Writer that appends to a timestamped file in a
// directory and starts a new file once the current one passes maxSize. Safe for
// concurrent use (logrus may write from multiple goroutines).
type rotatingFileWriter struct {
	mu      sync.Mutex
	dir     string
	maxSize int64
	clock   func() time.Time // injectable for tests

	file    *os.File
	path    string
	written int64
	seq     int
}

func newRotatingFileWriter(dir string, maxSize int64) (*rotatingFileWriter, error) {
	w := &rotatingFileWriter{dir: dir, maxSize: maxSize, clock: time.Now}
	if err := w.rotate(); err != nil {
		return nil, err
	}
	return w, nil
}

// rotate closes the current file (if any) and opens a new one. The sequence
// suffix guarantees a unique name even if two rotations land in the same second.
func (w *rotatingFileWriter) rotate() error {
	if w.file != nil {
		_ = w.file.Close()
	}
	if err := os.MkdirAll(w.dir, 0755); err != nil {
		return fmt.Errorf("failed to create log dir %s: %w", w.dir, err)
	}
	name := fmt.Sprintf("ktoc-%s-%d.log", w.clock().Format("20060102-150405"), w.seq)
	w.seq++
	path := filepath.Join(w.dir, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w", path, err)
	}
	w.file = f
	w.path = path
	w.written = 0
	return nil
}

func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.written > 0 && w.written+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}
	n, err := w.file.Write(p)
	w.written += int64(n)
	return n, err
}

// SetupFileLogging makes logrus mirror its output into a rotating file under
// logDir while still writing to stdout. Returns the path of the first log file.
func SetupFileLogging(logDir string) (string, error) {
	w, err := newRotatingFileWriter(logDir, maxLogSize)
	if err != nil {
		return "", err
	}
	log.SetOutput(io.MultiWriter(os.Stdout, w))
	return w.path, nil
}

// ZipLogs bundles every *.log file in logDir modified within maxAge of now,
// plus a meta.txt holding the provided metadata, into a zip written to outDir.
// Returns the zip's path. host names the file so multiple operators' bundles
// don't collide.
func ZipLogs(logDir, outDir, host, meta string, now time.Time, maxAge time.Duration) (string, error) {
	entries, err := os.ReadDir(logDir)
	if os.IsNotExist(err) {
		// No log directory yet (e.g. -zipLogs before the node ever ran). Produce
		// a bundle with just the metadata rather than failing.
		log.Warnf("Log directory %s does not exist; bundling metadata only", logDir)
		entries = nil
	} else if err != nil {
		return "", fmt.Errorf("failed to read log dir %s: %w", logDir, err)
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create output dir %s: %w", outDir, err)
	}
	zipName := fmt.Sprintf("ktoc-logs-%s-%s.zip", sanitizeForFilename(host), now.Format("20060102-150405"))
	zipPath := filepath.Join(outDir, zipName)
	out, err := os.Create(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to create zip %s: %w", zipPath, err)
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	mw, err := zw.Create("meta.txt")
	if err != nil {
		return "", err
	}
	if _, err := mw.Write([]byte(meta)); err != nil {
		return "", err
	}

	cutoff := now.Add(-maxAge)
	included := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".log") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			continue // too old; keep the bundle small and relevant
		}
		if err := addFileToZip(zw, logDir, e.Name()); err != nil {
			return "", err
		}
		included++
	}
	log.Infof("Bundled %d log file(s) into %s", included, zipPath)
	return zipPath, nil
}

func addFileToZip(zw *zip.Writer, dir, name string) error {
	f, err := os.Open(filepath.Join(dir, name))
	if err != nil {
		return err
	}
	defer f.Close()
	w, err := zw.Create(name)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, f)
	return err
}

// sanitizeForFilename strips characters that don't belong in a filename so a
// hostname can be embedded safely.
func sanitizeForFilename(s string) string {
	if s == "" {
		return "host"
	}
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			return r
		default:
			return '-'
		}
	}, s)
}

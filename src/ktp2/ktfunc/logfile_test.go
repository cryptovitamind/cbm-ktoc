package ktfunc

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMaxLogSizeIs50MB(t *testing.T) {
	assert.Equal(t, int64(50*1024*1024), maxLogSize)
}

func TestSetupFileLogging_WritesToTimestampedFile(t *testing.T) {
	dir := t.TempDir()

	// Preserve and restore the global logrus output so other tests aren't
	// affected.
	orig := logrus.StandardLogger().Out
	t.Cleanup(func() { logrus.SetOutput(orig) })

	path, err := SetupFileLogging(dir)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(filepath.Base(path), "ktoc-"))
	assert.True(t, strings.HasSuffix(path, ".log"))

	logrus.SetLevel(logrus.InfoLevel)
	logrus.Info("hello-from-test-logline")

	data, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(data), "hello-from-test-logline")
}

func TestRotatingFileWriter_RotatesWhenSizeExceeded(t *testing.T) {
	dir := t.TempDir()
	w, err := newRotatingFileWriter(dir, 1000) // tiny cap to force a rotation
	require.NoError(t, err)

	chunk := make([]byte, 600)
	for i := range chunk {
		chunk[i] = 'x'
	}
	// First write fits; second write would exceed 1000 → rotation to a new file.
	_, err = w.Write(chunk)
	require.NoError(t, err)
	_, err = w.Write(chunk)
	require.NoError(t, err)

	logs, err := filepath.Glob(filepath.Join(dir, "ktoc-*.log"))
	require.NoError(t, err)
	assert.Len(t, logs, 2, "writer should have rolled over to a second file")
}

func TestZipLogs_BundlesOnlyRecentLogsPlusMeta(t *testing.T) {
	logDir := t.TempDir()
	outDir := t.TempDir()
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

	// Two recent logs and one older than the 7-day window.
	recentA := filepath.Join(logDir, "ktoc-recent-a.log")
	recentB := filepath.Join(logDir, "ktoc-recent-b.log")
	old := filepath.Join(logDir, "ktoc-old.log")
	for _, p := range []string{recentA, recentB, old} {
		require.NoError(t, os.WriteFile(p, []byte("log "+filepath.Base(p)), 0644))
	}
	// Also a non-log file that must be ignored.
	require.NoError(t, os.WriteFile(filepath.Join(logDir, "notes.txt"), []byte("ignore me"), 0644))

	recentTime := now.Add(-1 * time.Hour)
	oldTime := now.Add(-8 * 24 * time.Hour)
	require.NoError(t, os.Chtimes(recentA, recentTime, recentTime))
	require.NoError(t, os.Chtimes(recentB, recentTime, recentTime))
	require.NoError(t, os.Chtimes(old, oldTime, oldTime))

	meta := "banner=v0.4.5-beta\nos=testos"
	zipPath, err := ZipLogs(logDir, outDir, "node1", meta, now, logZipMaxAge)
	require.NoError(t, err)

	names, contents := readZip(t, zipPath)
	assert.Contains(t, names, "meta.txt")
	assert.Contains(t, names, "ktoc-recent-a.log")
	assert.Contains(t, names, "ktoc-recent-b.log")
	assert.NotContains(t, names, "ktoc-old.log", "logs older than the window must be excluded")
	assert.NotContains(t, names, "notes.txt", "non-log files must be excluded")
	assert.Equal(t, meta, contents["meta.txt"])
}

func TestZipLogs_MissingLogDirBundlesMetaOnly(t *testing.T) {
	logrus.SetLevel(logrus.FatalLevel)
	outDir := t.TempDir()
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	now := time.Date(2026, 6, 17, 12, 0, 0, 0, time.UTC)

	zipPath, err := ZipLogs(missing, outDir, "node1", "banner=v0.4.5-beta", now, logZipMaxAge)
	require.NoError(t, err)

	names, contents := readZip(t, zipPath)
	assert.Equal(t, []string{"meta.txt"}, names)
	assert.Equal(t, "banner=v0.4.5-beta", contents["meta.txt"])
}

func readZip(t *testing.T, path string) ([]string, map[string]string) {
	t.Helper()
	r, err := zip.OpenReader(path)
	require.NoError(t, err)
	defer r.Close()
	var names []string
	contents := make(map[string]string)
	for _, f := range r.File {
		names = append(names, f.Name)
		rc, err := f.Open()
		require.NoError(t, err)
		b, err := io.ReadAll(rc)
		require.NoError(t, err)
		rc.Close()
		contents[f.Name] = string(b)
	}
	return names, contents
}

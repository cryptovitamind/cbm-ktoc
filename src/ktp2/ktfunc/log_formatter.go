package ktfunc

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// CustomFormatter provides colorful, artsy logging without timestamps.
type CustomFormatter struct{}

func LogOperationStart(operation string) {
	timestamp := time.Now().Format("15:04:05")
	log.Info("")
	log.Infof("---- [%s] %s ----", timestamp, strings.ToUpper(operation))
}

// Format implements the logrus Formatter interface.
func (f *CustomFormatter) Format(entry *log.Entry) ([]byte, error) {
	var buf bytes.Buffer
	levelColor := ""
	switch entry.Level {
	case log.InfoLevel:
		levelColor = "\033[1;32m" // Bright Green
	case log.WarnLevel:
		levelColor = "\033[1;33m" // Bright Yellow
	case log.ErrorLevel:
		levelColor = "\033[1;31m" // Bright Red
	case log.FatalLevel:
		levelColor = "\033[1;35m" // Bright Magenta
	default:
		levelColor = "\033[1;37m" // Bright White
	}

	resetColor := "\033[0m"
	valueColor := "\033[1;36m" // Cyan for values

	// Add a decorative prefix based on level
	prefix := "💕 "
	if entry.Level == log.ErrorLevel {
		prefix = "😡 "
	} else if entry.Level == log.FatalLevel {
		prefix = "☠️ "
	} else if entry.Level == log.WarnLevel {
		prefix = "😬 "
	}

	// Split message to colorize values after colons
	msgParts := strings.SplitN(entry.Message, ": ", 2)

	if entry.Message == "" {
		buf.WriteByte('\n')
	} else if len(msgParts) == 2 {
		buf.WriteString(fmt.Sprintf("%s%s %s%s: %s%s%s\n",
			levelColor, prefix, msgParts[0], resetColor,
			valueColor, msgParts[1], resetColor))
	} else {
		buf.WriteString(fmt.Sprintf("%s%s➤ %s%s\n",
			levelColor, prefix, entry.Message, resetColor))
	}

	return buf.Bytes(), nil
}

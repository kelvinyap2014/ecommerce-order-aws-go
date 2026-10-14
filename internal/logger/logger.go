package logger

import (
	"log"
)

// Info logs informational messages with a [INFO] tag.
func Info(format string, v ...any) {
	log.Printf("[INFO] "+format, v...)
}

// Warn logs warning messages with a [WARN] tag.
func Warn(format string, v ...any) {
	log.Printf("[WARN] "+format, v...)
}

// Error logs error messages with a [ERROR] tag.
func Error(format string, v ...any) {
	log.Printf("[ERROR] "+format, v...)
}

// utils/logger/logger.go

package utils

import (
	"log"
	"os"
)

// Logger is the global logger instance
var Logger *log.Logger

func init() {
	// Create a new logger with custom settings
	Logger = log.New(os.Stdout, "[chatAI] ", log.Ldate|log.Ltime|log.Lshortfile)
}

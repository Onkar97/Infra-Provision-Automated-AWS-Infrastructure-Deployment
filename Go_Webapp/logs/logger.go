package logs

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Global logger instance
var Log *zap.Logger

// InitLogger configures the logger based on the environment
func InitLogger() {
	env := os.Getenv("GO_ENV")
	logLevel := zap.InfoLevel

	// Set log level from env (simple implementation)
	if os.Getenv("LOG_LEVEL") == "debug" {
		logLevel = zap.DebugLevel
	}

	// 1. Encoder Configuration (JSON format)
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

	// JSON Encoder
	jsonEncoder := zapcore.NewJSONEncoder(encoderConfig)

	// 2. Define Cores (Transports)
	var cores []zapcore.Core

	// --- Console Transport (Always Add) ---
	// Equivalent to: new winston.transports.Console(...)
	consoleCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(os.Stdout), logLevel)
	cores = append(cores, consoleCore)

	// --- File Transport (Skip if Test) ---
	if env != "test" {
		var logFilePath string

		if env == "production" {
			// Equivalent to: process.env.LOG_FILE_PATH || '/opt/csye6225/webapp.log'
			logFilePath = os.Getenv("LOG_FILE_PATH")
			if logFilePath == "" {
				logFilePath = "/opt/csye6225/webapp.log"
			}
		} else {
			// Equivalent to: path.join(__dirname, '..', 'webapp.log');
			// Go typically runs from the project root, so "webapp.log" creates it there.
			cwd, _ := os.Getwd()
			logFilePath = filepath.Join(cwd, "webapp.log")
		}

		// Open file
		file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			// If we can't open the log file, we just log to console
			Error("Failed to open log file: " + err.Error())
		} else {
			fileCore := zapcore.NewCore(jsonEncoder, zapcore.AddSync(file), logLevel)
			cores = append(cores, fileCore)
		}
	}

	// 3. Create Logger with all cores
	core := zapcore.NewTee(cores...)

	// Equivalent to: defaultMeta: { service: 'webapp' }
	Log = zap.New(core, zap.AddCaller(), zap.Fields(zap.String("service", "webapp")))
}

// --- Helper Functions to match your existing 'logger.info' style ---

func Info(message string, fields ...zap.Field) {
	if Log != nil {
		Log.Info(message, fields...)
	}
}

func Warn(message string, fields ...zap.Field) {
	if Log != nil {
		Log.Warn(message, fields...)
	}
}

func Error(message string, fields ...zap.Field) {
	if Log != nil {
		Log.Error(message, fields...)
	}
}

func Fatal(message string, fields ...zap.Field) {
	if Log != nil {
		Log.Fatal(message, fields...)
	}
}

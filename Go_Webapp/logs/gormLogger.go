package logs

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// DBLogger adapts the GORM logger interface to our application's logging style.
type DBLogger struct {
	// You can add a specific logger instance here if you aren't using a global one
	// e.g., ZapLogger *zap.Logger
}

// LogMode sets the log level.
// GORM requires this method, but since we are writing custom logic in Trace(),
// we can usually just return the logger as-is.
func (l *DBLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	return l
}

// Info handles general DB info logs (equivalent to logSchemaBuild/logMigration)
func (l *DBLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	// Winston equivalent: logger.info(message, { context: 'DB' })
	log.Printf("[INFO] [DB] "+msg, data...)
}

// Warn handles DB warnings
func (l *DBLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	// Winston equivalent: logger.warn(message, { context: 'DB' })
	log.Printf("[WARN] [DB] "+msg, data...)
}

// Error handles DB errors
func (l *DBLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	// Winston equivalent: logger.error(message, { context: 'DB' })
	log.Printf("[ERROR] [DB] "+msg, data...)
}

// Trace is the core method. It handles both SQL query logging and Error logging.
// This is where we replicate your logic: Silence success, Log failure.
func (l *DBLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {

	// 1. Handle Errors (Equivalent to logQueryError)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		sql, _ := fc()

		// In your Node code: `Query failed: ${query} -- Parameters: ... -- Error: ...`
		// GORM's 'fc()' returns the SQL with parameters already filled in!
		message := fmt.Sprintf("Query failed: %s -- Error: %s", sql, err)

		log.Printf("[ERROR] [DB-Error] %s", message)
		return
	}

	// 2. Handle Success (Equivalent to logQuery)
	// Your Node code had an empty body for logQuery, so we do nothing here.
	// If you later want to log slow queries, you would check `time.Since(begin)` here.
}

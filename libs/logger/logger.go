package logger

import "fmt"

type LogLevel string

const (
	LogLevelEmergency LogLevel = "EMERGENCY"
	LogLevelAlert     LogLevel = "ALERT"
	LogLevelCritical  LogLevel = "CRITICAL"
	LogLevelError     LogLevel = "ERROR"
	LogLevelWarning   LogLevel = "WARNING"
	LogLevelNotice    LogLevel = "NOTICE"
	LogLevelInfo      LogLevel = "INFO"
	LogLevelDebug     LogLevel = "DEBUG"
)

func Log(level LogLevel, message string) {
	fmt.Printf("{ \"level\": \"%s\", \"message\": \"%s\"", level, message)
}

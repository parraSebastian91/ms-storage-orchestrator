package observability

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	ports "github.com/parraSebastian91/ms-storage-orchestrator.git/src/core/domain/ports/outbound"
)

var _ ports.ILoggerService = (*CustomLogger)(nil)

type CustomLogger struct {
	serviceName string
	minLevel    Level
	production  bool
	logger      *log.Logger
}

type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

const (
	reset  = "\033[0m"
	red    = "\033[31m"
	green  = "\033[32m"
	yellow = "\033[33m"
	blue   = "\033[34m"
	burple = "\033[35m"
)

func NewCustomLogger(ServiceName string, min string, production bool) *CustomLogger {
	var minLevel Level
	switch min {
	case "debug":
		minLevel = DebugLevel
	case "info":
		minLevel = InfoLevel
	case "warn":
		minLevel = WarnLevel
	case "error":
		minLevel = ErrorLevel
	case "fatal":
		minLevel = FatalLevel
	default:
		minLevel = InfoLevel
	}
	return &CustomLogger{
		serviceName: ServiceName,
		minLevel:    minLevel,
		production:  production,
		logger:      log.New(os.Stdout, "", 0),
	}
}

func (l *CustomLogger) log(level Level, msg string, color string, args ...interface{}) {
	if level < l.minLevel {
		return
	} // unwrap: si viene un solo map, tratarlo como campos
	var fields map[string]interface{}
	if len(args) == 1 {
		if m, ok := args[0].(map[string]interface{}); ok {
			fields = m
			args = nil
		}
	}

	if l.production {
		entry := map[string]interface{}{
			"ts":    time.Now().Format(time.RFC3339),
			"level": levelString(level),
			"msg":   msg,
		}
		if fields != nil {
			for k, v := range fields {
				entry[k] = v
			}
		} else if len(args) > 0 {
			entry["args"] = args
		}
		b, _ := json.Marshal(entry)
		l.logger.Println(string(b))
	} else {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		logmsg := fmt.Sprintf("%s[%s] [%s] %s: %s%s\n", color, timestamp, l.serviceName, levelString(level), msg, reset)
		if fields != nil {
			logmsg += fmt.Sprintf("%sFields: %v%s\n", burple, fields, reset)
		} else if len(args) > 0 {
			logmsg += fmt.Sprintf("%sArgs: %v%s\n", burple, args, reset)
		}
		fmt.Println(logmsg)
	}
}

func (l *CustomLogger) Debug(msg string, args ...interface{}) { l.log(DebugLevel, msg, blue, args...) }
func (l *CustomLogger) Info(msg string, args ...interface{})  { l.log(InfoLevel, msg, green, args...) }
func (l *CustomLogger) Warn(msg string, args ...interface{})  { l.log(WarnLevel, msg, yellow, args...) }
func (l *CustomLogger) Error(msg string, args ...interface{}) { l.log(ErrorLevel, msg, red, args...) }
func (l *CustomLogger) Fatal(msg string, args ...interface{}) {
	l.log(FatalLevel, msg, burple, args...)
}

func levelString(lvl Level) string {
	switch lvl {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case FatalLevel:
		return "fatal"
	default:
		return "info"
	}
}

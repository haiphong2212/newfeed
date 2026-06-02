package logger

import (
	"encoding/json"
	"log"
	"os"
	"time"
)

type Logger struct {
	logger *log.Logger
}

func New() *Logger {
	return &Logger{logger: log.New(os.Stdout, "", 0)}
}

func (l *Logger) Info(message string, fields map[string]any) {
	l.write("info", message, fields)
}

func (l *Logger) Error(message string, fields map[string]any) {
	l.write("error", message, fields)
}

func (l *Logger) write(level, message string, fields map[string]any) {
	if fields == nil {
		fields = map[string]any{}
	}
	fields["level"] = level
	fields["message"] = message
	fields["timestamp"] = time.Now().UTC().Format(time.RFC3339Nano)
	data, err := json.Marshal(fields)
	if err != nil {
		l.logger.Printf(`{"level":"error","message":"log marshal failed"}`)
		return
	}
	l.logger.Print(string(data))
}

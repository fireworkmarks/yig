package log

import (
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"

	log "github.com/sirupsen/logrus"
)

func ParseLevel(levelString string) log.Level {
	switch strings.ToLower(levelString) {
	case "info":
		return log.InfoLevel
	case "warn":
		return log.WarnLevel
	case "error":
		return log.ErrorLevel
	default:
		return log.InfoLevel
	}
}

type Logger struct {
	out       io.WriteCloser
	level     log.Level
	logger    *log.Logger
	requestID string
}

func NewFileLogger(path string, logLevel log.Level) Logger {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		panic(err)
	}
	return NewLogger(file, logLevel)
}

func NewLogger(out io.WriteCloser, logLevel log.Level) Logger {
	logger := log.New()
	formatter := log.TextFormatter{TimestampFormat: "2006-01-02 15:04:05"}
	logger.SetFormatter(&formatter)
	logger.Out = out
	l := Logger{
		out:    out,
		level:  logLevel,
		logger: logger,
	}
	return l
}

func (l Logger) NewWithRequestID(requestID string) Logger {
	return Logger{
		out:       l.out,
		level:     l.level,
		logger:    l.logger,
		requestID: requestID,
	}
}

func getCaller(skipCallDepth int) string {
	_, fullPath, line, ok := runtime.Caller(skipCallDepth)
	if !ok {
		return ""
	}
	fileParts := strings.Split(fullPath, "/")
	file := fileParts[len(fileParts)-1]
	return fmt.Sprintf("%s:%d", file, line)
}

func (l Logger) prefixArray() []interface{} {
	array := make([]interface{}, 0, 3)
	array = append(array, getCaller(3))
	if len(l.requestID) > 0 {
		array = append(array, l.requestID)
	}
	return array
}

func (l Logger) Info(args ...interface{}) {
	if l.level < log.InfoLevel {
		return
	}
	l.logger.SetLevel(log.InfoLevel)
	prefixArray := l.prefixArray()
	args = append(prefixArray, args...)
	l.logger.Println(args...)
}

func (l Logger) Warn(args ...interface{}) {
	if l.level < log.WarnLevel {
		return
	}
	l.logger.SetLevel(log.WarnLevel)
	prefixArray := l.prefixArray()
	args = append(prefixArray, args...)
	l.logger.Println(args...)
}

func (l Logger) Error(args ...interface{}) {
	if l.level < log.ErrorLevel {
		return
	}
	l.logger.SetLevel(log.ErrorLevel)
	prefixArray := l.prefixArray()
	args = append(prefixArray, args...)
	l.logger.Println(args...)
}

// Write a new line with args. Unless you really want to customize
// output format, use "Info", "Warn", "Error" instead
func (l Logger) Println(args ...interface{}) {
	_, _ = l.out.Write([]byte(fmt.Sprintln(args...)))
}

func (l Logger) Close() (err error) {
	err = l.out.Close()
	return
}

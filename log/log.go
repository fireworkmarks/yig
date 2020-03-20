package log

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/journeymidnight/yig/hashring"
	"github.com/minio/highwayhash"
)

type Level int

const (
	ErrorLevel Level = 0 // Errors should be properly handled
	WarnLevel  Level = 1 // Errors could be ignored; messages might need noticed
	InfoLevel  Level = 2 // Informational messages
)

const hashReplicationCount = 4096

const keyvalue = "000102030405060708090A0B0C0D0E0FF0E0D0C0B0A090807060504030201000" // This is the key for hash sum !

func ParseLevel(levelString string) Level {
	switch strings.ToLower(levelString) {
	case "info":
		return InfoLevel
	case "warn":
		return WarnLevel
	case "error":
		return ErrorLevel
	default:
		return InfoLevel
	}
}

type Logger struct {
	loggerHr  *hashring.HashRing
	out       []io.WriteCloser
	level     Level
	logger    []*log.Logger
	requestID string
	local     int
}

var (
	logFlags = log.Ldate | log.Ltime | log.Lmicroseconds
)

func NewFileLogger(paths []string, logLevel Level) Logger {
	key, err := hex.DecodeString(keyvalue)
	if err != nil {
		panic(err)
	}
	hash, err := highwayhash.New64(key)
	if err != nil {
		panic(err)
	}
	loggerHr := hashring.NewHashRing(hashReplicationCount, hash)
	var files []io.WriteCloser
	for i, path := range paths {
		f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			panic("Failed to open log file " + path)
		}
		files = append(files, f)
		err = loggerHr.Add(i)
		if err != nil {
			panic(err)
		}
	}
	return NewLogger(loggerHr, files, logLevel)
}

func NewLogger(loggerHr *hashring.HashRing, outs []io.WriteCloser, logLevel Level) Logger {
	var loggerinfo []*log.Logger
	for _, out := range outs {
		loggerinfo = append(loggerinfo, log.New(out, "", logFlags))
	}
	l := Logger{
		loggerHr: loggerHr,
		out:      outs,
		level:    logLevel,
		logger:   loggerinfo,
	}
	return l
}

func (l Logger) NewWithRequestID(requestID string) Logger {
	local, err := l.GetLocate(requestID)
	if err != nil {
		panic(err)
	}
	return Logger{
		out:       l.out,
		level:     l.level,
		logger:    l.logger,
		requestID: requestID,
		local:     local,
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
	if l.level < InfoLevel {
		return
	}
	prefixArray := l.prefixArray()
	prefixArray = append(prefixArray, "[INFO]")
	args = append(prefixArray, args...)
	l.logger[l.local].Println(args...)
}

func (l Logger) Warn(args ...interface{}) {
	if l.level < WarnLevel {
		return
	}
	prefixArray := l.prefixArray()
	prefixArray = append(prefixArray, "[WARN]")
	args = append(prefixArray, args...)
	l.logger[l.local].Println(args...)
}

func (l Logger) Error(args ...interface{}) {
	if l.level < ErrorLevel {
		return
	}
	prefixArray := l.prefixArray()
	prefixArray = append(prefixArray, "[ERROR]")
	args = append(prefixArray, args...)
	l.logger[l.local].Println(args...)
}

// Write a new line with args. Unless you really want to customize
// output format, use "Info", "Warn", "Error" instead
func (l Logger) Println(args ...interface{}) {
	_, _ = l.out[0].Write([]byte(fmt.Sprintln(args...)))
}

func (l Logger) Close() (err error) {
	for _, out := range l.out {
		err = out.Close()
	}
	return
}

func (l Logger) GetLocate(key string) (int, error) {
	n, err := l.loggerHr.Locate(key)
	if err != nil {
		return 0, err
	}
	return n.(int), nil
}

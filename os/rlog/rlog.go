package rlog

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

const (
	LevelDebug = iota
	LevelInfo
	LevelWarning
	LevelError
	LevelFatal
)

var (
	instances = make(map[string]*Logger)
	mu        sync.RWMutex
)

type Logger struct {
	name   string
	level  int
	logger *log.Logger
	mu     sync.RWMutex
}

func Instance(name ...string) *Logger {
	n := "default"
	if len(name) > 0 && name[0] != "" {
		n = name[0]
	}
	mu.RLock()
	if l, ok := instances[n]; ok {
		mu.RUnlock()
		return l
	}
	mu.RUnlock()

	mu.Lock()
	defer mu.Unlock()
	if l, ok := instances[n]; ok {
		return l
	}
	l := &Logger{
		name:   n,
		level:  LevelInfo,
		logger: log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile),
	}
	instances[n] = l
	return l
}

func (l *Logger) SetLevel(level int) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logger.SetOutput(w)
}

func (l *Logger) Debug(v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.level <= LevelDebug {
		l.logger.Output(2, "[DEBUG] "+fmt.Sprint(v...))
	}
}

func (l *Logger) Debugf(format string, v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.level <= LevelDebug {
		l.logger.Output(2, "[DEBUG] "+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Info(v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.level <= LevelInfo {
		l.logger.Output(2, "[INFO] "+fmt.Sprint(v...))
	}
}

func (l *Logger) Infof(format string, v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.level <= LevelInfo {
		l.logger.Output(2, "[INFO] "+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Warning(v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.level <= LevelWarning {
		l.logger.Output(2, "[WARN] "+fmt.Sprint(v...))
	}
}

func (l *Logger) Warningf(format string, v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.level <= LevelWarning {
		l.logger.Output(2, "[WARN] "+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Error(v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.level <= LevelError {
		l.logger.Output(2, "[ERROR] "+fmt.Sprint(v...))
	}
}

func (l *Logger) Errorf(format string, v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	if l.level <= LevelError {
		l.logger.Output(2, "[ERROR] "+fmt.Sprintf(format, v...))
	}
}

func (l *Logger) Fatal(v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Output(2, "[FATAL] "+fmt.Sprint(v...))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...any) {
	l.mu.RLock()
	defer l.mu.RUnlock()
	l.logger.Output(2, "[FATAL] "+fmt.Sprintf(format, v...))
	os.Exit(1)
}

func (l *Logger) Print(v ...any) {
	l.logger.Output(2, fmt.Sprint(v...))
}

func (l *Logger) Printf(format string, v ...any) {
	l.logger.Output(2, fmt.Sprintf(format, v...))
}

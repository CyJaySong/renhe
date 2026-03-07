package rlog

import (
	"encoding/json"
	"io"
	"os"

	"github.com/labstack/gommon/log"
)

type echoLogger struct {
	l      *Logger
	output io.Writer
	prefix string
	level  log.Lvl
}

// EchoLogger 返回实现了 echo.Logger 接口的适配器。
func (l *Logger) EchoLogger() *echoLogger {
	return &echoLogger{
		l:      l,
		output: os.Stdout,
		level:  log.INFO,
	}
}

func (e *echoLogger) Output() io.Writer         { return e.output }
func (e *echoLogger) SetOutput(w io.Writer)      { e.output = w }
func (e *echoLogger) Prefix() string             { return e.prefix }
func (e *echoLogger) SetPrefix(p string)         { e.prefix = p }
func (e *echoLogger) Level() log.Lvl             { return e.level }
func (e *echoLogger) SetLevel(v log.Lvl)         { e.level = v }
func (e *echoLogger) SetHeader(_ string)         {}

func (e *echoLogger) Print(i ...interface{})                { e.l.logger.Info(i...) }
func (e *echoLogger) Printf(format string, args ...interface{}) { e.l.logger.Infof(format, args...) }
func (e *echoLogger) Printj(j log.JSON)                     { e.l.logger.Info(jsonStr(j)) }

func (e *echoLogger) Debug(i ...interface{})                { e.l.logger.Debug(i...) }
func (e *echoLogger) Debugf(format string, args ...interface{}) { e.l.logger.Debugf(format, args...) }
func (e *echoLogger) Debugj(j log.JSON)                     { e.l.logger.Debug(jsonStr(j)) }

func (e *echoLogger) Info(i ...interface{})                { e.l.logger.Info(i...) }
func (e *echoLogger) Infof(format string, args ...interface{}) { e.l.logger.Infof(format, args...) }
func (e *echoLogger) Infoj(j log.JSON)                     { e.l.logger.Info(jsonStr(j)) }

func (e *echoLogger) Warn(i ...interface{})                { e.l.logger.Warn(i...) }
func (e *echoLogger) Warnf(format string, args ...interface{}) { e.l.logger.Warnf(format, args...) }
func (e *echoLogger) Warnj(j log.JSON)                     { e.l.logger.Warn(jsonStr(j)) }

func (e *echoLogger) Error(i ...interface{})                { e.l.logger.Error(i...) }
func (e *echoLogger) Errorf(format string, args ...interface{}) { e.l.logger.Errorf(format, args...) }
func (e *echoLogger) Errorj(j log.JSON)                     { e.l.logger.Error(jsonStr(j)) }

func (e *echoLogger) Fatal(i ...interface{})                { e.l.logger.Fatal(i...) }
func (e *echoLogger) Fatalf(format string, args ...interface{}) { e.l.logger.Fatalf(format, args...) }
func (e *echoLogger) Fatalj(j log.JSON)                     { e.l.logger.Fatal(jsonStr(j)) }

func (e *echoLogger) Panic(i ...interface{})                { e.l.logger.Panic(i...) }
func (e *echoLogger) Panicf(format string, args ...interface{}) { e.l.logger.Panicf(format, args...) }
func (e *echoLogger) Panicj(j log.JSON)                     { e.l.logger.Panic(jsonStr(j)) }

func jsonStr(j log.JSON) string {
	b, _ := json.Marshal(j)
	return string(b)
}

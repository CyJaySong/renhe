// Package rlog 提供全局日志服务，基于 gookit/slog 实现。
// 支持上下文字段注入、OpenTelemetry Trace 自动关联，并提供 echo.Logger 和 bun.QueryHook 适配器。
//
// 配置路径 logger（YAML 示例）:
//
//	logger:
//	  level: "debug"            # trace/debug/info/notice/warn/error/fatal（默认 info）
//	  format: "text"            # text/json（默认 text）
//	  output: "both"            # console/file/both（默认 console）
//	  callerSkip: 7             # 调用栈跳过层数（默认 7）
//	  stack: false              # Error/Warn/Fatal 级别是否追加完整调用栈（默认 false）
//	  file:
//	    path: "logs/app.log"    # 日志文件路径（output 包含 file 时必填）
//	    maxSize: 128             # 单文件最大体积 MB（默认 128）
//	    rotateTime: "every_day" # 按时间轮转: every_hour/every_day/every_30min 等（默认 every_day）
//	    backupNum: 20           # 保留旧文件数量（默认 20）
//	    backupTime: 168         # 保留旧文件最大小时数（默认 168 即 7 天）
//	    compress: false         # 是否 gzip 压缩旧文件（默认 false）
//	    buffSize: 8192          # 写缓冲字节数，0 禁用（默认 8192）
package rlog

import (
	"strings"
	"sync"

	"github.com/cyjaysong/renhe/os/rcfg"
	"github.com/gookit/slog"
	"github.com/gookit/slog/handler"
	"github.com/gookit/slog/rotatefile"
)

var (
	instance *Logger
	once     sync.Once
)

// Logger 封装 gookit/slog.Logger，提供带 context 的日志方法。
type Logger struct {
	logger *slog.Logger
	stack  bool // 是否在 Error/Warn/Fatal 级别追加堆栈
}

// logConfig 内部配置结构体，对应 YAML logger 节点。
type logConfig struct {
	Level      string     `yaml:"level"`
	Format     string     `yaml:"format"`
	Output     string     `yaml:"output"`
	CallerSkip int        `yaml:"callerSkip"`
	Stack      bool       `yaml:"stack"` // Error/Warn/Fatal 级别是否追加调用栈
	File       fileConfig `yaml:"file"`
}

type fileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"maxSize"`
	RotateTime string `yaml:"rotateTime"`
	BackupNum  int    `yaml:"backupNum"`
	BackupTime int    `yaml:"backupTime"`
	Compress   bool   `yaml:"compress"`
	BuffSize   int    `yaml:"buffSize"`
}

// Log 返回全局 Logger 单例。首次调用时根据配置初始化日志级别、格式和 Handler。
func Log() *Logger {
	once.Do(func() {
		cfg := loadConfig()
		level := parseLevel(cfg.Level)
		useJSON := strings.EqualFold(cfg.Format, "json")

		inner := slog.New()
		inner.ReportCaller = true
		inner.CallerSkip = cfg.CallerSkip
		inner.CallerFlag = slog.CallerFlagFpLine

		output := strings.ToLower(cfg.Output)

		// Console handler
		if output == "console" || output == "both" {
			ch := handler.NewConsoleHandler(levelsUpTo(level))
			if useJSON {
				ch.SetFormatter(slog.NewJSONFormatter())
			} else {
				applyTextFormatter(ch)
			}
			inner.AddHandler(ch)
		}

		// File handler
		if output == "file" || output == "both" {
			fh := buildFileHandler(cfg.File, level, useJSON)
			if fh != nil {
				inner.AddHandler(fh)
			}
		}

		instance = &Logger{logger: inner, stack: cfg.Stack}
	})
	return instance
}

// Close 刷新并关闭所有 handler（含文件缓冲）。应在程序退出前调用。
func (l *Logger) Close() {
	l.logger.Close()
}

// loadConfig 从 rcfg 读取 logger 配置并填充默认值。
func loadConfig() logConfig {
	cfg := logConfig{
		Level:      "info",
		Format:     "text",
		Output:     "console",
		CallerSkip: 7,
		File: fileConfig{
			Path:       "logs/app.log",
			MaxSize:    128,
			RotateTime: "every_day",
			BackupNum:  20,
			BackupTime: 168,
			Compress:   false,
			BuffSize:   8192,
		},
	}
	v := rcfg.Cfg()
	if v == nil {
		return cfg
	}
	sub := v.Sub("logger")
	if sub == nil {
		return cfg
	}
	_ = sub.Unmarshal(&cfg, rcfg.YamlTagOption)
	// 确保默认值
	if cfg.CallerSkip <= 0 {
		cfg.CallerSkip = 6
	}
	if cfg.File.MaxSize <= 0 {
		cfg.File.MaxSize = 128
	}
	if cfg.File.BackupNum <= 0 {
		cfg.File.BackupNum = 20
	}
	if cfg.File.BackupTime <= 0 {
		cfg.File.BackupTime = 168
	}
	return cfg
}

// parseLevel 解析日志级别字符串，默认 InfoLevel。
func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "trace":
		return slog.TraceLevel
	case "debug":
		return slog.DebugLevel
	case "info":
		return slog.InfoLevel
	case "notice":
		return slog.NoticeLevel
	case "warn", "warning":
		return slog.WarnLevel
	case "error":
		return slog.ErrorLevel
	case "fatal":
		return slog.FatalLevel
	default:
		return slog.InfoLevel
	}
}

// levelsUpTo 返回从最高级别到指定级别的所有级别列表。
func levelsUpTo(maxLevel slog.Level) []slog.Level {
	all := []slog.Level{
		slog.PanicLevel, slog.FatalLevel, slog.ErrorLevel,
		slog.WarnLevel, slog.NoticeLevel, slog.InfoLevel,
		slog.DebugLevel, slog.TraceLevel,
	}
	var levels []slog.Level
	for _, l := range all {
		if l <= maxLevel {
			levels = append(levels, l)
		}
	}
	return levels
}

// buildFileHandler 根据文件配置创建带轮转的文件 handler。
func buildFileHandler(fc fileConfig, level slog.Level, useJSON bool) slog.Handler {
	if fc.Path == "" {
		return nil
	}
	cfg := handler.NewEmptyConfig(
		handler.WithLogfile(fc.Path),
		handler.WithLogLevels(levelsUpTo(level)),
		handler.WithBuffMode(handler.BuffModeLine),
		handler.WithBuffSize(fc.BuffSize),
		handler.WithCompress(fc.Compress),
		handler.WithBackupNum(uint(fc.BackupNum)),
		handler.WithBackupTime(uint(fc.BackupTime)),
		handler.WithMaxSize(uint64(fc.MaxSize)*1024*1024),
		handler.WithRotateTime(parseRotateTime(fc.RotateTime)),
	)
	if useJSON {
		cfg.UseJSON = true
	}
	h, err := cfg.CreateHandler()
	if err != nil {
		panic("rlog: failed to create file handler: " + err.Error())
	}
	return h
}

// parseRotateTime 解析轮转时间配置字符串。
func parseRotateTime(s string) rotatefile.RotateTime {
	switch strings.ToLower(s) {
	case "every_hour", "everyhour", "1hour":
		return rotatefile.EveryHour
	case "every_day", "everyday", "1day":
		return rotatefile.EveryDay
	case "every_30min", "every30min", "30min":
		return rotatefile.Every30Min
	case "every_15min", "every15min", "15min":
		return rotatefile.Every15Min
	default:
		return rotatefile.EveryDay
	}
}

// customTemplate 自定义日志模板：caller 独占一行，便于 IDE 点击跳转。
// 输出效果：
//
//	[2025/01/15 14:30:25] [application] [INFO] message content
//	[/Users/me/project/main.go:42]
const customTemplate = "[{{datetime}}] [{{channel}}] [{{level}}] {{message}} {{data}} {{extra}}\n[{{caller}}]\n"

// applyTextFormatter 为 handler 设置自定义文本格式化器。
func applyTextFormatter(h slog.FormattableHandler) {
	f := slog.NewTextFormatter(customTemplate)
	f.EnableColor = true
	h.SetFormatter(f)
}

// Inner 返回底层 *slog.Logger，便于高级用法。
func (l *Logger) Inner() *slog.Logger {
	return l.logger
}

// AddHandler 为当前 Logger 追加 Handler。
func (l *Logger) AddHandler(h slog.Handler) {
	l.logger.AddHandler(h)
}

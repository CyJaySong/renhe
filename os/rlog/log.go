// Package rlog 提供全局日志服务，基于标准库 log/slog 实现。
// 支持上下文字段注入、OpenTelemetry Trace 自动关联，并提供 bun.QueryHook 适配器。
//
// 配置路径 logger（YAML 示例）:
//
//	logger:
//	  level: "debug"            # debug/info/warn/error（默认 info；旧值 trace/notice/fatal 会映射）
//	  format: "text"            # text/json（默认 text）
//	  output: "both"            # console/file/both（默认 console）
//	  stack: false              # Error/Fatal/Panic 级别是否追加完整调用栈（默认 false）
//	  file:
//	    path: "logs/app.log"    # 日志文件路径（output 包含 file 时必填）
//	    maxSize: 128             # 单文件最大体积 MB（默认 128）
//	    rotateTime: "every_day" # 兼容字段；当前按体积轮转（lumberjack）
//	    backupNum: 20           # 保留旧文件数量（默认 20）
//	    backupTime: 168         # 保留旧文件最大小时数（默认 168 即 7 天）
//	    compress: false         # 是否 gzip 压缩旧文件（默认 false）
//	    buffSize: 8192          # 兼容字段；写入由 lumberjack 处理
package rlog

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/cyjaysong/renhe/os/rcfg"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	instance *Logger
	once     sync.Once
)

// Logger 封装 *slog.Logger，提供带 context 的日志方法。
type Logger struct {
	mu       sync.Mutex
	logger   *slog.Logger
	handlers []slog.Handler
	stack    bool // 是否在 Error/Fatal/Panic 级别追加堆栈
	closers  []io.Closer
}

// logConfig 内部配置结构体，对应 YAML logger 节点。
type logConfig struct {
	Level  string     `yaml:"level"`
	Format string     `yaml:"format"`
	Output string     `yaml:"output"`
	Stack  bool       `yaml:"stack"` // Error/Fatal/Panic 级别是否追加调用栈
	File   fileConfig `yaml:"file"`
}

type fileConfig struct {
	Path       string `yaml:"path"`
	MaxSize    int    `yaml:"maxSize"`
	RotateTime string `yaml:"rotateTime"` // 兼容保留；lumberjack 以体积轮转为主
	BackupNum  int    `yaml:"backupNum"`
	BackupTime int    `yaml:"backupTime"`
	Compress   bool   `yaml:"compress"`
	BuffSize   int    `yaml:"buffSize"` // 兼容保留
}

// Log 返回全局 Logger 单例。首次调用时根据配置初始化日志级别、格式和 Handler。
func Log() *Logger {
	once.Do(func() {
		cfg := loadConfig()
		minLevel := parseLevel(cfg.Level)
		useJSON := strings.EqualFold(cfg.Format, "json")
		output := strings.ToLower(cfg.Output)

		var handlers []slog.Handler
		var closers []io.Closer

		// Console handler
		if output == "console" || output == "both" || output == "" {
			handlers = append(handlers, newHandler(os.Stdout, minLevel, useJSON))
		}

		// File handler（lumberjack 体积轮转）
		if output == "file" || output == "both" {
			fh, closer := buildFileHandler(cfg.File, minLevel, useJSON)
			if fh != nil {
				handlers = append(handlers, fh)
			}
			if closer != nil {
				closers = append(closers, closer)
			}
		}

		if len(handlers) == 0 {
			handlers = append(handlers, newHandler(os.Stdout, minLevel, useJSON))
		}

		root := &multiHandler{handlers: handlers}
		instance = &Logger{
			logger:   slog.New(root),
			handlers: handlers,
			stack:    cfg.Stack,
			closers:  closers,
		}
	})
	return instance
}

// Close 刷新并关闭文件等可关闭资源。应在程序退出前调用。
func (l *Logger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, c := range l.closers {
		_ = c.Close()
	}
	l.closers = nil
}

// loadConfig 从 rcfg 读取 logger 配置并填充默认值。
func loadConfig() logConfig {
	cfg := logConfig{
		Level:  "info",
		Format: "text",
		Output: "console",
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

// parseLevel 解析日志级别，仅保留 slog 四级：debug/info/warn/error。
// 旧配置值兼容映射：trace→debug、notice→info、fatal→error。
func parseLevel(s string) slog.Level {
	switch strings.ToLower(s) {
	case "trace", "debug":
		return slog.LevelDebug
	case "info", "notice":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error", "fatal":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// newHandler 创建 console/file 共用的 slog Handler。
func newHandler(w io.Writer, minLevel slog.Level, useJSON bool) slog.Handler {
	opts := &slog.HandlerOptions{
		Level: minLevel,
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			// 统一 level 展示名
			if a.Key == slog.LevelKey {
				if lv, ok := a.Value.Any().(slog.Level); ok {
					a.Value = slog.StringValue(levelName(lv))
				}
			}
			return a
		},
	}
	if useJSON {
		return slog.NewJSONHandler(w, opts)
	}
	return slog.NewTextHandler(w, opts)
}

// levelName 将 slog.Level 转为可读名称（仅四级）。
func levelName(l slog.Level) string {
	switch {
	case l <= slog.LevelDebug:
		return "DEBUG"
	case l <= slog.LevelInfo:
		return "INFO"
	case l <= slog.LevelWarn:
		return "WARN"
	default:
		return "ERROR"
	}
}

// buildFileHandler 根据文件配置创建带体积轮转的文件 handler。
func buildFileHandler(fc fileConfig, minLevel slog.Level, useJSON bool) (slog.Handler, io.Closer) {
	if fc.Path == "" {
		return nil, nil
	}
	// 确保目录存在
	if dir := filepath.Dir(fc.Path); dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	// backupTime 配置为小时，lumberjack.MaxAge 为天
	maxAgeDays := fc.BackupTime / 24
	if fc.BackupTime > 0 && maxAgeDays < 1 {
		maxAgeDays = 1
	}
	lj := &lumberjack.Logger{
		Filename:   fc.Path,
		MaxSize:    fc.MaxSize, // MB
		MaxBackups: fc.BackupNum,
		MaxAge:     maxAgeDays,
		Compress:   fc.Compress,
		LocalTime:  true,
	}
	return newHandler(lj, minLevel, useJSON), lj
}

// multiHandler 将日志扇出到多个 Handler。
type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	var firstErr error
	for _, h := range m.handlers {
		if !h.Enabled(ctx, r.Level) {
			continue
		}
		// Record 按值传递，各 handler 独立消费
		if err := h.Handle(ctx, r.Clone()); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	hs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hs[i] = h.WithAttrs(attrs)
	}
	return &multiHandler{handlers: hs}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	hs := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		hs[i] = h.WithGroup(name)
	}
	return &multiHandler{handlers: hs}
}

// Slog 返回底层 *slog.Logger，供 Echo 等组件注入。
func (l *Logger) Slog() *slog.Logger {
	return l.logger
}

// Inner 返回底层 *slog.Logger（与 Slog 相同，保留旧方法名）。
func (l *Logger) Inner() *slog.Logger {
	return l.logger
}

// AddHandler 为当前 Logger 追加 Handler。
func (l *Logger) AddHandler(h slog.Handler) {
	if h == nil {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	l.handlers = append(l.handlers, h)
	// 复制切片，避免后续 append 影响已发布的 multiHandler
	copied := append([]slog.Handler(nil), l.handlers...)
	l.logger = slog.New(&multiHandler{handlers: copied})
}

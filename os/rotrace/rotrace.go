// Package rotrace 提供 OpenTelemetry TracerProvider 的初始化与管理。
//
// 配置路径 trace（YAML 示例）:
//
//	trace:
//	  enable: true
//	  exporter: stdout           # stdout | otlp | otlphttp | none
//	  sampler: 1.0
//	  serviceName: "my-service"
//	  otlp:
//	    endpoint: "localhost:4318"
//	    insecure: true
//
// 下游零样板:
//
//	defer r.Close()
//	httpSrv := r.HttpSrv() // enable=true 时按 exporter 自动 Init
//	httpSrv.Run()
//
// gRPC OTLP 或其它自定义 exporter：在 HttpSrv 之前调用 InitWithExporter。
package rotrace

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/cyjaysong/renhe/os/rcfg"
	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/cyjaysong/renhe/os/rlog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	once        sync.Once
	mu          sync.Mutex
	globalTP    *sdktrace.TracerProvider
	serviceName string
)

type otlpConfig struct {
	Endpoint string `yaml:"endpoint"`
	Insecure bool   `yaml:"insecure"`
}

type traceConfig struct {
	Enable      *bool      `yaml:"enable"`
	Exporter    string     `yaml:"exporter"` // stdout | otlp | otlphttp | none
	Sampler     float64    `yaml:"sampler"`
	ServiceName string     `yaml:"serviceName"`
	OTLP        otlpConfig `yaml:"otlp"`
}

func (c traceConfig) enabled() bool {
	if c.Enable == nil {
		return false
	}
	return *c.Enable
}

// Enabled 返回 YAML trace.enable 是否为 true。
func Enabled() bool {
	return loadConfig().enabled()
}

// Ensure 按配置自动初始化。enable=false 或 exporter=none 时 no-op。
func Ensure(ctx context.Context) {
	cfg := loadConfig()
	if !cfg.enabled() {
		return
	}
	exp, err := newExporterFromConfig(ctx, cfg)
	if err != nil {
		rlog.Log().Error(ctx, "rotrace: create exporter failed", "err", err)
		return
	}
	if exp == nil {
		return
	}
	_ = InitWithExporter(ctx, exp)
}

// Init 使用 stdout exporter 初始化（开发调试）。
func Init(ctx context.Context) (shutdown func(context.Context) error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		rlog.Log().Error(ctx, "rotrace: failed to create stdout exporter", "err", err)
		return Shutdown
	}
	return InitWithExporter(ctx, exp)
}

// InitWithExporter 使用自定义 SpanExporter 初始化；仅首次生效。
func InitWithExporter(ctx context.Context, exp sdktrace.SpanExporter) (shutdown func(context.Context) error) {
	once.Do(func() {
		cfg := loadConfig()
		serviceName = cfg.ServiceName
		tp := newTracerProvider(ctx, cfg, exp)
		mu.Lock()
		globalTP = tp
		mu.Unlock()
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		rctx.ReExtractInitCtx()
	})
	return Shutdown
}

// Shutdown 关闭全局 TracerProvider。
func Shutdown(ctx context.Context) error {
	mu.Lock()
	tp := globalTP
	globalTP = nil
	mu.Unlock()
	if tp == nil {
		return nil
	}
	return tp.Shutdown(ctx)
}

// ServiceName 返回服务名。
func ServiceName() string {
	if serviceName != "" {
		return serviceName
	}
	return loadConfig().ServiceName
}

func loadConfig() traceConfig {
	cfg := traceConfig{
		Sampler:     1.0,
		ServiceName: "unknown-service",
		Exporter:    "stdout",
		OTLP: otlpConfig{
			Endpoint: "localhost:4318",
			Insecure: true,
		},
	}

	v := rcfg.Cfg()
	if v == nil {
		return cfg
	}
	sub := v.Sub("trace")
	if sub == nil {
		return cfg
	}
	_ = sub.Unmarshal(&cfg, rcfg.YamlTagOption)
	if cfg.Sampler <= 0 {
		cfg.Sampler = 0
	} else if cfg.Sampler > 1 {
		cfg.Sampler = 1
	}
	if cfg.ServiceName == "" {
		cfg.ServiceName = "unknown-service"
	}
	if strings.TrimSpace(cfg.Exporter) == "" {
		cfg.Exporter = "stdout"
	}
	if strings.TrimSpace(cfg.OTLP.Endpoint) == "" {
		cfg.OTLP.Endpoint = "localhost:4318"
	}
	return cfg
}

// resolveExporterKind 归一化：stdout | none | otlphttp
func resolveExporterKind(cfg traceConfig) string {
	e := strings.ToLower(strings.TrimSpace(cfg.Exporter))
	switch e {
	case "", "stdout":
		return "stdout"
	case "none", "off", "noop":
		return "none"
	case "otlp", "otlphttp", "http":
		return "otlphttp"
	case "otlpgrpc", "grpc":
		// 自动配置暂不内置 gRPC，避免额外依赖冲突；请 InitWithExporter
		return "otlpgrpc_unsupported"
	default:
		return "unknown:" + e
	}
}

func newExporterFromConfig(ctx context.Context, cfg traceConfig) (sdktrace.SpanExporter, error) {
	kind := resolveExporterKind(cfg)
	switch kind {
	case "none":
		return nil, nil
	case "stdout":
		return stdouttrace.New(stdouttrace.WithPrettyPrint())
	case "otlphttp":
		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(cfg.OTLP.Endpoint),
		}
		if cfg.OTLP.Insecure {
			opts = append(opts, otlptracehttp.WithInsecure())
		}
		return otlptracehttp.New(ctx, opts...)
	case "otlpgrpc_unsupported":
		return nil, fmt.Errorf("exporter otlpgrpc not auto-configured; use rotrace.InitWithExporter with otlptracegrpc, or set exporter: otlphttp")
	default:
		return nil, fmt.Errorf("unknown exporter %q (want stdout|otlp|otlphttp|none)", cfg.Exporter)
	}
}

func newTracerProvider(_ context.Context, cfg traceConfig, exp sdktrace.SpanExporter) *sdktrace.TracerProvider {
	res, _ := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(cfg.ServiceName),
		),
	)

	var sampler sdktrace.Sampler
	if cfg.Sampler >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if cfg.Sampler <= 0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(cfg.Sampler)
	}

	opts := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
	}
	if exp != nil {
		opts = append(opts, sdktrace.WithBatcher(exp))
	}
	return sdktrace.NewTracerProvider(opts...)
}

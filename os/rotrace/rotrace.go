// Package rotrace 提供 OpenTelemetry TracerProvider 的初始化与管理。
//
// 配置路径 trace（YAML 示例）:
//
//	trace:
//	  sampler: 1.0               # 采样率 0.0~1.0（默认 1.0 全采样）
//	  serviceName: "my-service"  # 服务名（默认 "unknown-service"）
//
// 基础用法（stdout exporter，开发调试用）:
//
//	shutdown := rotrace.Init(ctx)
//	defer shutdown(ctx)
//
// 生产用法（自定义 exporter，如 OTLP HTTP）:
//
//	exp, _ := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("localhost:4318"), otlptracehttp.WithInsecure())
//	shutdown := rotrace.InitWithExporter(ctx, exp)
//	defer shutdown(ctx)
package rotrace

import (
	"context"
	"sync"

	"github.com/cyjaysong/renhe/os/rcfg"
	"github.com/cyjaysong/renhe/os/rctx"
	"github.com/cyjaysong/renhe/os/rlog"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
)

var (
	once        sync.Once
	serviceName string
)

type traceConfig struct {
	Sampler     float64 `yaml:"sampler"`
	ServiceName string  `yaml:"serviceName"`
}

// Init 初始化全局 TracerProvider，使用 stdout exporter（适合开发调试）。
// 返回 shutdown 函数，应在程序退出前调用。
// 仅首次调用生效，重复调用返回空操作 shutdown。
func Init(ctx context.Context) (shutdown func(context.Context) error) {
	exp, err := stdouttrace.New(stdouttrace.WithPrettyPrint())
	if err != nil {
		rlog.Log().Error(ctx, "rotrace: failed to create stdout exporter", "err", err)
		return func(context.Context) error { return nil }
	}
	return InitWithExporter(ctx, exp)
}

// InitWithExporter 初始化全局 TracerProvider，使用调用方提供的 SpanExporter。
// 生产环境推荐使用此方法，传入 otlptracehttp 或 otlptracegrpc exporter。
//
// 示例:
//
//	exp, _ := otlptracehttp.New(ctx, otlptracehttp.WithEndpoint("localhost:4318"))
//	shutdown := rotrace.InitWithExporter(ctx, exp)
//	defer shutdown(ctx)
func InitWithExporter(ctx context.Context, exp sdktrace.SpanExporter) (shutdown func(context.Context) error) {
	var tp *sdktrace.TracerProvider
	once.Do(func() {
		cfg := loadConfig()
		serviceName = cfg.ServiceName
		tp = newTracerProvider(ctx, cfg, exp)
		otel.SetTracerProvider(tp)
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		))
		rctx.ReExtractInitCtx()
	})
	if tp == nil {
		return func(context.Context) error { return nil }
	}
	return tp.Shutdown
}

// ServiceName 返回配置中的服务名。
// 若 Init 尚未调用，返回默认值 "unknown-service"。
func ServiceName() string {
	if serviceName == "" {
		return "unknown-service"
	}
	return serviceName
}

func loadConfig() traceConfig {
	cfg := traceConfig{
		Sampler:     1.0,
		ServiceName: "unknown-service",
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
	return cfg
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

package rotrace

import "testing"

func TestTraceConfigEnabled(t *testing.T) {
	if (traceConfig{}).enabled() {
		t.Fatal("default false")
	}
	f, tr := false, true
	if (traceConfig{Enable: &f}).enabled() {
		t.Fatal("false")
	}
	if !(traceConfig{Enable: &tr}).enabled() {
		t.Fatal("true")
	}
}

func TestResolveExporterKind(t *testing.T) {
	cases := []struct {
		exp, want string
	}{
		{"", "stdout"},
		{"stdout", "stdout"},
		{"none", "none"},
		{"off", "none"},
		{"otlp", "otlphttp"},
		{"otlphttp", "otlphttp"},
		{"http", "otlphttp"},
		{"otlpgrpc", "otlpgrpc_unsupported"},
		{"grpc", "otlpgrpc_unsupported"},
		{"weird", "unknown:weird"},
	}
	for _, tc := range cases {
		cfg := traceConfig{Exporter: tc.exp}
		if got := resolveExporterKind(cfg); got != tc.want {
			t.Fatalf("exp=%q got=%q want=%q", tc.exp, got, tc.want)
		}
	}
}

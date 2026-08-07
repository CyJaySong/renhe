package rlog

import (
	"log/slog"
	"testing"
)

func TestParseLevel(t *testing.T) {
	cases := []struct {
		in   string
		want slog.Level
	}{
		{"debug", slog.LevelDebug},
		{"DEBUG", slog.LevelDebug},
		{"trace", slog.LevelDebug}, // 兼容映射
		{"info", slog.LevelInfo},
		{"notice", slog.LevelInfo}, // 兼容映射
		{"warn", slog.LevelWarn},
		{"warning", slog.LevelWarn},
		{"error", slog.LevelError},
		{"fatal", slog.LevelError}, // 兼容映射
		{"", slog.LevelInfo},
		{"unknown", slog.LevelInfo},
	}
	for _, tc := range cases {
		got := parseLevel(tc.in)
		if got != tc.want {
			t.Fatalf("parseLevel(%q)=%v want %v", tc.in, got, tc.want)
		}
	}
}

func TestLevelName(t *testing.T) {
	if levelName(slog.LevelDebug) != "DEBUG" {
		t.Fatal(levelName(slog.LevelDebug))
	}
	if levelName(slog.LevelInfo) != "INFO" {
		t.Fatal(levelName(slog.LevelInfo))
	}
	if levelName(slog.LevelWarn) != "WARN" {
		t.Fatal(levelName(slog.LevelWarn))
	}
	if levelName(slog.LevelError) != "ERROR" {
		t.Fatal(levelName(slog.LevelError))
	}
}

package rlog

import (
	"context"
	"testing"
)

func TestWithFields_CopiesMap(t *testing.T) {
	src := map[string]any{"a": 1}
	ctx := WithFields(context.Background(), src)
	src["a"] = 2
	src["b"] = 3

	// 从 ctx 取出的字段不应被外部 map 修改影响
	fields, _ := ctx.Value(ctxKey{}).(map[string]any)
	if fields["a"] != 1 {
		t.Fatalf("a=%v want 1 (must copy)", fields["a"])
	}
	if _, ok := fields["b"]; ok {
		t.Fatal("b should not appear after external mutate")
	}
}

func TestWithFields_Merge(t *testing.T) {
	ctx := WithFields(context.Background(), map[string]any{"a": 1})
	ctx = WithFields(ctx, map[string]any{"b": 2, "a": 9})
	fields, _ := ctx.Value(ctxKey{}).(map[string]any)
	if fields["a"] != 9 || fields["b"] != 2 {
		t.Fatalf("fields=%v", fields)
	}
}

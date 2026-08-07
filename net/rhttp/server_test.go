package rhttp

import (
	"errors"
	"net/http"
	"testing"
	"time"
)

func TestNormalizeStartErr(t *testing.T) {
	if normalizeStartErr(nil) != nil {
		t.Fatal("nil")
	}
	if normalizeStartErr(http.ErrServerClosed) != nil {
		t.Fatal("ErrServerClosed")
	}
	wrapped := errors.Join(errors.New("wrap"), http.ErrServerClosed)
	if normalizeStartErr(wrapped) != nil {
		t.Fatal("joined ErrServerClosed")
	}
	if normalizeStartErr(errors.New("listen failed")) == nil {
		t.Fatal("other error")
	}
}

func TestGracefulTimeoutDefault(t *testing.T) {
	if (Config{}).gracefulTimeout() != defaultGracefulTimeout {
		t.Fatal("default 10s")
	}
	if (Config{GracefulTimeout: 3 * time.Second}).gracefulTimeout() != 3*time.Second {
		t.Fatal("custom")
	}
}

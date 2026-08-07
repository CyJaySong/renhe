package rdb

import (
	"testing"
	"time"
)

func TestSlowQueryThreshold(t *testing.T) {
	if (Config{}).slowQueryThreshold() != defaultSlowQueryThreshold {
		t.Fatal("default")
	}
	if (Config{SlowQueryThreshold: 2 * time.Second}).slowQueryThreshold() != 2*time.Second {
		t.Fatal("custom")
	}
	if (Config{SlowQueryThreshold: -1}).slowQueryThreshold() != 0 {
		t.Fatal("negative")
	}
}

func TestHealthAndPingTimeouts(t *testing.T) {
	if (Config{}).healthCheckInterval() != defaultHealthCheckInterval {
		t.Fatal("health default")
	}
	if (Config{}).pingTimeout() != defaultPingTimeout {
		t.Fatal("ping default")
	}
	if (Config{HealthCheckInterval: time.Second}).healthCheckInterval() != time.Second {
		t.Fatal("health custom")
	}
	if (Config{PingTimeout: 2 * time.Second}).pingTimeout() != 2*time.Second {
		t.Fatal("ping custom")
	}
}

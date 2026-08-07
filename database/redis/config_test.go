package redis

import "testing"

func TestConfigIsCluster(t *testing.T) {
	cases := []struct {
		mode string
		addr []string
		want bool
	}{
		{"", []string{"a:6379"}, false},
		{"", []string{"a:6379", "b:6379"}, true}, // auto
		{"auto", []string{"a:6379", "b:6379"}, true},
		{"standalone", []string{"a:6379", "b:6379"}, false}, // 显式单机
		{"single", []string{"a:6379", "b:6379"}, false},
		{"cluster", []string{"a:6379"}, true}, // 显式集群
		{"CLUSTER", []string{"a:6379"}, true},
	}
	for _, tc := range cases {
		cfg := Config{Mode: tc.mode, Address: tc.addr}
		if got := cfg.isCluster(); got != tc.want {
			t.Fatalf("mode=%q addrs=%v got=%v want=%v", tc.mode, tc.addr, got, tc.want)
		}
	}
}

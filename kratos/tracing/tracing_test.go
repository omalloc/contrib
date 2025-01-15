package tracing

import (
	"testing"
)

func TestTracingOptions(t *testing.T) {
	tests := []struct {
		name     string
		opts     []Option
		expected traceConfig
	}{
		{
			name: "default config",
			opts: []Option{},
			expected: traceConfig{
				fraction: 1.0,
			},
		},
		{
			name: "with custom endpoint",
			opts: []Option{
				WithEndpoint("http://localhost:14268/api/traces"),
			},
			expected: traceConfig{
				endpoint: "http://localhost:14268/api/traces",
				fraction: 1.0,
			},
		},
		{
			name: "with all options",
			opts: []Option{
				WithEndpoint("http://localhost:14268/api/traces"),
				WithServiceName("test-service"),
				WithRatioBased(0.5),
			},
			expected: traceConfig{
				endpoint:    "http://localhost:14268/api/traces",
				serviceName: "test-service",
				fraction:    0.5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &traceConfig{
				fraction: 1.0,
			}

			// 应用所有选项
			for _, opt := range tt.opts {
				opt(cfg)
			}

			// 验证配置
			if cfg.endpoint != tt.expected.endpoint {
				t.Errorf("endpoint = %v, want %v", cfg.endpoint, tt.expected.endpoint)
			}
			if cfg.serviceName != tt.expected.serviceName {
				t.Errorf("serviceName = %v, want %v", cfg.serviceName, tt.expected.serviceName)
			}
			if cfg.fraction != tt.expected.fraction {
				t.Errorf("fraction = %v, want %v", cfg.fraction, tt.expected.fraction)
			}
		})
	}
}

func TestInitTracer(t *testing.T) {
	// 测试空endpoint的情况
	InitTracer()

	// 测试有效配置
	InitTracer(
		WithEndpoint("http://localhost:14268/api/traces"),
		WithServiceName("test-service"),
		WithRatioBased(0.5),
	)
}

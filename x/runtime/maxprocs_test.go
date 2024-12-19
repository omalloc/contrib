package runtime_test

import (
	"os"
	"testing"

	"github.com/omalloc/contrib/x/runtime"
)

func TestAutoGOMAXPROCS(t *testing.T) {
	tests := []struct {
		name           string
		envMaxThreads  string
		wantMaxThreads int
	}{
		{
			name:           "default_value",
			envMaxThreads:  "",
			wantMaxThreads: 1000,
		},
		{
			name:           "custom_max_threads",
			envMaxThreads:  "500",
			wantMaxThreads: 500,
		},
		{
			name:           "invalid_env_value",
			envMaxThreads:  "invalid",
			wantMaxThreads: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envMaxThreads != "" {
				os.Setenv("APP_MAXTHREADS", tt.envMaxThreads)
				defer os.Unsetenv("APP_MAXTHREADS")
			}

			_, gotMaxThreads := runtime.AutoGOMAXPROCS()
			if gotMaxThreads != tt.wantMaxThreads {
				t.Errorf("AutoGOMAXPROCS() maxThreads = %v, want %v", gotMaxThreads, tt.wantMaxThreads)
			}
		})
	}
}

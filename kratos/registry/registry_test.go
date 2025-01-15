package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
	clientv3 "go.etcd.io/etcd/client/v3"

	protobuf "github.com/omalloc/contrib/protobuf"
)

func TestNewRegistrar(t *testing.T) {
	tests := []struct {
		name   string
		client *clientv3.Client
		config *protobuf.Registry
		want   bool // true if expect non-nil result
	}{
		{
			name:   "nil client should return nil",
			client: nil,
			config: &protobuf.Registry{},
			want:   false,
		},
		{
			name:   "only discovery should return nil",
			client: &clientv3.Client{},
			config: &protobuf.Registry{OnlyDiscovery: true},
			want:   false,
		},
		{
			name:   "normal case without namespace",
			client: &clientv3.Client{},
			config: &protobuf.Registry{},
			want:   true,
		},
		{
			name:   "normal case with namespace",
			client: &clientv3.Client{},
			config: &protobuf.Registry{Namespace: "test"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewRegistrar(tt.client, tt.config)
			if tt.want {
				assert.NotNil(t, got)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

func TestNewDiscovery(t *testing.T) {
	tests := []struct {
		name   string
		client *clientv3.Client
		config *protobuf.Registry
		want   bool // true if expect non-nil result
	}{
		{
			name:   "nil client should return nil",
			client: nil,
			config: &protobuf.Registry{},
			want:   false,
		},
		{
			name:   "normal case without namespace",
			client: &clientv3.Client{},
			config: &protobuf.Registry{},
			want:   true,
		},
		{
			name:   "normal case with namespace",
			client: &clientv3.Client{},
			config: &protobuf.Registry{Namespace: "test"},
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewDiscovery(tt.client, tt.config)
			if tt.want {
				assert.NotNil(t, got)
			} else {
				assert.Nil(t, got)
			}
		})
	}
}

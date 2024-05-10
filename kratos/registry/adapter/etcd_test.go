package adapter_test

import (
	"testing"

	"github.com/omalloc/contrib/kratos/registry/adapter"
	"github.com/omalloc/contrib/protobuf"
)

func TestNewEtcdAdapter(t *testing.T) {
	cli1, _ := adapter.NewEtcdAdapter(&protobuf.Registry{
		Enabled: false,
	}, nil)
	if cli1 != nil {
		t.Fatal("expected nil client")
	}

	cli2, _ := adapter.NewEtcdAdapter(&protobuf.Registry{
		Enabled:   true,
		Endpoints: []string{"etcd://127.0.0.1:2379", "etcd://127.0.0.1:3379"},
	}, func(f func()) {
		f()
	})

	if cli2 == nil {
		t.Fatal("expected non-nil client")
	}
}

package registry_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/omalloc/contrib/v2/kratos/registry"
	"github.com/omalloc/contrib/v2/kratos/registry/adapter"
	"github.com/omalloc/contrib/v2/protobuf"
)

func TestRegistry(t *testing.T) {
	c := &protobuf.Registry{
		Enabled:   false,
		Endpoints: []string{},
	}

	client, err := adapter.NewEtcdAdapter(c, func(f func()) {
		t.Logf("cleanup: %s", time.Now())

		f()
	})

	if errors.Is(err, registry.ErrNoConfigureEndpoints) {
		t.Fatal(err)
	}

	if client != nil {

		rsp, err := client.MemberList(context.Background())

		if err != nil {
			t.Fatal(err)
		}

		for _, m := range rsp.Members {
			t.Logf("member: %s", m.Name)
		}
	}
}

package registry

import (
	"errors"

	clientv3 "go.etcd.io/etcd/client/v3"

	protobuf "github.com/omalloc/contrib/protobuf"
)

var (
	ErrNoConfigureEndpoints = errors.New("registry:: no configure endpoints")
	ErrNoEnabledRegistry    = errors.New("registry:: no enabled registry")
)

var (
	emptyCallback = func() {}
)

func NewEtcd(c *protobuf.Registry) (*clientv3.Client, func(), error) {
	// not enabled registry
	if !c.Enabled {
		return nil, emptyCallback, nil
	}

	if c.GetEndpoints() == nil {
		return nil, emptyCallback, ErrNoConfigureEndpoints
	}

	client, err := clientv3.New(clientv3.Config{
		Endpoints: c.GetEndpoints(),
	})

	if err != nil {
		return nil, emptyCallback, err
	}

	cleanup := func() {
		if client != nil {
			_ = client.Close()
		}
	}

	return client, cleanup, nil
}

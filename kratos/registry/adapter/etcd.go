package adapter

import (
	"errors"

	clientv3 "go.etcd.io/etcd/client/v3"

	"github.com/omalloc/contrib/kratos/registry"
	"github.com/omalloc/contrib/protobuf"
)

type CleanupFunc func(func())

func NewEtcdAdapter(c *protobuf.Registry, fn CleanupFunc) (*clientv3.Client, error) {
	if !c.Enabled {
		return nil, registry.ErrNoEnabledRegistry
	}

	if c.GetEndpoints() == nil {
		return nil, registry.ErrNoConfigureEndpoints
	}

	cli, cleanup, err := registry.NewEtcd(c)
	if err != nil {
		return nil, errors.New("failed to connect to etcd: " + err.Error())
	}

	fn(cleanup)

	// lc.Append(fx.Hook{
	// 	OnStop: func(_ context.Context) error {
	// 		cleanup()
	// 		return nil
	// 	},
	// })
	return cli, nil
}

package adapter

import (
	"context"
	"errors"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/fx"

	"github.com/omalloc/contrib/kratos/registry"
	"github.com/omalloc/contrib/protobuf"
)

func NewEtcdAdapter(lc fx.Lifecycle, c *protobuf.Registry) (*clientv3.Client, error) {
	cli, cleanup, err := registry.NewEtcd(c)
	if err != nil {
		return nil, errors.New("failed to connect to etcd: " + err.Error())
	}

	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			cleanup()
			return nil
		},
	})
	return cli, nil
}

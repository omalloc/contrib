package registry

import (
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"

	protobuf "github.com/omalloc/contrib/protobuf"
)

// NewRegistrar ... init ectd Registry
func NewRegistrar(client *clientv3.Client, c *protobuf.Registry) registry.Registrar {
	if client == nil {
		return nil
	}

	if c.OnlyDiscovery {
		return nil
	}

	opts := make([]etcd.Option, 0)
	if c.Namespace != "" {
		opts = append(opts, etcd.Namespace(c.Namespace))
	}
	return etcd.New(client, opts...)
}

// NewDiscovery ... init etcd Discovery
func NewDiscovery(client *clientv3.Client, c *protobuf.Registry) registry.Discovery {
	if client == nil {
		return nil
	}

	opts := make([]etcd.Option, 0)
	if c.Namespace != "" {
		opts = append(opts, etcd.Namespace(c.Namespace))
	}

	return etcd.New(client, opts...)
}

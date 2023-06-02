package registry

import (
	"github.com/go-kratos/kratos/contrib/registry/etcd/v2"
	"github.com/go-kratos/kratos/v2/registry"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// NewRegistrar ... init ectd Registry
func NewRegistrar(client *clientv3.Client) registry.Registrar {
	if client == nil {
		return nil
	}
	return etcd.New(client)
}

// NewDiscovery ... init etcd Discovery
func NewDiscovery(client *clientv3.Client) registry.Discovery {
	if client == nil {
		return nil
	}
	return etcd.New(client)
}

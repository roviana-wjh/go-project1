package data

import (
	"review-O/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
)

func newConsulDiscovery(c *conf.Registry) registry.Discovery {
	if c == nil {
		panic("registry config is required for Consul discovery")
	}
	cc := c.GetConsul()
	if cc == nil {
		panic("registry.consul is required")
	}
	cfg := api.DefaultConfig()
	cfg.Address = cc.GetAddress()
	cfg.Scheme = cc.GetScheme()
	client, err := api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	return consul.New(client, consul.WithHealthCheck(true))
}

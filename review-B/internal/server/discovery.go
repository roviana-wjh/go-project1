package server

import (
	"review-B/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/hashicorp/consul/api"
)

// NewDiscovery returns Consul-backed service discovery for outbound gRPC clients.
func NewDiscovery(c *conf.Registry) registry.Discovery {
	if c == nil {
		panic("registry config is required")
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

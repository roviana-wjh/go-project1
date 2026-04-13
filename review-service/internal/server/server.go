package server

import (
	"review-service/internal/conf"

	"github.com/go-kratos/kratos/contrib/registry/consul/v2"
	"github.com/go-kratos/kratos/v2/registry"
	"github.com/google/wire"
	"github.com/hashicorp/consul/api"
)

// ProviderSet is server providers.
var ProviderSet = wire.NewSet(NewGRPCServer, NewHTTPServer)

// NewRegistrar returns a Consul-backed service registrar from config.
func NewRegistrar(c *conf.Registry) registry.Registrar {
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

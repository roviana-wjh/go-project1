//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"review-service/internal/biz"
	"review-service/internal/conf"
	"review-service/internal/data"
	"review-service/internal/server"
	"review-service/internal/service"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(
	confServer *conf.Server,
	confData *conf.Data,
	registry *conf.Registry,
	confEs *conf.Elasticsearch,
	logger log.Logger,
) (*kratos.App, func(), error) {
	panic(wire.Build(
		server.NewRegistrar,
		server.ProviderSet,
		data.ProviderSet,
		biz.ProviderSet,
		service.ProviderSet,
		newApp,
	))
}

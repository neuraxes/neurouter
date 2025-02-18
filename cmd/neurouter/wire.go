//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"git.xdea.xyz/Turing/neurouter/internal/biz"
	"git.xdea.xyz/Turing/neurouter/internal/conf"
	"git.xdea.xyz/Turing/neurouter/internal/server"
	"git.xdea.xyz/Turing/neurouter/internal/service"
	"git.xdea.xyz/Turing/neurouter/internal/upstream"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

// wireApp init kratos application.
func wireApp(*conf.Server, *conf.Data, *conf.Upstream, log.Logger) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet /*, data.ProviderSet*/, upstream.ProviderSet, biz.ProviderSet, service.ProviderSet, newApp))
}

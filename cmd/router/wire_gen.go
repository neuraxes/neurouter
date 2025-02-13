// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"git.xdea.xyz/Turing/router/internal/biz"
	"git.xdea.xyz/Turing/router/internal/conf"
	"git.xdea.xyz/Turing/router/internal/server"
	"git.xdea.xyz/Turing/router/internal/service"
	"git.xdea.xyz/Turing/router/internal/upstream"
	"git.xdea.xyz/Turing/router/internal/upstream/openai"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, data *conf.Data, confUpstream *conf.Upstream, logger log.Logger) (*kratos.App, func(), error) {
	openAIChatRepoFactory := openai.NewOpenAIChatRepoFactory()
	anthropicChatRepoFactory := upstream.NewAnthropicChatRepoFactory()
	chatUseCase := biz.NewChatUseCase(confUpstream, openAIChatRepoFactory, anthropicChatRepoFactory, logger)
	routerService := service.NewRouterService(chatUseCase, logger)
	grpcServer := server.NewGRPCServer(confServer, routerService, logger)
	httpServer := server.NewHTTPServer(confServer, routerService, logger)
	app := newApp(logger, grpcServer, httpServer)
	return app, func() {
	}, nil
}

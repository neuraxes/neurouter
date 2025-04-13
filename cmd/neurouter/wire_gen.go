// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/neuraxes/neurouter/internal/biz/chat"
	"github.com/neuraxes/neurouter/internal/biz/embedding"
	"github.com/neuraxes/neurouter/internal/biz/model"
	"github.com/neuraxes/neurouter/internal/conf"
	"github.com/neuraxes/neurouter/internal/data/upstream/anthropic"
	"github.com/neuraxes/neurouter/internal/data/upstream/deepseek"
	"github.com/neuraxes/neurouter/internal/data/upstream/google"
	"github.com/neuraxes/neurouter/internal/data/upstream/neurouter"
	"github.com/neuraxes/neurouter/internal/data/upstream/openai"
	"github.com/neuraxes/neurouter/internal/server"
	"github.com/neuraxes/neurouter/internal/service"
)

import (
	_ "go.uber.org/automaxprocs"
)

// Injectors from wire.go:

// wireApp init kratos application.
func wireApp(confServer *conf.Server, data *conf.Data, upstream *conf.Upstream, logger log.Logger) (*kratos.App, func(), error) {
	upstreamFactory := anthropic.NewAnthropicChatRepoFactory()
	repositoryUpstreamFactory := deepseek.NewDeepSeekChatRepoFactory()
	upstreamFactory2 := google.NewGoogleFactory()
	upstreamFactory3 := neurouter.NewNeurouterFactory()
	upstreamFactory4 := openai.NewOpenAIFactory()
	useCaseImpl := model.NewModelUseCase(upstream, upstreamFactory, repositoryUpstreamFactory, upstreamFactory2, upstreamFactory3, upstreamFactory4, logger)
	useCase := chat.NewChatUseCase(useCaseImpl, logger)
	embeddingUseCase := embedding.NewUseCase(useCaseImpl, logger)
	routerService := service.NewRouterService(useCase, useCaseImpl, embeddingUseCase, logger)
	grpcServer := server.NewGRPCServer(confServer, routerService, logger)
	httpServer := server.NewHTTPServer(confServer, routerService, logger)
	app := newApp(logger, grpcServer, httpServer)
	return app, func() {
	}, nil
}

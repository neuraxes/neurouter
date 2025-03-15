package data

import (
	"git.xdea.xyz/Turing/neurouter/internal/conf"
	"git.xdea.xyz/Turing/neurouter/internal/data/upstream"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/wire"
)

var ProviderSet = wire.NewSet(NewData, upstream.ProviderSet)

type Data struct {
}

func NewData(c *conf.Data, logger log.Logger) (*Data, func(), error) {
	cleanup := func() {
		log.NewHelper(logger).Info("closing the data resources")
	}
	return &Data{}, cleanup, nil
}

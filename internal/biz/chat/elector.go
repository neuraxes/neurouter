package chat

import (
	"git.xdea.xyz/Turing/neurouter/internal/biz/repository"
	"git.xdea.xyz/Turing/neurouter/internal/conf"
)

type Elector interface {
	ElectForChat(uri string) (repo repository.ChatRepo, model *conf.Model, err error)
}

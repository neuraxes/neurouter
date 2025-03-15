package service

import (
	"context"

	v1 "git.xdea.xyz/Turing/neurouter/api/neurouter/v1"
)

func (s *RouterService) ListModel(ctx context.Context, req *v1.ListModelReq) (resp *v1.ListModelResp, err error) {
	models, err := s.model.ListAvailableModels(ctx)
	if err != nil {
		return
	}

	respModels := make([]*v1.ModelSpec, len(models))
	for i, m := range models {
		respModels[i] = (*v1.ModelSpec)(m)
	}

	resp = &v1.ListModelResp{
		Models: respModels,
	}
	return
}

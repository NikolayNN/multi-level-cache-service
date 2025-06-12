package integration

import (
	"aur-cache-service/api/dto"
	"context"
)

type Controller interface {
	GetAll(ctx context.Context, reqs []*dto.ResolvedCacheId) *dto.GetResult
}

func CreateController(service Service) Controller {
	return &ControllerImpl{
		service: service,
	}
}

type ControllerImpl struct {
	service Service
}

func (c *ControllerImpl) GetAll(ctx context.Context, reqs []*dto.ResolvedCacheId) *dto.GetResult {
	return c.service.GetAll(ctx, reqs)
}

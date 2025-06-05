package integration

import (
	"aur-cache-service/api/dto"
	"context"
)

type Controller interface {
	GetAll(reqs []*dto.ResolvedCacheId) *dto.GetResult
}

type ControllerImpl struct {
	service Service
}

func (c *ControllerImpl) GetAll(ctx context.Context, reqs []*dto.ResolvedCacheId) *dto.GetResult {
	return c.service.GetAll(ctx, reqs)
}

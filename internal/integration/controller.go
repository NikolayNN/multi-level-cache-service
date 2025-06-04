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

func (c *ControllerImpl) GetAll(reqs []*dto.ResolvedCacheId) *dto.GetResult {
	ctx := context.Background()
	return c.service.GetAll(ctx, reqs)
}

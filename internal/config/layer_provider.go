package config

import "fmt"

type LayerProvider struct {
	Mode     LayerMode
	Provider Provider
}

type LayerProviderService struct {
	layerProviders []*LayerProvider
}

func NewLayerProviderService(cfg *AppConfig) *LayerProviderService {
	providersMap := providersToMap(cfg.Provider)
	layerProviders := createLayerProviders(cfg.Layers, providersMap)
	return &LayerProviderService{
		layerProviders: layerProviders,
	}
}

func createLayerProviders(layers []Layer, providersMap map[string]Provider) (layerProviders []*LayerProvider) {
	layerProviders = make([]*LayerProvider, len(layers))
	for i, layer := range layers {
		provider, found := providersMap[layer.Name]
		if !found {
			panic(fmt.Errorf("can't create layer providers. can't find provider with name: %q", layer.Name))
		}
		layerProviders[i] = &LayerProvider{
			Mode:     layer.Mode,
			Provider: provider,
		}
	}
	return
}

func providersToMap(providers []Provider) (providersMap map[string]Provider) {
	providersMap = make(map[string]Provider)
	for _, provider := range providers {
		providersMap[provider.GetName()] = provider
	}
	return
}

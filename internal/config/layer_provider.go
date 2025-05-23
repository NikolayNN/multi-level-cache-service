package config

import "fmt"

type LayerProvider struct {
	Mode     LayerMode
	Provider Provider
}

func LoadLayersProviders(path string) (layerProviders []LayerProvider, err error) {
	layerList, err := loadLayers(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load layers: %w", err)
	}
	pds, err := LoadProviders(path)
	if err != nil {
		return nil, fmt.Errorf("failed to load providers: %w", err)
	}
	providersMap := providersToMap(pds.Providers)
	layerProviders, err = createLayerProviders(layerList.Layers, providersMap)
	return
}

func createLayerProviders(layers []Layer, providersMap map[string]Provider) (layerProviders []LayerProvider, err error) {
	layerProviders = make([]LayerProvider, len(layers))
	for i, layer := range layers {
		provider, found := providersMap[layer.Name]
		if !found {
			return nil, fmt.Errorf("Can't create layer providers. can't find provider with name: %q", layer.Name)
		}
		layerProviders[i] = LayerProvider{
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

package put

import (
	"aur-cache-service/internal/config"
)

////////////////////////////////////////////////////////////////////////////////
// Factory-helpers
////////////////////////////////////////////////////////////////////////////////

// createConcreteResolvers строит срез конкретных резолверов по количеству
// провайдеров слоёв (layerProviders). Отключённые слои получают
// ConcreteResolverDisabled.
func createConcreteResolvers(cacheConfigService config.CacheService, layerProviders []config.LayerProvider) (concreteResolvers []ConcreteResolver) {
	concreteResolvers = make([]ConcreteResolver, len(layerProviders))
	for i, lp := range layerProviders {
		var concreteResolver ConcreteResolver
		if lp.Mode == config.LayerModeDisabled {
			concreteResolver = NewConcreteResolverDisabled()
		} else {
			concreteResolver = NewConcreteResolverLevel(cacheConfigService, i)
		}
		concreteResolvers[i] = concreteResolver
	}
	return
}

// createMainResolver — удобный конструктор верхнеуровневого резолвера.
func createMainResolver(cacheConfigService config.CacheService, layerProviders []config.LayerProvider) *MainResolver {
	resolvers := createConcreteResolvers(cacheConfigService, layerProviders)
	return &MainResolver{
		resolvers: resolvers,
	}
}

////////////////////////////////////////////////////////////////////////////////
// Main resolver (агрегирует резолверы конкретных слоёв)
////////////////////////////////////////////////////////////////////////////////

// MainResolver управляет последовательным выполнением резолверов
// для каждого уровня кэша.
type MainResolver struct {
	resolvers []ConcreteResolver
}

// NewMainResolver создаёт MainResolver из уже подготовленных резолверов.
func NewMainResolver(resolvers []ConcreteResolver) *MainResolver {
	return &MainResolver{
		resolvers: resolvers,
	}
}

// Resolve выполняет разрешение по всем доступным уровням.
func (m *MainResolver) Resolve(reqs []CacheReq) (resolvedMatrix [][]CacheReqResolved) {
	resolvedMatrix = m.ResolveTo(reqs, m.maxLevel())
	return
}

// ResolveTo выполняет разрешение до указанного уровня включительно.
func (m *MainResolver) ResolveTo(reqs []CacheReq, toLevel int) (resolvedMatrix [][]CacheReqResolved) {
	resolvedMatrix = m.initMatrix(toLevel, len(reqs))
	for resolverIndex := 0; resolverIndex <= toLevel; resolverIndex++ {
		resolver := m.resolvers[resolverIndex]
		resolvedReqs := resolver.resolve(reqs)
		resolvedMatrix[resolverIndex] = resolvedReqs
	}
	return
}

// initMatrix подготавливает (lvl+1)×cols матрицу результатов.
func (m *MainResolver) initMatrix(toLevel int, cols int) (resolvedMatrix [][]CacheReqResolved) {
	if toLevel > m.maxLevel() {
		toLevel = m.maxLevel()
	}
	resolvedMatrix = make([][]CacheReqResolved, toLevel+1)
	for i := 0; i <= toLevel; i++ {
		resolvedMatrix[i] = make([]CacheReqResolved, cols)
	}
	return
}

// maxLevel возвращает индекс последнего доступного уровня.
func (m *MainResolver) maxLevel() (maxLevel int) {
	maxLevel = len(m.resolvers) - 1
	return
}

////////////////////////////////////////////////////////////////////////////////
// Resolver contract
////////////////////////////////////////////////////////////////////////////////

// ConcreteResolver описывает единый контракт «разрешителя» конкретного слоя.
type ConcreteResolver interface {
	resolve(reqs []CacheReq) []CacheReqResolved
}

////////////////////////////////////////////////////////////////////////////////
// Disabled-layer resolver
////////////////////////////////////////////////////////////////////////////////

// ConcreteResolverDisabled всегда возвращает пустой результат
// (используется для отключённых слоёв кэша).
type ConcreteResolverDisabled struct {
}

func NewConcreteResolverDisabled() *ConcreteResolverDisabled {
	return &ConcreteResolverDisabled{}
}

func (r *ConcreteResolverDisabled) resolve(reqs []CacheReq) (resolvedRequests []CacheReqResolved) {
	return []CacheReqResolved{}
}

////////////////////////////////////////////////////////////////////////////////
// Active-layer resolver
////////////////////////////////////////////////////////////////////////////////

// ConcreteResolverLevel обрабатывает запросы для конкретного уровня кэша.
type ConcreteResolverLevel struct {
	level              int
	cacheConfigService config.CacheService
}

// NewConcreteResolverLevel создаёт резолвер заданного уровня.
func NewConcreteResolverLevel(cacheConfigService config.CacheService, level int) *ConcreteResolverLevel {
	return &ConcreteResolverLevel{
		level:              level,
		cacheConfigService: cacheConfigService,
	}
}

// resolve пытается «прорезолвить» каждый запрос на своём уровне.
func (r *ConcreteResolverLevel) resolve(reqs []CacheReq) (resolvedRequests []CacheReqResolved) {
	for _, req := range reqs {
		resolvedReq := r.resolveSingle(req)
		if resolvedReq != nil {
			resolvedRequests = append(resolvedRequests, *resolvedReq)
		}
	}
	return
}

// resolveSingle возвращает результат, если слой включён,
// иначе — nil.
func (r *ConcreteResolverLevel) resolveSingle(req CacheReq) (resolvedReq *CacheReqResolved) {
	levelOpts := r.cacheConfigService.GetByName(req.GetCacheName()).Layers[r.level]
	if !levelOpts.Enabled {
		return nil
	}
	resolvedReq = &CacheReqResolved{
		Req:      &req,
		CacheKey: r.cacheConfigService.CacheKey(req),
		Ttl:      levelOpts.TTL,
	}
	return
}

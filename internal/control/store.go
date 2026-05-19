package control

import (
	"fmt"
	"sync"
	"time"
)

type Store interface {
	UpsertConfig(key, value, scope string) ConfigEntry
	GetConfig(key string) (ConfigEntry, bool)
	ListConfigs() []ConfigEntry
	CreateABTest(test ABTest) ABTest
	ListABTests() []ABTest
	CreateRollout(rollout Rollout) Rollout
	ListRollouts() []Rollout
	ListServices() []ServiceState
	ApplyServiceAction(name, action string) (ServiceState, error)
	ListNodes() []NodeLoad
	CreateUpdatePlan(plan UpdatePlan) UpdatePlan
	ListUpdatePlans() []UpdatePlan
	CreateServiceVersion(version ServiceVersion) ServiceVersion
	ListServiceVersions() []ServiceVersion
	CreateRelease(release Release) (Release, error)
	ListReleases() []Release
	CreateRollback(rollback Rollback) (Rollback, error)
	ListRollbacks() []Rollback
}

type MemoryStore struct {
	mu        sync.RWMutex
	configs   map[string]ConfigEntry
	abTests   []ABTest
	rollouts  []Rollout
	services  map[string]ServiceState
	nodes     []NodeLoad
	updates   []UpdatePlan
	versions  []ServiceVersion
	releases  []Release
	rollbacks []Rollback
	nextID    int64
}

func NewMemoryStore() *MemoryStore {
	now := time.Now().UTC()
	return &MemoryStore{
		configs: map[string]ConfigEntry{},
		services: map[string]ServiceState{
			"gateway-service":  {Name: "gateway-service", Status: "running", Version: "dev", UpdatedAt: now},
			"platform-service": {Name: "platform-service", Status: "running", Version: "dev", UpdatedAt: now},
			"gameplay-service": {Name: "gameplay-service", Status: "running", Version: "dev", UpdatedAt: now},
			"wallet-service":   {Name: "wallet-service", Status: "running", Version: "dev", UpdatedAt: now},
		},
		nodes: []NodeLoad{
			{NodeID: "local-1", Service: "gameplay-service", Status: "healthy", CPUPercent: 1, MemPercent: 1, UpdatedAt: now},
		},
		nextID: 1,
	}
}

func (s *MemoryStore) UpsertConfig(key, value, scope string) ConfigEntry {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	entry := s.configs[key]
	entry.Key = key
	entry.Value = value
	entry.Scope = scope
	entry.Version++
	entry.UpdatedAt = now
	s.configs[key] = entry
	return entry
}

func (s *MemoryStore) GetConfig(key string) (ConfigEntry, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, ok := s.configs[key]
	return entry, ok
}

func (s *MemoryStore) ListConfigs() []ConfigEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()

	configs := make([]ConfigEntry, 0, len(s.configs))
	for _, config := range s.configs {
		configs = append(configs, config)
	}
	return configs
}

func (s *MemoryStore) CreateABTest(test ABTest) ABTest {
	s.mu.Lock()
	defer s.mu.Unlock()

	test.ID = s.nextIDString("ab")
	test.Status = defaultString(test.Status, "draft")
	test.CreatedAt = time.Now().UTC()
	s.abTests = append(s.abTests, test)
	return test
}

func (s *MemoryStore) ListABTests() []ABTest {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tests := make([]ABTest, len(s.abTests))
	copy(tests, s.abTests)
	return tests
}

func (s *MemoryStore) CreateRollout(rollout Rollout) Rollout {
	s.mu.Lock()
	defer s.mu.Unlock()

	rollout.ID = s.nextIDString("rollout")
	rollout.Status = defaultString(rollout.Status, "draft")
	rollout.CreatedAt = time.Now().UTC()
	s.rollouts = append(s.rollouts, rollout)
	return rollout
}

func (s *MemoryStore) ListRollouts() []Rollout {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rollouts := make([]Rollout, len(s.rollouts))
	copy(rollouts, s.rollouts)
	return rollouts
}

func (s *MemoryStore) ListServices() []ServiceState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	services := make([]ServiceState, 0, len(s.services))
	for _, service := range s.services {
		services = append(services, service)
	}
	return services
}

func (s *MemoryStore) ApplyServiceAction(name, action string) (ServiceState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	service, ok := s.services[name]
	if !ok {
		return ServiceState{}, fmt.Errorf("service %q not found", name)
	}
	switch action {
	case "start", "restart":
		service.Status = "running"
	case "stop":
		service.Status = "stopped"
	default:
		return ServiceState{}, fmt.Errorf("unsupported action %q", action)
	}
	service.UpdatedAt = time.Now().UTC()
	s.services[name] = service
	return service, nil
}

func (s *MemoryStore) ListNodes() []NodeLoad {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := make([]NodeLoad, len(s.nodes))
	copy(nodes, s.nodes)
	return nodes
}

func (s *MemoryStore) CreateUpdatePlan(plan UpdatePlan) UpdatePlan {
	s.mu.Lock()
	defer s.mu.Unlock()

	plan.ID = s.nextIDString("update")
	plan.Strategy = defaultString(plan.Strategy, "rolling")
	plan.Status = defaultString(plan.Status, "planned")
	plan.RequestedAt = time.Now().UTC()
	s.updates = append(s.updates, plan)
	return plan
}

func (s *MemoryStore) ListUpdatePlans() []UpdatePlan {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plans := make([]UpdatePlan, len(s.updates))
	copy(plans, s.updates)
	return plans
}

func (s *MemoryStore) CreateServiceVersion(version ServiceVersion) ServiceVersion {
	s.mu.Lock()
	defer s.mu.Unlock()

	version.ID = s.nextIDString("version")
	version.Status = defaultString(version.Status, "available")
	version.CreatedAt = time.Now().UTC()
	s.versions = append(s.versions, version)
	return version
}

func (s *MemoryStore) ListServiceVersions() []ServiceVersion {
	s.mu.RLock()
	defer s.mu.RUnlock()

	versions := make([]ServiceVersion, len(s.versions))
	copy(versions, s.versions)
	return versions
}

func (s *MemoryStore) CreateRelease(release Release) (Release, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	service, ok := s.services[release.Service]
	if !ok {
		return Release{}, fmt.Errorf("service %q not found", release.Service)
	}
	release.ID = s.nextIDString("release")
	release.PreviousVersion = service.Version
	release.Strategy = defaultString(release.Strategy, "rolling")
	release.Status = defaultString(release.Status, "released")
	release.CreatedAt = time.Now().UTC()
	s.releases = append(s.releases, release)

	service.Version = release.Version
	service.Status = "running"
	service.UpdatedAt = release.CreatedAt
	s.services[release.Service] = service
	return release, nil
}

func (s *MemoryStore) ListReleases() []Release {
	s.mu.RLock()
	defer s.mu.RUnlock()

	releases := make([]Release, len(s.releases))
	copy(releases, s.releases)
	return releases
}

func (s *MemoryStore) CreateRollback(rollback Rollback) (Rollback, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	service, ok := s.services[rollback.Service]
	if !ok {
		return Rollback{}, fmt.Errorf("service %q not found", rollback.Service)
	}
	rollback.ID = s.nextIDString("rollback")
	rollback.FromVersion = service.Version
	rollback.Status = defaultString(rollback.Status, "rolled_back")
	rollback.CreatedAt = time.Now().UTC()
	s.rollbacks = append(s.rollbacks, rollback)

	service.Version = rollback.TargetVersion
	service.Status = "running"
	service.UpdatedAt = rollback.CreatedAt
	s.services[rollback.Service] = service
	return rollback, nil
}

func (s *MemoryStore) ListRollbacks() []Rollback {
	s.mu.RLock()
	defer s.mu.RUnlock()

	rollbacks := make([]Rollback, len(s.rollbacks))
	copy(rollbacks, s.rollbacks)
	return rollbacks
}

func (s *MemoryStore) nextIDString(prefix string) string {
	id := fmt.Sprintf("%s-%d", prefix, s.nextID)
	s.nextID++
	return id
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

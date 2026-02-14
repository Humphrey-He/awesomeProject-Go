package config_registry

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	ErrConfigNotFound   = errors.New("config not found")
	ErrInstanceNotFound = errors.New("instance not found")
)

// ---------------- Config Center ----------------

type ConfigEvent struct {
	Key       string
	Value     string
	Version   int64
	UpdatedAt time.Time
}

type ConfigCenter struct {
	mu       sync.RWMutex
	values   map[string]ConfigEvent
	watchers map[string][]chan ConfigEvent
}

func NewConfigCenter() *ConfigCenter {
	return &ConfigCenter{
		values:   make(map[string]ConfigEvent),
		watchers: make(map[string][]chan ConfigEvent),
	}
}

func (c *ConfigCenter) Set(key, value string) ConfigEvent {
	c.mu.Lock()
	defer c.mu.Unlock()

	prev := c.values[key]
	ev := ConfigEvent{
		Key:       key,
		Value:     value,
		Version:   prev.Version + 1,
		UpdatedAt: time.Now(),
	}
	c.values[key] = ev

	for _, ch := range c.watchers[key] {
		select {
		case ch <- ev:
		default:
			// drop slow watcher event in MVP for non-blocking updates
		}
	}
	return ev
}

func (c *ConfigCenter) Get(key string) (ConfigEvent, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	v, ok := c.values[key]
	if !ok {
		return ConfigEvent{}, ErrConfigNotFound
	}
	return v, nil
}

func (c *ConfigCenter) Watch(key string, buffer int) (<-chan ConfigEvent, func()) {
	if buffer < 1 {
		buffer = 1
	}
	ch := make(chan ConfigEvent, buffer)

	c.mu.Lock()
	c.watchers[key] = append(c.watchers[key], ch)
	c.mu.Unlock()

	cancel := func() {
		c.mu.Lock()
		defer c.mu.Unlock()
		list := c.watchers[key]
		for i := range list {
			if list[i] == ch {
				list = append(list[:i], list[i+1:]...)
				break
			}
		}
		c.watchers[key] = list
		close(ch)
	}
	return ch, cancel
}

func (c *ConfigCenter) Snapshot() map[string]ConfigEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make(map[string]ConfigEvent, len(c.values))
	for k, v := range c.values {
		out[k] = v
	}
	return out
}

// ---------------- Service Registry ----------------

type Instance struct {
	Service  string
	ID       string
	Endpoint string
	Meta     map[string]string
	ExpireAt time.Time
}

type Registry struct {
	mu        sync.RWMutex
	services  map[string]map[string]*Instance // service -> id -> instance
	stopCh    chan struct{}
	stopped   bool
	cleanTick time.Duration
}

func NewRegistry(cleanTick time.Duration) *Registry {
	if cleanTick <= 0 {
		cleanTick = 200 * time.Millisecond
	}
	r := &Registry{
		services:  make(map[string]map[string]*Instance),
		stopCh:    make(chan struct{}),
		cleanTick: cleanTick,
	}
	go r.cleaner()
	return r
}

func (r *Registry) Register(service, id, endpoint string, ttl time.Duration, meta map[string]string) {
	if ttl <= 0 {
		ttl = 1 * time.Second
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.services[service]; !ok {
		r.services[service] = make(map[string]*Instance)
	}
	r.services[service][id] = &Instance{
		Service:  service,
		ID:       id,
		Endpoint: endpoint,
		Meta:     cloneMap(meta),
		ExpireAt: time.Now().Add(ttl),
	}
}

func (r *Registry) Heartbeat(service, id string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = 1 * time.Second
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	svc, ok := r.services[service]
	if !ok {
		return ErrInstanceNotFound
	}
	ins, ok := svc[id]
	if !ok {
		return ErrInstanceNotFound
	}
	ins.ExpireAt = time.Now().Add(ttl)
	return nil
}

func (r *Registry) Discover(service string) []Instance {
	now := time.Now()
	r.mu.RLock()
	defer r.mu.RUnlock()
	svc := r.services[service]
	out := make([]Instance, 0, len(svc))
	for _, ins := range svc {
		if now.Before(ins.ExpireAt) {
			out = append(out, *ins)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (r *Registry) Deregister(service, id string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if svc, ok := r.services[service]; ok {
		delete(svc, id)
		if len(svc) == 0 {
			delete(r.services, service)
		}
	}
}

func (r *Registry) Stop() {
	r.mu.Lock()
	if r.stopped {
		r.mu.Unlock()
		return
	}
	r.stopped = true
	close(r.stopCh)
	r.mu.Unlock()
}

func (r *Registry) cleaner() {
	ticker := time.NewTicker(r.cleanTick)
	defer ticker.Stop()
	for {
		select {
		case <-r.stopCh:
			return
		case <-ticker.C:
			now := time.Now()
			r.mu.Lock()
			for service, instances := range r.services {
				for id, ins := range instances {
					if now.After(ins.ExpireAt) {
						delete(instances, id)
					}
				}
				if len(instances) == 0 {
					delete(r.services, service)
				}
			}
			r.mu.Unlock()
		}
	}
}

func cloneMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

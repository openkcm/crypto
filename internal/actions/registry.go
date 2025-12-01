package actions

import (
	"maps"
	"sync"

	"github.com/openkcm/crypto/kmip"
)

var (
	registeredActions = map[kmip.Operation]Action{}
)

func init() {
	Register(&createAction{})
}

func Register(action Action) {
	registeredActions[action.Operation()] = action
}

type ReadRegistry interface {
	Lookup(operation kmip.Operation) Action
}

type Registry interface {
	ReadRegistry

	Add(operations ...kmip.Operation)
	Remove(operations ...kmip.Operation)
	KeepOnly(operations ...kmip.Operation)
}

type registry struct {
	mu      sync.RWMutex
	actions map[kmip.Operation]Action
}

func NewRegistry() Registry {
	tmp := make(map[kmip.Operation]Action)

	maps.Copy(tmp, registeredActions)
	return &registry{
		mu:      sync.RWMutex{},
		actions: tmp,
	}
}

func (r *registry) Add(operations ...kmip.Operation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, op := range operations {
		if action, ok := registeredActions[op]; ok {
			r.actions[op] = action
		}
	}
}

func (r *registry) Remove(operations ...kmip.Operation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, op := range operations {
		delete(r.actions, op)
	}
}

func (r *registry) KeepOnly(operations ...kmip.Operation) {
	r.mu.Lock()
	defer r.mu.Unlock()

	keepOps := make(map[kmip.Operation]struct{})

	for _, op := range operations {
		keepOps[op] = struct{}{}
	}

	for op := range r.actions {
		if _, ok := keepOps[op]; !ok {
			delete(r.actions, op)
		}
	}
}

func (r *registry) Lookup(operation kmip.Operation) Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	action, ok := r.actions[operation]
	if !ok {
		return nil
	}
	return action
}

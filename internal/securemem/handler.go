package securemem

import (
	"context"
	"log"
	"runtime/secret"
	"sync"
)

type (
	Handler        func(context.Context, *HandlerRequest) error
	HandlerRequest struct {
		mux       sync.RWMutex
		vault     *MemVault
		toPersist map[string]struct{}
	}
	HandlerResponse = HandlerRequest
)

func newHandlerRequest() *HandlerRequest {
	return &HandlerRequest{
		toPersist: make(map[string]struct{}),
		vault:     NewMemVault(),
	}
}

func (r *HandlerRequest) Put(name string, data []byte) error {
	return r.vault.Put(name, data)
}

func (r *HandlerRequest) Get(name string) ([]byte, bool) {
	return r.vault.Get(name)
}

func (r *HandlerRequest) Persist(name string, data []byte) error {
	err := r.Put(name, data)
	if err != nil {
		return err
	}

	r.mux.Lock()
	defer r.mux.Unlock()

	r.toPersist[name] = struct{}{}
	return nil
}

func (r *HandlerRequest) Reserve(name string, size int) ([]byte, error) {
	return r.vault.Reserve(name, size)
}

func (r *HandlerRequest) DestroyAll() error {
	r.mux.Lock()
	defer r.mux.Unlock()

	clear(r.toPersist)
	return r.vault.DestroyAll()
}

func (r *HandlerRequest) Destroy(name string) error {
	r.mux.Lock()
	defer r.mux.Unlock()

	delete(r.toPersist, name)
	return r.vault.Destroy(name)
}

func Run(ctx context.Context, handler Handler) (*HandlerResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req := newHandlerRequest()

	defer func() {
		_ = req.DestroyAll()
	}()

	var err error
	secret.Do(func() {
		err = handler(ctx, req)
	})

	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	resp := newHandlerRequest()
	err = transferPersistedValues(req, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func transferPersistedValues(req *HandlerRequest, resp *HandlerResponse) error {
	req.mux.RLock()
	defer req.mux.RUnlock()

	if len(req.toPersist) > 0 {
		for name := range req.toPersist {
			data, ok := req.Get(name)
			if !ok {
				continue
			}
			err := resp.Put(name, data)
			if err != nil {
				err1 := resp.DestroyAll()
				log.Printf("failed to destroy response after transfer error: %v", err1)
				return err
			}
		}
	}
	return nil
}

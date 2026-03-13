package securemem

import (
	"context"
	"log/slog"
	"runtime/secret"
)

type (
	Handler        func(context.Context, *HandlerRequest) error
	HandlerRequest struct {
		tmpVault *MemVault
		vault    *MemVault
	}
	HandlerResponse struct {
		vault *MemVault
	}
)

func newHandlerRequest() *HandlerRequest {
	return &HandlerRequest{
		tmpVault: NewMemVault(),
		vault:    NewMemVault(),
	}
}

func (r *HandlerRequest) PersistentVault() *MemVault {
	return r.vault
}

func (r *HandlerRequest) TmpVault() *MemVault {
	return r.tmpVault
}

func (r *HandlerResponse) MemVault() *MemVault {
	return r.vault
}

func Run(ctx context.Context, handler Handler) (*HandlerResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req := newHandlerRequest()
	var resp *HandlerResponse

	defer func() {
		err := req.TmpVault().DestroyAll()
		if err != nil {
			slog.Error("failed to destroy temp vault", "error", err)
		}
		if resp == nil {
			err = req.PersistentVault().DestroyAll()
			if err != nil {
				slog.Error("failed to destroy persistent vault after handler", "error", err)
			}
		}
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

	persistentVault := req.PersistentVault()
	err = persistentVault.MarkAllReadOnly()
	if err != nil {
		return nil, err
	}

	resp = &HandlerResponse{vault: persistentVault}

	return resp, nil
}

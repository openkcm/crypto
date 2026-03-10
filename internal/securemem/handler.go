package securemem

import (
	"context"
	"runtime/secret"
)

type (
	SessionHandler func(context.Context, *VaultSession) error
	VaultSession   struct {
		vault     *Vault
		toPersist map[string]struct{}
	}
	VaultState = VaultSession
)

func newVaultSession() *VaultSession {
	return &VaultSession{
		toPersist: make(map[string]struct{}),
		vault:     NewVault(),
	}
}

func (r *VaultSession) Put(name string, data []byte) error {
	return r.vault.Put(name, data)
}

func (r *VaultSession) Get(name string) ([]byte, bool) {
	return r.vault.Get(name)
}

func (r *VaultSession) Persist(name string, data []byte) error {
	err := r.Put(name, data)
	if err != nil {
		return err
	}
	r.toPersist[name] = struct{}{}
	return nil
}

func (r *VaultSession) Reserve(name string, size int) ([]byte, error) {
	return r.vault.Reserve(name, size)
}

func (r *VaultSession) DestroyAll() error {
	clear(r.toPersist)
	return r.vault.DestroyAll()
}

func (r *VaultSession) Destroy(name string) error {
	delete(r.toPersist, name)
	return r.vault.Destroy(name)
}

func Run(ctx context.Context, handler SessionHandler) (*VaultState, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	sess := newVaultSession()
	state := newVaultSession()

	defer func() {
		_ = sess.DestroyAll()
	}()

	var err error
	secret.Do(func() {
		err = handler(ctx, sess)
	})

	if err != nil {
		return nil, err
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	err = transferPersistedValues(sess, state)
	if err != nil {
		return nil, err
	}

	return state, nil
}

func transferPersistedValues(sess *VaultSession, state *VaultSession) error {
	if len(sess.toPersist) > 0 {
		for name := range sess.toPersist {
			data, ok := sess.Get(name)
			if !ok {
				continue
			}
			err := state.Put(name, data)
			if err != nil {
				_ = state.DestroyAll()
				return err
			}
		}
	}
	return nil
}

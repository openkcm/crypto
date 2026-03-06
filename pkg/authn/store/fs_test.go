package store_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/openkcm/krypton/pkg/authn"
	"github.com/openkcm/krypton/pkg/authn/store"
	"github.com/stretchr/testify/assert"
)

func TestFS_StoreAndGet(t *testing.T) {
	// given
	dir := t.TempDir()
	s := store.NewFSWithDir(dir)
	token := &authn.Token{
		Type:      "bearer",
		Value:     []byte("test-token-value"),
		ExpiredAt: 1234567890,
	}

	// when
	err := s.Store(t.Context(), token)

	// then
	assert.NoError(t, err)
	_, err = os.Stat(filepath.Join(dir, store.TokenFileName))
	assert.NoError(t, err)

	// when
	got, err := s.Get(t.Context())

	// then
	assert.NoError(t, err)
	assert.Equal(t, token.Type, got.Type)
	assert.Equal(t, token.Value, got.Value)
	assert.Equal(t, token.ExpiredAt, got.ExpiredAt)
}

func TestFS_Store_NilToken(t *testing.T) {
	// given
	dir := t.TempDir()
	s := store.NewFSWithDir(dir)

	// when
	err := s.Store(t.Context(), nil)

	// then
	assert.Error(t, err)
	assert.ErrorIs(t, err, authn.ErrTokenNil)
}

func TestFS_Get_NotFound(t *testing.T) {
	// given
	dir := t.TempDir()
	s := store.NewFSWithDir(dir)

	// when
	_, err := s.Get(t.Context())

	// then
	assert.ErrorIs(t, err, authn.ErrTokenNotFound)
}

func TestFS_Delete(t *testing.T) {
	tests := []struct {
		name       string
		storeFirst bool
	}{
		{
			name:       "deletes existing token",
			storeFirst: true,
		},
		{
			name:       "succeeds when token does not exist",
			storeFirst: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// given
			dir := t.TempDir()
			s := store.NewFSWithDir(dir)

			if tt.storeFirst {
				err := s.Store(t.Context(), &authn.Token{Type: "bearer", Value: []byte("test")})
				assert.NoError(t, err)
			}

			// when
			err := s.Delete(t.Context())

			// then
			assert.NoError(t, err)
			_, err = s.Get(t.Context())
			assert.ErrorIs(t, err, authn.ErrTokenNotFound)
		})
	}
}

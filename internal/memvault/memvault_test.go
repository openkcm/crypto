package memvault_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openkcm/krypton/internal/memvault"
)

func TestInit(t *testing.T) {
	t.Run("should return error if the input is nil", func(t *testing.T) {
		// when
		vault, err := memvault.New(nil)
		// then
		assert.ErrorIs(t, err, memvault.ErrInvalidInput)
		assert.Nil(t, vault)
	})

	t.Run("should initialize without any error", func(t *testing.T) {
		// given
		input := []byte("secret")

		// when
		vault, err := memvault.New(input)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, vault)
	})

	t.Run("should clear input secret to zeros", func(t *testing.T) {
		// given
		input := []byte("secret")

		// when
		_, err := memvault.New(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0}, input) // original input should be cleared
	})
}

func TestRead(t *testing.T) {
	t.Run("should able to read multiple times from the vault", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.New(input)
		assert.NoError(t, err)

		// when
		for range 2 {
			err = vault.Read(func(data []byte) error {
				// then
				assert.Equal(t, []byte("secret"), data)
				return nil
			})
			// then
			assert.NoError(t, err)
		}
	})

	t.Run("should return error if the input function returns error", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.New(input)
		assert.NoError(t, err)

		// when
		err = vault.Read(func(data []byte) error {
			return assert.AnError
		})

		// then
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestWipe(t *testing.T) {
	t.Run("should be able to wipe the secret", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.New(input)
		assert.NoError(t, err)

		// when
		err = vault.Wipe()

		// then
		assert.NoError(t, err)
	})

	t.Run("should be able to wipe secrets multiple times", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.New(input)
		assert.NoError(t, err)

		for range 2 {
			// when
			err = vault.Wipe()

			// then
			assert.NoError(t, err)
		}
	})

	t.Run("should not be able to read secrets after a wipe", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.New(input)
		assert.NoError(t, err)

		err = vault.Wipe()
		assert.NoError(t, err)

		// when then
		// trying to read after wipe
		err = vault.Read(func(data []byte) error {
			// then
			assert.Fail(t, "this function should not be called as the bytes have been wiped") // after wipe, the vault should be cleared
			return nil
		})
		assert.ErrorIs(t, err, memvault.ErrVaultWiped) // should return error when trying to read from a wiped vault
	})
}

func TestReadAndWipe(t *testing.T) {
	t.Run("should be able to ReadAndWipe the secret", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.New(input)
		assert.NoError(t, err)

		// when
		err = vault.ReadAndWipe(func(data []byte) error {
			// then
			assert.Equal(t, []byte("secret"), data) // should read the secret before wiping
			return nil
		})

		// then
		assert.NoError(t, err)
	})

	t.Run("should return error if the input function returns error", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.New(input)
		assert.NoError(t, err)

		// when
		err = vault.ReadAndWipe(func(data []byte) error {
			return assert.AnError
		})

		// then
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("should not be able read secrets if the input function returned an error first time", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.New(input)
		assert.NoError(t, err)

		// when
		err = vault.ReadAndWipe(func(data []byte) error {
			return assert.AnError
		})

		// then
		assert.ErrorIs(t, err, assert.AnError)

		err = vault.ReadAndWipe(func(data []byte) error {
			assert.Fail(t, "this function should not be called as the bytes have been wiped") // after wipe, the vault should be cleared
			return nil
		})

		assert.ErrorIs(t, err, memvault.ErrVaultWiped)
	})

	t.Run("should not be able to read secrets after a ReadAndWipe", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.New(input)
		assert.NoError(t, err)

		err = vault.ReadAndWipe(func(data []byte) error {
			// then
			assert.Equal(t, []byte("secret"), data) // should read the secret before wiping
			return nil
		})

		// then
		assert.NoError(t, err)

		// when then
		err = vault.Read(func(data []byte) error {
			// then
			assert.Fail(t, "this function should not be called as the bytes have been wiped") // after wipe, the vault should be cleared
			return nil
		})
		assert.ErrorIs(t, err, memvault.ErrVaultWiped) // should return error when trying to read from a wiped vault
	})
}

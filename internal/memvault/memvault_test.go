package memvault_test

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/openkcm/krypton/internal/memvault"
)

func TestExampleEncryption(t *testing.T) {
	// 32-byte masterKey for AES-256
	masterKey := []byte("passphrasewhichneedstobe32bytes!")
	nonce := []byte("unique_nonce")
	clearText := []byte("...encrypted_data...")

	vaultMasterKey, err := memvault.NewWithSecret(masterKey)
	require.NoError(t, err)

	defer vaultMasterKey.Wipe()

	block, err := aes.NewCipher(vaultMasterKey.Bytes())
	require.NoError(t, err)

	aesGCM, err := cipher.NewGCM(block)
	require.NoError(t, err)

	encryptedStore, err := memvault.NewWithCapacity(36)
	require.NoError(t, err)
	defer encryptedStore.Wipe()

	_ = aesGCM.Seal(encryptedStore.Bytes()[:0], nonce, clearText, nil)

	// this is the place we have decrypted data
	decryptedStore, err := memvault.NewWithCapacity(len(clearText))
	assert.NoError(t, err)
	defer decryptedStore.Wipe()

	_, err = aesGCM.Open(decryptedStore.Bytes()[:0], nonce, encryptedStore.Bytes(), nil)
	assert.NoError(t, err)

	assert.NoError(t, err)
	assert.Equal(t, clearText, decryptedStore.Bytes())
}

func TestExampleEncryption2(t *testing.T) {
	veryImportantSecret := []byte("passphrasewhichneedstobe32bytes!")
	nonce := []byte("unique_nonce")
	clearText := []byte("...encrypted_data...")

	vaultMasterKey, err := memvault.NewWithSecret(veryImportantSecret)
	require.NoError(t, err)

	var aesGCM cipher.AEAD
	err = vaultMasterKey.ReadAndWipe(func(data []byte) error {
		block, err := aes.NewCipher(vaultMasterKey.Bytes())
		require.NoError(t, err)

		aesGCM, err = cipher.NewGCM(block)
		require.NoError(t, err)
		return nil
	})
	require.NoError(t, err)

	encryptedStore, err := memvault.NewWithCapacity(36)
	require.NoError(t, err)
	defer encryptedStore.Wipe()

	err = encryptedStore.Read(func(data []byte) error {
		encrypted := aesGCM.Seal(data[:0], nonce, clearText, nil)
		fmt.Printf("%s\n", string(encrypted))

		return nil
	})
	assert.NoError(t, err)

	secureStore, err := memvault.NewWithCapacity(len(clearText))
	assert.NoError(t, err)

	err = secureStore.ReadAndWipe(func(secureData []byte) error {
		err = encryptedStore.ReadAndWipe(func(encryptedData []byte) error {
			_, err = aesGCM.Open(secureData[:0], nonce, encryptedData, nil)

			assert.NoError(t, err)
			assert.Equal(t, clearText, secureData)
			return err
		})
		return err
	})

	require.NoError(t, err)
}

func TestWithCapacity(t *testing.T) {
	t.Run("should return error if the input is zero", func(t *testing.T) {
		// when
		vault, err := memvault.NewWithCapacity(0)

		// then
		assert.ErrorIs(t, err, memvault.ErrInvalidInput)
		assert.Nil(t, vault)
	})

	t.Run("should return error if the input is less than zero", func(t *testing.T) {
		// when
		vault, err := memvault.NewWithCapacity(-1)

		// then
		assert.ErrorIs(t, err, memvault.ErrInvalidInput)
		assert.Nil(t, vault)
	})
}

func TestNewWithSecret(t *testing.T) {
	t.Run("should return error if the input is nil", func(t *testing.T) {
		// when
		vault, err := memvault.NewWithSecret(nil)
		// then
		assert.ErrorIs(t, err, memvault.ErrInvalidInput)
		assert.Nil(t, vault)
	})

	t.Run("should initialize without any error", func(t *testing.T) {
		// given
		input := []byte("secret")

		// when
		vault, err := memvault.NewWithSecret(input)

		// then
		assert.NoError(t, err)
		assert.NotNil(t, vault)
	})

	t.Run("should clear input secret to zeros", func(t *testing.T) {
		// given
		input := []byte("secret")

		// when
		_, err := memvault.NewWithSecret(input)

		// then
		assert.NoError(t, err)
		assert.Equal(t, []byte{0, 0, 0, 0, 0, 0}, input) // original input should be cleared
	})
}

func TestRead(t *testing.T) {
	t.Run("should able to read multiple times from the vault", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.NewWithSecret(input)
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
		vault, err := memvault.NewWithSecret(input)
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
		vault, err := memvault.NewWithSecret(input)
		assert.NoError(t, err)

		// when
		err = vault.Wipe()

		// then
		assert.NoError(t, err)
	})

	t.Run("should be able to wipe secrets multiple times", func(t *testing.T) {
		// given
		input := []byte("secret")
		vault, err := memvault.NewWithSecret(input)
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
		vault, err := memvault.NewWithSecret(input)
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
		vault, err := memvault.NewWithSecret(input)
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
		vault, err := memvault.NewWithSecret(input)
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
		vault, err := memvault.NewWithSecret(input)
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
		vault, err := memvault.NewWithSecret(input)
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

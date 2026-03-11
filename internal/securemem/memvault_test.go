package securemem_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openkcm/krypton/internal/securemem"
)

func TestNewVault(t *testing.T) {
	// given
	// when
	vault := securemem.NewMemVault()

	// then
	assert.NotNil(t, vault)
}

func TestGet(t *testing.T) {
	t.Run("should return data from vault", func(t *testing.T) {
		// given
		vault := securemem.NewMemVault()
		name := "test"
		data := []byte("secret")

		err := vault.Put(name, data)
		assert.NoError(t, err)

		// when
		actResult, ok := vault.Get(name)

		// then
		assert.True(t, ok)
		assert.Equal(t, data, actResult)
	})

	t.Run("should return false when data does not exist in vault", func(t *testing.T) {
		// given
		vault := securemem.NewMemVault()
		name := "non-existing"

		// when
		actResult, ok := vault.Get(name)

		// then
		assert.False(t, ok)
		assert.Nil(t, actResult)
	})
}

func TestPut(t *testing.T) {
	t.Run("should put data into vault", func(t *testing.T) {
		// given
		vault := securemem.NewMemVault()
		keys := []string{"test1", "test2", "test3"}
		data := []byte("secret")

		// when
		for _, name := range keys {
			err := vault.Put(name, data)
			// then
			assert.NoError(t, err)
		}

		for _, name := range keys {
			// then
			actResult, ok := vault.Get(name)
			assert.True(t, ok)
			assert.Equal(t, data, actResult)
		}
	})

	t.Run("should not change the original data after putting into vault", func(t *testing.T) {
		// given
		vault := securemem.NewMemVault()
		name := "test"
		data := []byte("secret")

		// when
		err := vault.Put(name, data)

		// then
		assert.NoError(t, err)

		actResult, ok := vault.Get(name)
		assert.True(t, ok)
		assert.Equal(t, data, actResult)

		data[0] = 'S'

		actResult, ok = vault.Get(name)
		assert.True(t, ok)
		assert.Equal(t, []byte("secret"), actResult)
	})

	t.Run("should return error when data is", func(t *testing.T) {
		tts := []struct {
			name string
			data []byte
		}{
			{name: "empty", data: []byte("")},
			{name: "nil", data: nil},
		}

		for _, tt := range tts {
			t.Run(tt.name, func(t *testing.T) {
				// given
				vault := securemem.NewMemVault()
				name := "foo"

				// when
				err := vault.Put(name, tt.data)

				// then
				assert.ErrorIs(t, err, securemem.ErrInvalidSize)
			})
		}
	})

	t.Run("should be able to put data with same name", func(t *testing.T) {
		// given
		vault := securemem.NewMemVault()
		name := "test"
		data1 := []byte("secret1")
		data2 := []byte("secret2")

		// when
		err := vault.Put(name, data1)

		// then
		assert.NoError(t, err)

		actResult, ok := vault.Get(name)
		assert.True(t, ok)
		assert.Equal(t, data1, actResult)

		// when
		err = vault.Put(name, data2)

		// then
		assert.NoError(t, err)

		actResult, ok = vault.Get(name)
		assert.True(t, ok)
		assert.Equal(t, data2, actResult)
	})
}

func TestReserve(t *testing.T) {
	t.Run("should reserve a buffer in the vault", func(t *testing.T) {
		// given
		vault := securemem.NewMemVault()
		name := "test"
		data := []byte("secret")

		// when
		actBytes, err := vault.Reserve(name, len(data))

		// then
		assert.NoError(t, err)
		assert.Len(t, actBytes, len(data))
		for i := range actBytes {
			assert.Equal(t, byte(0), actBytes[i])
		}

		copy(actBytes, data)

		actBytes, ok := vault.Get(name)
		assert.True(t, ok)
		assert.Equal(t, data, actBytes)
	})

	t.Run("should return error when reserve size is invalid", func(t *testing.T) {
		tts := []struct {
			name string
			size int
		}{
			{name: "size 0", size: 0},
			{name: "size negative", size: -1},
		}

		for _, tt := range tts {
			t.Run(tt.name, func(t *testing.T) {
				// given
				vault := securemem.NewMemVault()
				name := "foo"

				// when
				actBytes, err := vault.Reserve(name, tt.size)

				// then
				assert.ErrorIs(t, err, securemem.ErrInvalidSize)
				assert.Nil(t, actBytes)
			})
		}
	})

	t.Run("should return an error if we reserve data with same name", func(t *testing.T) {
		// given
		vault := securemem.NewMemVault()
		name := "test"
		data := []byte("secret1")

		// when
		actBytes, err := vault.Reserve(name, len(data))

		// then
		assert.NoError(t, err)
		assert.Len(t, actBytes, len(data))

		// when
		actBytes, err = vault.Reserve(name, len(data))

		// then
		assert.Error(t, err)
		assert.Nil(t, actBytes)
	})
}

func TestPutAndReserve(t *testing.T) {
	t.Run("should reserve and put data into vault with same name", func(t *testing.T) {
		// given
		vault := securemem.NewMemVault()
		name := "test"
		data := []byte("secret")

		// when
		actBytes, err := vault.Reserve(name, 12)

		// then
		assert.NoError(t, err)
		assert.Len(t, actBytes, 12)

		actBytes, ok := vault.Get(name)
		assert.True(t, ok)
		for i := range actBytes {
			assert.Equal(t, byte(0), actBytes[i])
		}

		// when
		err = vault.Put(name, data)

		// then
		assert.NoError(t, err)

		actResult, ok := vault.Get(name)
		assert.True(t, ok)
		assert.Equal(t, data, actResult)
	})

	t.Run("should put and reserve data into vault with same name", func(t *testing.T) {
		// given
		vault := securemem.NewMemVault()
		name := "test"
		data := []byte("secret")

		// when
		err := vault.Put(name, data)
		// then
		assert.NoError(t, err)

		actResult, ok := vault.Get(name)
		assert.True(t, ok)
		assert.Equal(t, data, actResult)

		// when
		actBytes, err := vault.Reserve(name, 12)

		// then
		assert.NoError(t, err)
		assert.Len(t, actBytes, 12)

		actBytes, ok = vault.Get(name)
		assert.True(t, ok)
		for i := range actBytes {
			assert.Equal(t, byte(0), actBytes[i])
		}
	})
}

func TestVaultDestroy(t *testing.T) {
	t.Run("destroy", func(t *testing.T) {
		t.Run("should destroy a specific data in vault", func(t *testing.T) {
			// given
			vault := securemem.NewMemVault()
			name1 := "test1"
			name2 := "test2"
			name3 := "test3"
			data1 := []byte("secret1")
			data2 := []byte("secret2")
			data3 := []byte("secret3")

			err := vault.Put(name1, data1)
			assert.NoError(t, err)

			err = vault.Put(name2, data2)
			assert.NoError(t, err)

			_, err = vault.Reserve(name3, len(data3))
			assert.NoError(t, err)

			// when
			err = vault.Destroy(name1)
			assert.NoError(t, err)

			err = vault.Destroy(name3)
			assert.NoError(t, err)

			// then
			actBytes, ok := vault.Get(name1)
			assert.False(t, ok)
			assert.Nil(t, actBytes)

			actBytes, ok = vault.Get(name2)
			assert.True(t, ok)
			assert.Equal(t, data2, actBytes)

			actBytes, ok = vault.Get(name3)
			assert.False(t, ok)
			assert.Nil(t, actBytes)
		})

		t.Run("should be idempotent when destroying data", func(t *testing.T) {
			// given
			vault := securemem.NewMemVault()
			name := "test"
			data := []byte("secret")

			err := vault.Put(name, data)
			assert.NoError(t, err)

			// when
			err = vault.Destroy(name)
			assert.NoError(t, err)

			err = vault.Destroy(name)
			assert.NoError(t, err)

			// then
			actBytes, ok := vault.Get(name)
			assert.False(t, ok)
			assert.Nil(t, actBytes)
		})

		t.Run("should not return an error when destroying non-existing data", func(t *testing.T) {
			// given
			vault := securemem.NewMemVault()
			name := "test"

			// when
			err := vault.Destroy(name)

			// then
			assert.NoError(t, err)
		})

		t.Run("should be able to reuse the name after destroying data", func(t *testing.T) {
			// given
			vault := securemem.NewMemVault()
			name := "test"
			data1 := []byte("secret1")
			data2 := []byte("secret2")

			err := vault.Put(name, data1)
			assert.NoError(t, err)

			// when
			err = vault.Destroy(name)
			assert.NoError(t, err)

			err = vault.Put(name, data2)
			assert.NoError(t, err)

			// then
			actBytes, ok := vault.Get(name)
			assert.True(t, ok)
			assert.Equal(t, data2, actBytes)
		})
	})

	t.Run("destroy all", func(t *testing.T) {
		t.Run("should destroy all data in vault", func(t *testing.T) {
			// given
			vault := securemem.NewMemVault()
			name1 := "test1"
			name2 := "test2"
			name3 := "test3"
			data1 := []byte("secret1")
			data2 := []byte("secret2")

			err := vault.Put(name1, data1)
			assert.NoError(t, err)

			err = vault.Put(name2, data2)
			assert.NoError(t, err)

			_, err = vault.Reserve(name3, len(data1))
			assert.NoError(t, err)

			// when
			err = vault.DestroyAll()
			assert.NoError(t, err)

			// then
			actBytes, ok := vault.Get(name1)
			assert.False(t, ok)
			assert.Nil(t, actBytes)

			actBytes, ok = vault.Get(name2)
			assert.False(t, ok)
			assert.Nil(t, actBytes)

			actBytes, ok = vault.Get(name3)
			assert.False(t, ok)
			assert.Nil(t, actBytes)
		})

		t.Run("should be idempotent when destroying all data", func(t *testing.T) {
			// given
			vault := securemem.NewMemVault()
			name := "test"
			data := []byte("secret")

			err := vault.Put(name, data)
			assert.NoError(t, err)

			// when
			err = vault.DestroyAll()
			assert.NoError(t, err)

			err = vault.DestroyAll()
			assert.NoError(t, err)

			// then
			actBytes, ok := vault.Get(name)
			assert.False(t, ok)
			assert.Nil(t, actBytes)
		})

		t.Run("should not return an error when vault is empty", func(t *testing.T) {
			// given
			vault := securemem.NewMemVault()

			// when
			err := vault.DestroyAll()

			// then
			assert.NoError(t, err)
		})

		t.Run("should be able to reuse the names after destroying all data", func(t *testing.T) {
			// given
			vault := securemem.NewMemVault()
			name1 := "test1"
			name2 := "test2"
			data1 := []byte("secret1")
			data2 := []byte("secret2")

			err := vault.Put(name1, data1)
			assert.NoError(t, err)

			err = vault.Put(name2, data2)
			assert.NoError(t, err)

			// when
			err = vault.DestroyAll()
			assert.NoError(t, err)

			err = vault.Put(name1, data1)
			assert.NoError(t, err)

			err = vault.Put(name2, data2)
			assert.NoError(t, err)

			// then
			actBytes, ok := vault.Get(name1)
			assert.True(t, ok)
			assert.Equal(t, data1, actBytes)

			actBytes, ok = vault.Get(name2)
			assert.True(t, ok)
			assert.Equal(t, data2, actBytes)
		})
	})
}

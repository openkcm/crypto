package securemem_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openkcm/krypton/internal/securemem"
)

func TestNewWithSize(t *testing.T) {
	t.Run("should create vault with specified size", func(t *testing.T) {
		// given when
		subj, err := securemem.NewVaultData("test-region", 64)
		assert.NoError(t, err)

		t.Cleanup(func() {
			err := subj.Destroy()
			assert.NoError(t, err)
		})

		// then
		data := subj.Data()
		assert.Len(t, data, 64)

		// mmap'd anonymous memory should be zeroed
		for _, b := range data {
			assert.Equal(t, byte(0), b)
		}
	})

	t.Run("should return error for invalid sizes", func(t *testing.T) {
		tts := []struct {
			name string
			size int
		}{
			{name: "negative size", size: -1},
			{name: "zero size", size: 0},
		}

		for _, tt := range tts {
			t.Run("for "+tt.name, func(t *testing.T) {
				// given when
				subj, err := securemem.NewVaultData("test-"+tt.name, tt.size)

				// then
				assert.ErrorIs(t, err, securemem.ErrInvalidSize)
				assert.Nil(t, subj)
			})
		}
	})
}

func TestNewWithData(t *testing.T) {
	t.Run("should create vault with provided data", func(t *testing.T) {
		// given
		input := []byte("secret123")

		// when
		subj, err := securemem.NewVaultDataFrom("test-secret", input)
		assert.NoError(t, err)

		t.Cleanup(func() {
			err := subj.Destroy()
			assert.NoError(t, err)
		})

		// then
		assert.Equal(t, input, subj.Data(), "vault data does not match input")
	})

	t.Run("should return error for invalid data inputs", func(t *testing.T) {
		tts := []struct {
			name string
			data []byte
		}{
			{name: "empty input", data: []byte{}},
			{name: "nil input", data: nil},
		}

		for _, tt := range tts {
			t.Run("should return error for "+tt.name, func(t *testing.T) {
				// given when
				subj, err := securemem.NewVaultDataFrom("test-"+tt.name, tt.data)

				// then
				assert.ErrorIs(t, err, securemem.ErrInvalidSize)
				assert.Nil(t, subj)
			})
		}
	})

	t.Run("should create independent copy of input data", func(t *testing.T) {
		// given
		input := []byte("original")

		// when
		subj, err := securemem.NewVaultDataFrom("test-copy", input)
		assert.NoError(t, err)

		t.Cleanup(func() {
			err := subj.Destroy()
			assert.NoError(t, err)
		})

		// then
		// Mutate the original slice
		input[0] = 'X'

		assert.Equal(t, byte('o'), subj.Data()[0])
	})
}

func TestDestroy(t *testing.T) {
	t.Run("should securely destroy vault data", func(t *testing.T) {
		// given
		subj, err := securemem.NewVaultData("test-destroy", 128)
		assert.NoError(t, err)

		// when
		err = subj.Destroy()

		// then
		assert.NoError(t, err)
		assert.Nil(t, subj.Data())
	})

	t.Run("should be idempotent", func(t *testing.T) {
		// given
		subj, err := securemem.NewVaultData("test-idempotent", 64)
		assert.NoError(t, err)

		// when
		err = subj.Destroy()

		// then
		assert.NoError(t, err)

		// when
		err = subj.Destroy()

		// then
		assert.NoError(t, err)
		assert.Nil(t, subj.Data())
	})
}

func TestReadonly(t *testing.T) {
	t.Run("should set vault to readonly mode", func(t *testing.T) {
		// given
		subj, err := securemem.NewVaultData("test-readonly", 40)
		assert.NoError(t, err)

		t.Cleanup(func() {
			err := subj.Destroy()
			assert.NoError(t, err)
		})

		// when
		err = subj.Readonly()

		// then
		assert.NoError(t, err)
	})

	t.Run("should be idempotent", func(t *testing.T) {
		// given
		subj, err := securemem.NewVaultData("test-readonly-idempotent", 40)
		assert.NoError(t, err)

		t.Cleanup(func() {
			err := subj.Destroy()
			assert.NoError(t, err)
		})

		// when
		err = subj.Readonly()

		// then
		assert.NoError(t, err)

		// when
		err = subj.Readonly()

		// then
		assert.NoError(t, err)
	})
}

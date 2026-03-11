package securemem_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/openkcm/krypton/internal/securemem"
)

func TestHandler(t *testing.T) {
	t.Run("Persist", func(t *testing.T) {
		t.Run("should store the value", func(t *testing.T) {
			// given
			exp := []byte("hello world")
			subj := securemem.NewHandlerRequest()

			// when
			err := subj.Persist("foo", exp)

			// then
			assert.NoError(t, err)

			actResult, ok := subj.Get("foo")
			assert.True(t, ok)
			assert.Equal(t, exp, actResult)
		})

		t.Run("should return an error when data is", func(t *testing.T) {
			tts := []struct {
				name string
				data []byte
			}{
				{name: "empty", data: []byte{}},
				{name: "nil", data: nil},
			}

			for _, tt := range tts {
				t.Run(tt.name, func(t *testing.T) {
					// given
					subj := securemem.NewHandlerRequest()

					// when
					err := subj.Persist("foo", tt.data)

					// then
					assert.Error(t, err)
				})
			}
		})

		t.Run("should overwrite existing value", func(t *testing.T) {
			// given
			subj := securemem.NewHandlerRequest()
			err := subj.Persist("foo", []byte("hello world"))
			assert.NoError(t, err)

			// when
			err = subj.Persist("foo", []byte("hello new world"))

			// then
			assert.NoError(t, err)

			actResult, ok := subj.Get("foo")
			assert.True(t, ok)
			assert.Equal(t, []byte("hello new world"), actResult)
		})
	})

	t.Run("Put", func(t *testing.T) {
		t.Run("should store the value", func(t *testing.T) {
			// given
			exp := []byte("hello world")
			subj := securemem.NewHandlerRequest()

			// when
			err := subj.Put("foo", exp)

			// then
			assert.NoError(t, err)

			actResult, ok := subj.Get("foo")
			assert.True(t, ok)
			assert.Equal(t, exp, actResult)
		})

		t.Run("should return an error when data is", func(t *testing.T) {
			tts := []struct {
				name string
				data []byte
			}{
				{name: "empty", data: []byte{}},
				{name: "nil", data: nil},
			}

			for _, tt := range tts {
				t.Run(tt.name, func(t *testing.T) {
					// given
					subj := securemem.NewHandlerRequest()

					// when
					err := subj.Put("foo", tt.data)

					// then
					assert.Error(t, err)
				})
			}
		})
	})

	t.Run("Get", func(t *testing.T) {
		t.Run("should return false if the value does not exist", func(t *testing.T) {
			// given
			subj := securemem.NewHandlerRequest()

			// when
			b, ok := subj.Get("foo")

			// then
			assert.False(t, ok)
			assert.Nil(t, b)
		})

		t.Run("should return the value if it exists", func(t *testing.T) {
			// given
			exp := []byte("hello world")
			subj := securemem.NewHandlerRequest()
			err := subj.Put("foo", exp)
			assert.NoError(t, err)

			// when
			actResult, ok := subj.Get("foo")

			// then
			assert.True(t, ok)
			assert.Equal(t, exp, actResult)
		})
	})

	t.Run("Reserve", func(t *testing.T) {
		t.Run("should return a byte slice of the specified size", func(t *testing.T) {
			// given
			subj := securemem.NewHandlerRequest()

			// when
			b, err := subj.Reserve("foo", 10)

			// then
			assert.NoError(t, err)
			assert.Len(t, b, 10)
		})

		t.Run("should return an error if the size is", func(t *testing.T) {
			tts := []struct {
				name string
				size int
			}{
				{name: "zero", size: 0},
				{name: "negative", size: -1},
			}

			for _, tt := range tts {
				t.Run(tt.name, func(t *testing.T) {
					// given
					subj := securemem.NewHandlerRequest()

					// when
					b, err := subj.Reserve("foo", tt.size)

					// then
					assert.Error(t, err)
					assert.Nil(t, b)
				})
			}
		})

		t.Run("should store the entry in the vault so it can be retrieved with Get", func(t *testing.T) {
			// given
			subj := securemem.NewHandlerRequest()
			b, err := subj.Reserve("foo", 3)
			assert.NoError(t, err)
			assert.Len(t, b, 3)

			// copy the data to the reserved byte slice
			copy(b, []byte("bar"))

			// when
			actResult, ok := subj.Get("foo")

			// then
			assert.True(t, ok)
			assert.Equal(t, []byte("bar"), actResult)
		})
	})

	t.Run("Destroy", func(t *testing.T) {
		t.Run("should remove the entry from the vault", func(t *testing.T) {
			// given
			subj := securemem.NewHandlerRequest()
			err := subj.Persist("foo", []byte("hello world"))
			assert.NoError(t, err)

			// when
			err = subj.Destroy("foo")

			// then
			assert.NoError(t, err)

			actResult, ok := subj.Get("foo")
			assert.False(t, ok)
			assert.Nil(t, actResult)
		})

		t.Run("should not return an error if the entry does not exist", func(t *testing.T) {
			// given
			subj := securemem.NewHandlerRequest()

			// when
			err := subj.Destroy("foo")

			// then
			assert.NoError(t, err)
		})

		t.Run("should not affect other entries in the vault", func(t *testing.T) {
			// given
			subj := securemem.NewHandlerRequest()
			err := subj.Persist("foo", []byte("hello world"))
			assert.NoError(t, err)

			err = subj.Put("baz", []byte("hello world"))
			assert.NoError(t, err)

			// when
			err = subj.Destroy("foo")

			// then
			assert.NoError(t, err)

			actResult, ok := subj.Get("foo")
			assert.False(t, ok)
			assert.Nil(t, actResult)

			actResult, ok = subj.Get("baz")
			assert.True(t, ok)
			assert.Equal(t, []byte("hello world"), actResult)
		})
	})

	t.Run("DestroyAll should clear all entries in the vault", func(t *testing.T) {
		// given
		subj := securemem.NewHandlerRequest()
		key1 := "foo"
		key2 := "baz"
		key3 := "qux"
		err := subj.Persist(key1, []byte("hello world"))
		assert.NoError(t, err)

		err = subj.Put(key2, []byte("hello world"))
		assert.NoError(t, err)

		_, err = subj.Reserve(key3, 10)
		assert.NoError(t, err)

		// when
		err = subj.DestroyAll()

		// then
		assert.NoError(t, err)

		actResult, ok := subj.Get(key1)
		assert.False(t, ok)
		assert.Nil(t, actResult)

		actResult, ok = subj.Get(key2)
		assert.False(t, ok)
		assert.Nil(t, actResult)

		actResult, ok = subj.Get(key3)
		assert.False(t, ok)
		assert.Nil(t, actResult)
	})
}

func TestRun(t *testing.T) {
	t.Run("should return the persisted data in the handler response", func(t *testing.T) {
		// given
		exp := []byte("hello world")
		keys := []string{"foo", "bar", "baz"}

		// when
		resp, err := securemem.Run(context.Background(), func(ctx context.Context, req *securemem.HandlerRequest) error {
			for _, key := range keys {
				err := req.Persist(key, exp)
				assert.NoError(t, err)
			}
			return nil
		})

		// then
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		for _, key := range keys {
			actResult, ok := resp.Get(key)
			assert.True(t, ok)
			assert.Equal(t, exp, actResult)
		}
	})

	t.Run("should only return the persisted data in the handler response", func(t *testing.T) {
		// given
		key1 := "foo"
		key2 := "bar"
		key3 := "baz"
		data1 := []byte("data1")
		data2 := []byte("data2")

		// when
		resp, err := securemem.Run(context.Background(), func(ctx context.Context, req *securemem.HandlerRequest) error {
			err := req.Put(key1, data1)
			assert.NoError(t, err)

			err = req.Persist(key2, data2)
			assert.NoError(t, err)

			_, err = req.Reserve(key3, 10)
			assert.NoError(t, err)
			return nil
		})

		// then
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		actResult, ok := resp.Get(key2)
		assert.True(t, ok)
		assert.Equal(t, data2, actResult)

		actResult, ok = resp.Get(key1)
		assert.False(t, ok)
		assert.Nil(t, actResult)

		actResult, ok = resp.Get(key3)
		assert.False(t, ok)
		assert.Nil(t, actResult)
	})

	t.Run("should not return data Put in the handler request", func(t *testing.T) {
		// given
		exp := []byte("hello world")
		keys := []string{"foo", "bar", "baz"}

		// when
		resp, err := securemem.Run(context.Background(), func(ctx context.Context, req *securemem.HandlerRequest) error {
			for _, key := range keys {
				err := req.Put(key, exp)
				assert.NoError(t, err)
			}
			return nil
		})

		// then
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		for _, key := range keys {
			actResult, ok := resp.Get(key)
			assert.False(t, ok)
			assert.Nil(t, actResult)
		}
	})
	t.Run("should not return data reserved in the handler request", func(t *testing.T) {
		// given
		keys := []string{"foo", "bar", "baz"}

		// when
		resp, err := securemem.Run(context.Background(), func(ctx context.Context, req *securemem.HandlerRequest) error {
			for _, key := range keys {
				_, err := req.Reserve(key, 10)
				assert.NoError(t, err)
			}
			return nil
		})

		// then
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		for _, key := range keys {
			actResult, ok := resp.Get(key)
			assert.False(t, ok)
			assert.Nil(t, actResult)
		}
	})

	t.Run("should Destroy all vault entries in the internal handler request", func(t *testing.T) {
		// given
		var actualReq *securemem.HandlerRequest
		exp := []byte("hello world")
		keys := []string{"foo", "bar", "baz"}

		// when
		resp, err := securemem.Run(context.Background(), func(ctx context.Context, req *securemem.HandlerRequest) error {
			for _, key := range keys {
				err := req.Persist(key, exp)
				assert.NoError(t, err)
			}
			actualReq = req
			return nil
		})

		// then
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		for _, key := range keys {
			actResult, ok := actualReq.Get(key)
			assert.False(t, ok)
			assert.Nil(t, actResult)
		}
	})

	t.Run("should return an error if the handler returns an error", func(t *testing.T) {
		// given

		// when
		resp, err := securemem.Run(context.Background(), func(ctx context.Context, req *securemem.HandlerRequest) error {
			return assert.AnError
		})

		// then
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, resp)
	})

	t.Run("should not persist the keys in the handler response if the handler returns an error", func(t *testing.T) {
		// given
		var actReq *securemem.HandlerRequest

		// when
		resp, err := securemem.Run(context.Background(), func(ctx context.Context, req *securemem.HandlerRequest) error {
			actReq = req

			err := req.Persist("foo", []byte("hello world"))
			assert.NoError(t, err)

			return assert.AnError
		})

		// then
		assert.ErrorIs(t, err, assert.AnError)
		assert.Nil(t, resp)

		actResult, ok := actReq.Get("foo")
		assert.False(t, ok)
		assert.Nil(t, actResult)
	})

	t.Run("should not persist the entry if it is destroyed and added again in put", func(t *testing.T) {
		// given

		// when
		resp, err := securemem.Run(context.Background(), func(ctx context.Context, req *securemem.HandlerRequest) error {
			err := req.Persist("foo", []byte("hello world"))
			assert.NoError(t, err)

			err = req.Destroy("foo")
			assert.NoError(t, err)

			err = req.Put("foo", []byte("hello world"))
			assert.NoError(t, err)

			return nil
		})

		// then
		assert.NoError(t, err)
		assert.NotNil(t, resp)

		actResult, ok := resp.Get("foo")
		assert.False(t, ok)
		assert.Nil(t, actResult)
	})
}

func TestTransferPersistedValues(t *testing.T) {
	t.Run("should transfer persisted values from the handler request to the handler response", func(t *testing.T) {
		// given
		req := securemem.NewHandlerRequest()
		resp := securemem.NewHandlerRequest()

		keys := []string{"foo", "bar", "baz"}

		for _, key := range keys {
			err := req.Persist(key, []byte("hello world"))
			assert.NoError(t, err)
		}

		// when
		err := securemem.TransferPersistedValues(req, resp)

		// then
		assert.NoError(t, err)

		for _, key := range keys {
			actResult, ok := resp.Get(key)
			assert.True(t, ok)
			assert.Equal(t, []byte("hello world"), actResult)
		}
	})

	t.Run("should not transfer Put values from the handler request to the handler response", func(t *testing.T) {
		// given
		req := securemem.NewHandlerRequest()
		resp := securemem.NewHandlerRequest()

		keys := []string{"foo", "bar", "baz"}

		for _, key := range keys {
			err := req.Put(key, []byte("hello world"))
			assert.NoError(t, err)
		}

		// when
		err := securemem.TransferPersistedValues(req, resp)

		// then
		assert.NoError(t, err)

		for _, key := range keys {
			actResult, ok := resp.Get(key)
			assert.False(t, ok)
			assert.Nil(t, actResult)
		}
	})

	t.Run("should not persist keys that are destroyed in the handler request", func(t *testing.T) {
		// given
		req := securemem.NewHandlerRequest()
		resp := securemem.NewHandlerRequest()

		err := req.Persist("foo", []byte("hello world"))
		assert.NoError(t, err)

		err = req.Destroy("foo")
		assert.NoError(t, err)

		// when
		err = securemem.TransferPersistedValues(req, resp)

		// then
		assert.NoError(t, err)

		actResult, ok := resp.Get("foo")
		assert.False(t, ok)
		assert.Nil(t, actResult)
	})
}

func TestContextCancel(t *testing.T) {
	t.Run("should return an error if the context is canceled before running the handler", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		// when
		resp, err := securemem.Run(ctx, func(ctx context.Context, req *securemem.HandlerRequest) error {
			return nil
		})

		// then
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, resp)
	})

	t.Run("should return an error if the context is canceled during the handler execution", func(t *testing.T) {
		// given
		ctx, cancel := context.WithCancel(context.Background())

		// when
		resp, err := securemem.Run(ctx, func(ctx context.Context, req *securemem.HandlerRequest) error {
			cancel()
			return nil
		})

		// then
		assert.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, resp)
	})
}

// TODO: This benchmark is not very meaningful, as it only tests the overhead of creating a new handler request and running
// a simple operation. A more realistic benchmark would involve multiple operations and possibly concurrent access to the vault.
func BenchmarkNewVaultData(b *testing.B) {
	for b.Loop() {
		resp, err := securemem.Run(b.Context(), func(ctx context.Context, req *securemem.HandlerRequest) error {
			secret := []byte("hello world")
			err := req.Persist("foo", secret)
			if err != nil {
				return err
			}

			resBytes, err := req.Reserve("bar", 1024)
			if err != nil {
				return err
			}
			copy(resBytes, secret)
			return nil
		})
		assert.NoError(b, err)
		assert.NotNil(b, resp)

		err = resp.DestroyAll()
		assert.NoError(b, err)
	}
}

package server

import (
	"bytes"
	"context"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/internal/kmip"
)

var KMIPMessagesHandler = func(config *config.Config) kmip.Handler {
	processor := kmip.NewProcessor()

	// the handler should be stateless
	return func(ctx context.Context, data []byte) ([]byte, error) {
		m := map[string]any{}
		err := processor.Decode(data, &m)
		if err != nil {
			return nil, err
		}

		//nolint: forcetypeassert
		buf := defaultPool.Get().(*bytes.Buffer)
		defer func() {
			buf.Reset()
			defaultPool.Put(buf)
		}()

		err = processor.Encode(m, buf)
		if err != nil {
			return nil, err
		}

		return buf.Bytes(), nil
	}
}

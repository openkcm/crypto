package payloads_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	_ "unsafe"

	"github.com/openkcm/crypto/kmip"
	"github.com/openkcm/crypto/kmip/payloads"
	"github.com/openkcm/crypto/kmip/ttlv"
)

func TestPayloadsTypes(t *testing.T) {
	for op := range ttlv.EnumValues[kmip.Operation]() {
		assert.Equal(t, op, newRequestPayload(op).Operation())
		assert.Equal(t, op, newResponsePayload(op).Operation())
	}
}

//go:linkname newRequestPayload github.com/openkcm/crypto/kmip.newRequestPayload
func newRequestPayload(op kmip.Operation) kmip.OperationPayload

//go:linkname newResponsePayload github.com/openkcm/crypto/kmip.newResponsePayload
func newResponsePayload(op kmip.Operation) kmip.OperationPayload

func TestRegisterRequestPayload_Encode_Decode(t *testing.T) {
	secret := []byte("foobar")
	req := &payloads.RegisterRequestPayload{
		ObjectType:        kmip.ObjectTypeSecretData,
		TemplateAttribute: kmip.TemplateAttribute{},
		Object: &kmip.SecretData{
			SecretDataType: kmip.SecretDataTypePassword,
			KeyBlock:       kmip.KeyBlock{KeyFormatType: kmip.KeyFormatTypeRaw, KeyValue: &kmip.KeyValue{Plain: &kmip.PlainKeyValue{KeyMaterial: kmip.KeyMaterial{Bytes: &secret}}}},
		},
	}

	enc := ttlv.NewTTLVEncoder()
	enc.TagAny(kmip.TagRequestPayload, req)
	ttlvReq := enc.Bytes()
	decodedReq := &payloads.RegisterRequestPayload{}
	dec, err := ttlv.NewTTLVDecoder(ttlvReq)
	require.NoError(t, err)
	err = dec.TagAny(kmip.TagRequestPayload, decodedReq)
	require.NoError(t, err)
	require.Equal(t, req, decodedReq)
}

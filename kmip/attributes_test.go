package kmip

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/openkcm/crypto/kmip/ttlv"
)

func TestAttribute_EncodeDecode(t *testing.T) {
	attr := Attribute{
		AttributeName: AttributeNameName,
		AttributeValue: Name{
			NameType:  NameTypeUninterpretedTextString,
			NameValue: "foobar",
		},
	}
	bytes := ttlv.MarshalTTLV(&attr)

	newAttr := Attribute{}
	err := ttlv.UnmarshalTTLV(bytes, &newAttr)
	require.NoError(t, err)
	require.Equal(t, attr, newAttr)

	index := int32(12)
	attr.AttributeIndex = &index
	bytes = ttlv.MarshalTTLV(&attr)

	newAttr = Attribute{}
	err = ttlv.UnmarshalTTLV(bytes, &newAttr)
	require.NoError(t, err)
	require.Equal(t, attr, newAttr)
}

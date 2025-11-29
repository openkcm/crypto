package kmip

import (
	"bytes"
	"encoding/hex"

	"github.com/gemalto/kmip-go/ttlv"
)

type Processor struct {
}

func NewProcessor() *Processor {
	return &Processor{}
}

func (k *Processor) Decode(data []byte, val any) error {
	decodeData, err := hex.DecodeString(string(data))
	if err != nil {
		return err
	}

	decoder := ttlv.NewDecoder(bytes.NewBuffer(decodeData))
	return decoder.Decode(val)
}

func (k *Processor) Encode(val any, buf *bytes.Buffer) error {
	return ttlv.NewEncoder(buf).Encode(val)
}

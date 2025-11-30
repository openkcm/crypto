package server

import (
	"bytes"
	"context"

	"github.com/openkcm/crypto/internal/config"
)

func process(ctx context.Context, _ *config.Config, payload []byte, out *bytes.Buffer) error {
	//req := &kmip.RequestMessage{}
	//dec, err := ttlv.NewTTLVDecoder(payload)
	//if err != nil {
	//	return fmt.Errorf("failed to create decoder: %w", err)
	//}
	//
	//
	//if err := dec.Decode(req); err != nil {
	//	return fmt.Errorf("KMIP decode failed: %w", err)
	//}
	//
	//responsePayload := &kmip.ResponseMessage{
	//	Header: kmip.ResponseHeader{
	//		ProtocolVersion: kmip.ProtocolVersion{
	//			ProtocolVersionMajor: 2,
	//			ProtocolVersionMinor: 0,
	//		},
	//		TimeStamp:              time.Now().UTC(),
	//		ClientCorrelationValue: "<unknown to figure out>",
	//		ServerCorrelationValue: "<unknown to figure out>",
	//	},
	//	BatchItem: make([]kmip.ResponseBatchItem, 0, len(req.BatchItem)),
	//}
	//
	//for _, batch := range req.BatchItem {
	//	res := kmip.ResponseBatchItem{
	//		Operation:         batch.Operation,
	//		UniqueBatchItemID: batch.UniqueBatchItemID,
	//		ResultStatus:      kmip.ResultStatusSuccess,
	//	}
	//
	//	op, err := operations.Lookup(batch.Operation)
	//	if err != nil {
	//		msg := fmt.Sprintf("unsupported operation %v", batch.Operation)
	//		res.ResultStatus = kmip.ResultStatusOperationFailed
	//		res.ResultMessage = msg
	//	}
	//	respPayload, err := op.Execute(ctx, batch.RequestPayload)
	//	if err != nil {
	//		msg := fmt.Sprintf("failed to execute operation %v: %v", batch.Operation, err)
	//		res.ResultStatus = kmip.ResultStatusOperationFailed
	//		res.ResultMessage = msg
	//	}
	//	res.ResponsePayload = respPayload
	//
	//	responsePayload.BatchItem = append(responsePayload.BatchItem, res)
	//}
	//responsePayload.ResponseHeader.BatchCount = len(responsePayload.BatchItem)
	//
	//enc := ttlv.NewTTLVEncoder().
	//
	//if err := enc.Encode(responsePayload); err != nil {
	//	return fmt.Errorf("KMIP encode failed: %w", err)
	//}

	return nil
}

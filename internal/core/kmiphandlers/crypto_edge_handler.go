package kmiphandlers

import (
	"context"

	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/internal/core"
	"github.com/openkcm/crypto/internal/core/operations"
	"github.com/openkcm/crypto/kmip"
)

type CryptoEdgeHandler struct {
	svcRegistry core.ServiceRegistry
	registry    operations.OperationReadRegistry
}

func NewCryptoEdgeHandler(registry operations.OperationReadRegistry, cfg *config.Config) (*CryptoEdgeHandler, error) {
	return &CryptoEdgeHandler{
		svcRegistry: core.NewServiceRegistry(cfg),

		registry: registry,
	}, nil
}

func (h *CryptoEdgeHandler) HandleRequest(ctx context.Context, req *kmip.RequestMessage) *kmip.ResponseMessage {
	responseItems := []kmip.ResponseBatchItem{}
	for _, item := range req.BatchItem {
		respItem := kmip.ResponseBatchItem{
			Operation:         item.Operation,
			UniqueBatchItemID: item.UniqueBatchItemID,
			ResultStatus:      kmip.ResultStatusSuccess,
		}

		op := h.registry.Lookup(item.Operation)
		if op == nil {
			respItem.ResultStatus = kmip.ResultStatusOperationFailed
			respItem.ResultReason = kmip.ResultReasonOperationNotSupported
			responseItems = append(responseItems, respItem)
			continue
		}
		result, err := op.Execute(ctx, h.svcRegistry)
		if err != nil {
			respItem.ResultStatus = kmip.ResultStatusOperationFailed
			respItem.ResultReason = kmip.ResultReasonIllegalOperation
			respItem.ResultMessage = err.Error()
			responseItems = append(responseItems, respItem)
			continue
		}
		respItem.ResponsePayload = result

		responseItems = append(responseItems, respItem)
	}
	// Process KMIP request and return response
	// Implement your key management logic here
	return &kmip.ResponseMessage{
		Header: kmip.ResponseHeader{
			ProtocolVersion: req.Header.ProtocolVersion,
			BatchCount:      req.Header.BatchCount,
		},
		BatchItem: responseItems,
	}
}

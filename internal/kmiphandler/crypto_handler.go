package kmiphandler

import (
	"context"

	"github.com/openkcm/crypto/internal/actions"
	"github.com/openkcm/crypto/internal/config"
	"github.com/openkcm/crypto/kmip"
)

type CryptoHandler struct {
	config   *config.Config
	registry actions.ReadRegistry
}

func NewCryptoHandler(registry actions.ReadRegistry, config *config.Config) (*CryptoHandler, error) {
	return &CryptoHandler{
		config: config,

		registry: registry,
	}, nil
}

func (h *CryptoHandler) HandleRequest(ctx context.Context, req *kmip.RequestMessage) *kmip.ResponseMessage {
	responseItems := []kmip.ResponseBatchItem{}
	for _, item := range req.BatchItem {
		respItem := kmip.ResponseBatchItem{
			Operation:         item.Operation,
			UniqueBatchItemID: item.UniqueBatchItemID,
			ResultStatus:      kmip.ResultStatusSuccess,
		}

		action := h.registry.Lookup(item.Operation)
		if action == nil {
			respItem.ResultStatus = kmip.ResultStatusOperationFailed
			respItem.ResultReason = kmip.ResultReasonOperationNotSupported
			responseItems = append(responseItems)
			continue
		}
		result, err := action.Execute(ctx)
		if err != nil {
			respItem.ResultStatus = kmip.ResultStatusOperationFailed
			respItem.ResultReason = kmip.ResultReasonIllegalOperation
			respItem.ResultMessage = err.Error()
			responseItems = append(responseItems)
			continue
		}
		respItem.ResponsePayload = result

		responseItems = append(responseItems)
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

package kmiphandlers

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"io"
	"net/http"

	"github.com/openkcm/crypto/internal/core"
	"github.com/openkcm/crypto/internal/core/authorization"
	"github.com/openkcm/crypto/internal/core/operations"
	"github.com/openkcm/crypto/kmip"
	"github.com/openkcm/crypto/kmip/kmipserver"
	"github.com/openkcm/crypto/kmip/ttlv"
	slogctx "github.com/veqryn/slog-context"
)

const (
	HeaderHttpClientCertChain = "X-Client-Cert-Chain"
	HeaderHttpContentType     = "Content-Type"
)

type CryptoHandler struct {
	svcRegistry  core.ServiceRegistry
	registry     operations.OperationReadRegistry
	authZHandler authorization.AuthorizationHandler

	proxyHttpKMIP *HttpKMIP
}

type HttpKMIP struct {
	Endpoint string
}

func NewCryptoHandler(
	registry operations.OperationReadRegistry,
	svcRegistry core.ServiceRegistry,
	authZHandler authorization.AuthorizationHandler,
	proxyHttpKMIP *HttpKMIP,
) (*CryptoHandler, error) {
	return &CryptoHandler{
		svcRegistry:   svcRegistry,
		registry:      registry,
		authZHandler:  authZHandler,
		proxyHttpKMIP: proxyHttpKMIP,
	}, nil
}

func (h *CryptoHandler) HandleRequest(ctx context.Context, req *kmip.RequestMessage) *kmip.ResponseMessage {
	isProxyEnabledForNotSupportedOperations := h.proxyHttpKMIP != nil

	nonDelegatedOps := []kmip.Operation{}
	for _, item := range req.BatchItem {
		if isProxyEnabledForNotSupportedOperations {
			op := h.registry.Lookup(item.Operation)
			if op != nil {
				nonDelegatedOps = append(nonDelegatedOps, item.Operation)
			}
		} else {
			nonDelegatedOps = append(nonDelegatedOps, item.Operation)
		}
	}

	authorized := &authorization.CheckResponse{
		Result: true,
	}
	if !isProxyEnabledForNotSupportedOperations {
		certificates := kmipserver.PeerCertificates(ctx)
		if len(certificates) == 0 {
			certificates = extractCertificatesFromHeaders(kmipserver.RequestHeaders(ctx))
		}
		authorized = h.authZHandler(kmipserver.PeerCertificates(ctx), nonDelegatedOps).Check()
	}

	response := &kmip.ResponseMessage{
		Header: kmip.ResponseHeader{
			ProtocolVersion: req.Header.ProtocolVersion,
			BatchCount:      req.Header.BatchCount,
		},
		BatchItem: []kmip.ResponseBatchItem{},
	}

	if !authorized.Result {
		for _, item := range req.BatchItem {
			if ok, found := authorized.PerOperation[item.Operation]; found && !ok {
				response.BatchItem = append(response.BatchItem, kmip.ResponseBatchItem{
					Operation:         item.Operation,
					UniqueBatchItemID: item.UniqueBatchItemID,
					ResultStatus:      kmip.ResultStatusOperationFailed,
					ResultReason:      kmip.ResultReasonPermissionDenied,
				})
			}
		}
		return response
	}

	if isProxyEnabledForNotSupportedOperations {
		// here all operations not supported will be delegated to the configured kmip address
		newReq := &kmip.RequestMessage{
			Header:    req.Header,
			BatchItem: []kmip.RequestBatchItem{},
		}

		for _, item := range req.BatchItem {
			op := h.registry.Lookup(item.Operation)
			if op == nil {
				newReq.BatchItem = append(newReq.BatchItem, item)
			}
		}

		if len(newReq.BatchItem) > 0 {
			resp, err := h.delegateRequestToProxy(ctx, newReq)
			if err != nil {
				slogctx.Error(ctx, "error delegating request to proxy", "error", err)

				for _, item := range newReq.BatchItem {
					response.BatchItem = append(response.BatchItem, kmip.ResponseBatchItem{
						Operation:         item.Operation,
						UniqueBatchItemID: item.UniqueBatchItemID,
						ResultStatus:      kmip.ResultStatusOperationFailed,
						ResultReason:      kmip.ResultReasonGeneralFailure,
						ResultMessage:     "",
					})
				}
				return response
			}
			response.BatchItem = append(response.BatchItem, resp.BatchItem...)
		}
	}

	responseItems := []kmip.ResponseBatchItem{}
	for _, item := range req.BatchItem {
		respItem := kmip.ResponseBatchItem{
			Operation:         item.Operation,
			UniqueBatchItemID: item.UniqueBatchItemID,
			ResultStatus:      kmip.ResultStatusSuccess,
		}

		op := h.registry.Lookup(item.Operation)
		if op == nil {
			if isProxyEnabledForNotSupportedOperations {
				// Was already delegated and processed
				continue
			}
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
	return response
}

func (h *CryptoHandler) delegateRequestToProxy(ctx context.Context, req *kmip.RequestMessage) (*kmip.ResponseMessage, error) {
	clientCertificates := kmipserver.PeerCertificates(ctx)
	if len(clientCertificates) == 0 {
		return nil, errors.New("no client certificates")
	}

	cc := CertChain{}
	for _, c := range clientCertificates {
		cc.Chain = append(cc.Chain, base64.StdEncoding.EncodeToString(c.Raw))
	}
	b, _ := json.Marshal(cc)

	// Forward request to upstream
	httpReq, err := http.NewRequest(http.MethodPost, h.proxyHttpKMIP.Endpoint, bytes.NewReader(ttlv.MarshalTTLV(req)))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set(HeaderHttpClientCertChain, string(b))
	httpReq.Header.Set(HeaderHttpContentType, "application/octet-stream")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer httpResp.Body.Close()

	dataBody, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, err
	}

	delegatedReq := &kmip.ResponseMessage{}
	err = ttlv.UnmarshalTTLV(dataBody, delegatedReq)
	if err != nil {
		return nil, err
	}

	return delegatedReq, nil
}

func extractCertificatesFromHeaders(headers http.Header) []*x509.Certificate {
	certificates := make([]*x509.Certificate, 0)

	chainData := headers[HeaderHttpClientCertChain]
	if len(chainData) > 0 {
		cc := CertChain{}
		_ = json.Unmarshal([]byte(chainData[0]), &cc)

		for _, certPem := range cc.Chain {
			block, _ := pem.Decode([]byte(certPem))
			if block == nil {
				continue
			}

			cert, _ := x509.ParseCertificate(block.Bytes)
			if cert == nil {
				continue
			}

			certificates = append(certificates, cert)
		}
	}
	return certificates
}

type CertChain struct {
	Chain []string `json:"chain"` // each is base64 DER
}

package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/forta-network/forta-json-rpc-proxy/service"
)

type attesterClient struct {
	attesterUrl string
	authToken   string
}

// NewAttesterClient creates a new attester client.
func NewAttesterClient(attesterUrl, authToken string) *attesterClient {
	return &attesterClient{attesterUrl: attesterUrl, authToken: authToken}
}

type errorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// AttestWithTx retrieves back an attestation.
func (ac *attesterClient) AttestWithTx(ctx context.Context, attReq *service.AttestRequest) (tx hexutil.Bytes, err error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(attReq); err != nil {
		return nil, fmt.Errorf("failed to encode attest request: %v", err)
	}
	req, err := http.NewRequestWithContext(ctx, "POST", ac.attesterUrl+"/attest-tx", &b)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request: %v", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", ac.authToken))

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("attest request failed: %v", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200:
		var respBody struct {
			Tx hexutil.Bytes `json:"tx"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return nil, fmt.Errorf("failed to decode 200 body from attest response: %v", err)
		}
		return respBody.Tx, nil

	case 409:
		var respBody errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
			return nil, fmt.Errorf("failed to decode 409 body from attest response: %v", err)
		}
		return nil, errors.New(respBody.Message)

	default:
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read %d body from attest response: %v", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("attest request failed with code %d: %s", resp.StatusCode, string(b))
	}
}

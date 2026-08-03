package mpesa

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"time"
	"fmt"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	passkey    string
	tokenManager *TokenManager
}

func NewMpesaClient(
	baseURL string,
	httpClient *http.Client,
	passkey string,
) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: httpClient,
		passkey: passkey,
	}
}

func (c *Client) STKPush(ctx context.Context, req STKPushRequest) (*STKPushResponse, error) {

	
	req.Timestamp = generateTimeStamp()
	req.Password = generatePassword(
		req.BusinessShortCode,
		c.passkey,
		req.Timestamp,
	)

	body, err := json.Marshal(req)

	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/mpesa/stkpush/v1/processrequest",
		bytes.NewReader(body),
	)

	if err != nil {
		return nil, err
	}

	token, err := c.tokenManager.GetAccessToken(ctx)

	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Authorization", "Bearer "+token)

	httpReq.Header.Set("Idempotency-Key", req.IdempotencyKey)

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
	return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}
	if err != nil {
		return nil, err
	}
	defer func() {
		//nolint:errcheck
		resp.Body.Close()
	}()

	var stkResp STKPushResponse

	err = json.NewDecoder(resp.Body).Decode(&stkResp)
	if err != nil {
		return nil, err
	}

	return &stkResp, nil
	// req, err := http.NewRequest("POST", "/mpesa/stkpush/v1/processrequest", )
}

func generateTimeStamp() string {
	return time.Now().Format("20060102150405")
}

func generatePassword(shortcode, passkey, timeStamp string) string {
	raw := shortcode + passkey + timeStamp

	return base64.StdEncoding.EncodeToString([]byte(raw))
}

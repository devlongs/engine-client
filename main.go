package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type EngineClient struct {
	endpoint  string
	jwtSecret []byte
	client    *http.Client
}

type PayloadAttributes struct {
	Timestamp             string `json:"timestamp"`
	PrevRandao            string `json:"prevRandao"`
	SuggestedFeeRecipient string `json:"suggestedFeeRecipient"`
}

type ForkChoiceState struct {
	HeadBlockHash      string `json:"headBlockHash"`
	SafeBlockHash      string `json:"safeBlockHash"`
	FinalizedBlockHash string `json:"finalizedBlockHash"`
}

func NewEngineClient(endpoint string, jwtSecret []byte) *EngineClient {
	return &EngineClient{
		endpoint:  endpoint,
		jwtSecret: jwtSecret,
		client:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *EngineClient) generateJWT() (string, error) {
	claims := jwt.MapClaims{
		"iat": time.Now().Unix(),
		"exp": time.Now().Add(time.Minute).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(c.jwtSecret)
}

func (c *EngineClient) makeRequest(ctx context.Context, method string, params interface{}) (map[string]interface{}, error) {
	// Create JSON-RPC request
	request := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      1,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	token, err := c.generateJWT()
	if err != nil {
		return nil, err
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", c.endpoint, bytes.NewReader(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Make the request
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %d", resp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return result, nil
}

// ForkchoiceUpdated sends a forkchoiceUpdated request
func (c *EngineClient) ForkchoiceUpdated(ctx context.Context, state ForkChoiceState, attributes *PayloadAttributes) (map[string]interface{}, error) {
	params := []interface{}{state}
	if attributes != nil {
		params = append(params, attributes)
	}
	return c.makeRequest(ctx, "engine_forkchoiceUpdatedV1", params)
}

// NewPayload sends a newPayload request
func (c *EngineClient) NewPayload(ctx context.Context, payload map[string]interface{}) (map[string]interface{}, error) {
	return c.makeRequest(ctx, "engine_newPayloadV1", []interface{}{payload})
}

func main() {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		fmt.Println("JWT_SECRET environment variable is not set")
		return
	}

	client := NewEngineClient(
		"http://localhost:8551",
		[]byte(jwtSecret),
	)

	forkChoice := ForkChoiceState{
		HeadBlockHash:      "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef",
		SafeBlockHash:      "0xabcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890",
		FinalizedBlockHash: "0x7890abcdef1234567890abcdef1234567890abcdef1234567890abcdef123456",
	}

	attributes := PayloadAttributes{
		Timestamp:             fmt.Sprintf("0x%x", time.Now().Unix()),
		PrevRandao:            "0xabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdefabcdef",
		SuggestedFeeRecipient: "0xabc123abc123abc123abc123abc123abc123abc1",
	}

	ctx := context.Background()
	result, err := client.ForkchoiceUpdated(ctx, forkChoice, &attributes)
	if err != nil {
		fmt.Printf("Error making forkchoice update request: %v\n", err)
		return
	}

	prettyResult, _ := json.MarshalIndent(result, "", "  ")
	fmt.Printf("Forkchoice update result: %s\n", string(prettyResult))
}

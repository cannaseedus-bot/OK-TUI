package xjson

import (
	"encoding/json"
	"testing"
)

func TestNewInferRequest(t *testing.T) {
	req := NewInferRequest("llama2", "Hello, world!")

	if req.Model != "llama2" {
		t.Errorf("expected model llama2, got %s", req.Model)
	}
	if req.Prompt != "Hello, world!" {
		t.Errorf("expected prompt 'Hello, world!', got %s", req.Prompt)
	}
	if req.Runner != "lam.o" {
		t.Errorf("expected runner lam.o, got %s", req.Runner)
	}
	if req.Mode != "chat" {
		t.Errorf("expected mode chat, got %s", req.Mode)
	}
	if req.Params == nil {
		t.Error("expected params to be initialized")
	}
}

func TestInferRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     *InferRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     NewInferRequest("llama2", "test prompt"),
			wantErr: false,
		},
		{
			name: "missing runner",
			req: &InferRequest{
				Model:  "llama2",
				Prompt: "test",
			},
			wantErr: true,
		},
		{
			name: "missing model",
			req: &InferRequest{
				Runner: "lam.o",
				Prompt: "test",
			},
			wantErr: true,
		},
		{
			name: "missing prompt and context",
			req: &InferRequest{
				Runner: "lam.o",
				Model:  "llama2",
			},
			wantErr: true,
		},
		{
			name: "valid with context",
			req: &InferRequest{
				Runner:  "lam.o",
				Model:   "llama2",
				Context: []Message{{Role: "user", Content: "hello"}},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestInferRequestWithParams(t *testing.T) {
	params := &InferParams{
		Temperature: 0.5,
		TopP:        0.8,
		MaxTokens:   1024,
	}

	req := NewInferRequest("llama2", "test")
	req = req.WithParams(params)

	if req.Params.Temperature != 0.5 {
		t.Errorf("expected temperature 0.5, got %f", req.Params.Temperature)
	}
	if req.Params.TopP != 0.8 {
		t.Errorf("expected top_p 0.8, got %f", req.Params.TopP)
	}
	if req.Params.MaxTokens != 1024 {
		t.Errorf("expected max_tokens 1024, got %d", req.Params.MaxTokens)
	}
}

func TestNewCompletionResponse(t *testing.T) {
	resp := NewCompletionResponse("llama2", "lam.o", "Hello, response!")

	if resp.Model != "llama2" {
		t.Errorf("expected model llama2, got %s", resp.Model)
	}
	if resp.Runner != "lam.o" {
		t.Errorf("expected runner lam.o, got %s", resp.Runner)
	}
	if resp.Text != "Hello, response!" {
		t.Errorf("expected text 'Hello, response!', got %s", resp.Text)
	}
	if !resp.Done {
		t.Error("expected done to be true")
	}
	if resp.SCXQ2 == "" {
		t.Error("expected SCXQ2 fingerprint to be generated")
	}
}

func TestCompletionResponseWithTokens(t *testing.T) {
	resp := NewCompletionResponse("llama2", "lam.o", "response")
	resp = resp.WithTokens(10, 20)

	if resp.Tokens == nil {
		t.Error("expected tokens to be set")
		return
	}
	if resp.Tokens.Input != 10 {
		t.Errorf("expected input tokens 10, got %d", resp.Tokens.Input)
	}
	if resp.Tokens.Output != 20 {
		t.Errorf("expected output tokens 20, got %d", resp.Tokens.Output)
	}
	if resp.Tokens.Total != 30 {
		t.Errorf("expected total tokens 30, got %d", resp.Tokens.Total)
	}
}

func TestCompletionResponseFingerprint(t *testing.T) {
	resp1 := NewCompletionResponse("llama2", "lam.o", "response")
	resp2 := NewCompletionResponse("llama2", "lam.o", "response")

	if resp1.SCXQ2 != resp2.SCXQ2 {
		t.Error("expected same fingerprint for identical responses")
	}

	resp3 := NewCompletionResponse("llama2", "lam.o", "different")
	if resp1.SCXQ2 == resp3.SCXQ2 {
		t.Error("expected different fingerprints for different responses")
	}
}

func TestNewErrorResponse(t *testing.T) {
	errResp := NewErrorResponse("lam.o", "Test error message", 500)

	if errResp.Runner != "lam.o" {
		t.Errorf("expected runner lam.o, got %s", errResp.Runner)
	}
	if errResp.Message != "Test error message" {
		t.Errorf("expected message 'Test error message', got %s", errResp.Message)
	}
	if errResp.Code != 500 {
		t.Errorf("expected code 500, got %d", errResp.Code)
	}
}

func TestErrorResponseWithDetails(t *testing.T) {
	errResp := NewErrorResponse("lam.o", "error", 500)
	errResp = errResp.WithDetails("additional context")

	if errResp.Details != "additional context" {
		t.Errorf("expected details 'additional context', got %s", errResp.Details)
	}
}

func TestXJSONEnvelopeMarshal(t *testing.T) {
	req := NewInferRequest("llama2", "test")
	data, err := req.ToJSON()

	if err != nil {
		t.Errorf("ToJSON() error = %v", err)
		return
	}

	var envelope XJSONEnvelope
	err = json.Unmarshal(data, &envelope)
	if err != nil {
		t.Errorf("json.Unmarshal() error = %v", err)
		return
	}

	if !envelope.IsInferRequest() {
		t.Error("expected envelope to be an infer request")
	}
	if envelope.IsCompletion() || envelope.IsError() {
		t.Error("expected envelope to only be an infer request")
	}
}

func TestXJSONEnvelopeDetection(t *testing.T) {
	tests := []struct {
		name         string
		envelope     *XJSONEnvelope
		expectInfer  bool
		expectComplete bool
		expectError  bool
	}{
		{
			name: "infer request",
			envelope: &XJSONEnvelope{
				Infer: NewInferRequest("llama2", "test"),
			},
			expectInfer: true,
		},
		{
			name: "completion response",
			envelope: &XJSONEnvelope{
				Completion: NewCompletionResponse("llama2", "lam.o", "response"),
			},
			expectComplete: true,
		},
		{
			name: "error response",
			envelope: &XJSONEnvelope{
				Error: NewErrorResponse("lam.o", "error", 500),
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envelope.IsInferRequest() != tt.expectInfer {
				t.Errorf("IsInferRequest() = %v, want %v", tt.envelope.IsInferRequest(), tt.expectInfer)
			}
			if tt.envelope.IsCompletion() != tt.expectComplete {
				t.Errorf("IsCompletion() = %v, want %v", tt.envelope.IsCompletion(), tt.expectComplete)
			}
			if tt.envelope.IsError() != tt.expectError {
				t.Errorf("IsError() = %v, want %v", tt.envelope.IsError(), tt.expectError)
			}
		})
	}
}

func TestParseXJSON(t *testing.T) {
	req := NewInferRequest("llama2", "test prompt")
	data, _ := req.ToJSON()

	envelope, err := ParseXJSON(data)
	if err != nil {
		t.Errorf("ParseXJSON() error = %v", err)
		return
	}

	if !envelope.IsInferRequest() {
		t.Error("expected parsed envelope to be an infer request")
		return
	}
	if envelope.Infer.Model != "llama2" {
		t.Errorf("expected model llama2, got %s", envelope.Infer.Model)
	}
	if envelope.Infer.Prompt != "test prompt" {
		t.Errorf("expected prompt 'test prompt', got %s", envelope.Infer.Prompt)
	}
}

func TestParseInvalidXJSON(t *testing.T) {
	_, err := ParseXJSON([]byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDefaultParams(t *testing.T) {
	params := DefaultParams()

	if params.Temperature != 0.7 {
		t.Errorf("expected default temperature 0.7, got %f", params.Temperature)
	}
	if params.TopP != 0.9 {
		t.Errorf("expected default top_p 0.9, got %f", params.TopP)
	}
	if params.TopK != 40 {
		t.Errorf("expected default top_k 40, got %d", params.TopK)
	}
	if params.MaxTokens != 2048 {
		t.Errorf("expected default max_tokens 2048, got %d", params.MaxTokens)
	}
}

func TestCreateInferRequest(t *testing.T) {
	opts := map[string]interface{}{
		"model":  "llama2",
		"prompt": "test",
		"mode":   "code",
		"params": map[string]interface{}{
			"temperature": 0.3,
			"top_p":       0.5,
			"max_tokens":  512.0,
		},
	}

	req := CreateInferRequest(opts)

	if req.Model != "llama2" {
		t.Errorf("expected model llama2, got %s", req.Model)
	}
	if req.Mode != "code" {
		t.Errorf("expected mode code, got %s", req.Mode)
	}
	if req.Params.Temperature != 0.3 {
		t.Errorf("expected temperature 0.3, got %f", req.Params.Temperature)
	}
	if req.Params.MaxTokens != 512 {
		t.Errorf("expected max_tokens 512, got %d", req.Params.MaxTokens)
	}
}

func TestCreateCompletionResponse(t *testing.T) {
	opts := map[string]interface{}{
		"model":  "llama2",
		"text":   "response text",
		"tokens": map[string]interface{}{
			"input":  5.0,
			"output": 15.0,
		},
		"metrics": map[string]interface{}{
			"latency_ms": 100.0,
			"backend":    "ollama",
		},
	}

	resp := CreateCompletionResponse(opts)

	if resp.Model != "llama2" {
		t.Errorf("expected model llama2, got %s", resp.Model)
	}
	if resp.Text != "response text" {
		t.Errorf("expected text 'response text', got %s", resp.Text)
	}
	if resp.Tokens == nil || resp.Tokens.Input != 5 {
		t.Error("expected tokens to be set with input 5")
	}
	if resp.Metrics == nil || resp.Metrics.LatencyMS != 100.0 {
		t.Error("expected metrics to be set with latency 100")
	}
}

func TestCreateError(t *testing.T) {
	opts := map[string]interface{}{
		"runner":  "lam.o",
		"message": "test error",
		"code":    400.0,
		"details": "error details",
	}

	errResp := CreateError(opts)

	if errResp.Runner != "lam.o" {
		t.Errorf("expected runner lam.o, got %s", errResp.Runner)
	}
	if errResp.Message != "test error" {
		t.Errorf("expected message 'test error', got %s", errResp.Message)
	}
	if errResp.Code != 400 {
		t.Errorf("expected code 400, got %d", errResp.Code)
	}
	if errResp.Details != "error details" {
		t.Errorf("expected details 'error details', got %s", errResp.Details)
	}
}

func TestInferRequestFingerprint(t *testing.T) {
	req1 := NewInferRequest("llama2", "test")
	req2 := NewInferRequest("llama2", "test")

	fp1 := req1.Fingerprint()
	fp2 := req2.Fingerprint()

	if fp1 != fp2 {
		t.Error("expected same fingerprint for identical requests")
	}

	req3 := NewInferRequest("llama2", "different")
	fp3 := req3.Fingerprint()

	if fp1 == fp3 {
		t.Error("expected different fingerprints for different requests")
	}
}

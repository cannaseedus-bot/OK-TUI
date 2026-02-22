package xjson

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"
)

// MockRequest simulates an XJSON request with validation
type MockRequest struct {
	XJSONEnvelope
}

// ValidateXJSONRequest is a middleware function that validates XJSON requests
func ValidateXJSONRequest(data []byte) (*XJSONEnvelope, error) {
	var envelope XJSONEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return nil, err
	}
	return &envelope, nil
}

// ValidateXJSONResponse ensures responses are properly formatted
func ValidateXJSONResponse(data interface{}) ([]byte, error) {
	switch v := data.(type) {
	case *CompletionResponse:
		envelope := &XJSONEnvelope{Completion: v}
		return json.Marshal(envelope)
	case *ErrorResponse:
		envelope := &XJSONEnvelope{Error: v}
		return json.Marshal(envelope)
	case *InferRequest:
		envelope := &XJSONEnvelope{Infer: v}
		return json.Marshal(envelope)
	default:
		return nil, nil
	}
}

func TestValidateXJSONRequest(t *testing.T) {
	req := NewInferRequest("llama2", "test")
	envelope := &XJSONEnvelope{Infer: req}
	data, _ := json.Marshal(envelope)

	validated, err := ValidateXJSONRequest(data)
	if err != nil {
		t.Errorf("ValidateXJSONRequest() error = %v", err)
		return
	}

	if !validated.IsInferRequest() {
		t.Error("expected validated envelope to be an infer request")
	}
}

func TestValidateXJSONRequestInvalid(t *testing.T) {
	_, err := ValidateXJSONRequest([]byte("not json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestXJSONRequestStream(t *testing.T) {
	req := NewInferRequest("llama2", "test")
	envelope := &XJSONEnvelope{Infer: req}
	data, _ := json.Marshal(envelope)

	// Simulate reading from a stream
	stream := io.NopCloser(bytes.NewReader(data))
	defer stream.Close()

	body, _ := io.ReadAll(stream)
	validated, err := ValidateXJSONRequest(body)
	if err != nil {
		t.Errorf("ValidateXJSONRequest() error = %v", err)
		return
	}

	if !validated.IsInferRequest() {
		t.Error("expected stream-read envelope to be an infer request")
	}
}

func TestXJSONResponseSerialization(t *testing.T) {
	tests := []struct {
		name     string
		response interface{}
		isValid  bool
	}{
		{
			name:     "completion response",
			response: NewCompletionResponse("llama2", "lam.o", "test"),
			isValid:  true,
		},
		{
			name:     "error response",
			response: NewErrorResponse("lam.o", "error", 500),
			isValid:  true,
		},
		{
			name:     "infer request",
			response: NewInferRequest("llama2", "test"),
			isValid:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := ValidateXJSONResponse(tt.response)
			if err != nil {
				t.Errorf("ValidateXJSONResponse() error = %v", err)
				return
			}
			if data == nil {
				t.Error("expected serialized XJSON data")
				return
			}

			var result XJSONEnvelope
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Errorf("json.Unmarshal() error = %v", err)
			}
		})
	}
}

func TestXJSONChaining(t *testing.T) {
	// Simulate a request-response chain
	req := NewInferRequest("llama2", "What is 2+2?")
	req = req.WithMode("math")

	// Serialize request
	reqEnvelope := &XJSONEnvelope{Infer: req}
	reqData, _ := json.Marshal(reqEnvelope)

	// Validate and deserialize request
	validatedReq, _ := ValidateXJSONRequest(reqData)

	// Create response based on request
	resp := NewCompletionResponse(
		validatedReq.Infer.Model,
		validatedReq.Infer.Runner,
		"2+2=4",
	)
	resp = resp.WithTokens(5, 4)

	// Serialize response
	respData, _ := ValidateXJSONResponse(resp)

	// Validate response
	validatedResp, _ := ValidateXJSONRequest(respData)
	if !validatedResp.IsCompletion() {
		t.Error("expected response to be a completion")
	}
	if validatedResp.Completion.Text != "2+2=4" {
		t.Errorf("expected response text '2+2=4', got %s", validatedResp.Completion.Text)
	}
}

func TestXJSONErrorHandling(t *testing.T) {
	errResp := NewErrorResponse("lam.o", "Model not found", 404)
	errResp = errResp.WithDetails("llama2 model is not available")

	data, _ := ValidateXJSONResponse(errResp)
	validated, _ := ValidateXJSONRequest(data)

	if !validated.IsError() {
		t.Error("expected validated envelope to be an error")
		return
	}

	if validated.Error.Code != 404 {
		t.Errorf("expected error code 404, got %d", validated.Error.Code)
	}
	if validated.Error.Details != "llama2 model is not available" {
		t.Errorf("expected error details, got %s", validated.Error.Details)
	}
}

func TestXJSONRoundTrip(t *testing.T) {
	tests := []struct {
		name      string
		original  interface{}
		detector  func(*XJSONEnvelope) bool
	}{
		{
			name:     "infer request",
			original: NewInferRequest("llama2", "test"),
			detector: func(e *XJSONEnvelope) bool { return e.IsInferRequest() },
		},
		{
			name:     "completion response",
			original: NewCompletionResponse("llama2", "lam.o", "response"),
			detector: func(e *XJSONEnvelope) bool { return e.IsCompletion() },
		},
		{
			name:     "error response",
			original: NewErrorResponse("lam.o", "error", 500),
			detector: func(e *XJSONEnvelope) bool { return e.IsError() },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Serialize
			serialized, err := ValidateXJSONResponse(tt.original)
			if err != nil {
				t.Errorf("ValidateXJSONResponse() error = %v", err)
				return
			}

			// Deserialize
			deserialized, err := ValidateXJSONRequest(serialized)
			if err != nil {
				t.Errorf("ValidateXJSONRequest() error = %v", err)
				return
			}

			// Verify type
			if !tt.detector(deserialized) {
				t.Error("expected deserialized envelope to match original type")
			}
		})
	}
}

func TestXJSONMultipleRequests(t *testing.T) {
	requests := []*InferRequest{
		NewInferRequest("llama2", "first prompt"),
		NewInferRequest("mistral", "second prompt"),
		NewInferRequest("neural-chat", "third prompt"),
	}

	for i, req := range requests {
		envelope := &XJSONEnvelope{Infer: req}
		data, _ := json.Marshal(envelope)

		validated, _ := ValidateXJSONRequest(data)
		if validated.Infer.Model != requests[i].Model {
			t.Errorf("request %d: expected model %s, got %s", i, requests[i].Model, validated.Infer.Model)
		}
	}
}

func TestXJSONContentValidation(t *testing.T) {
	req := NewInferRequest("llama2", "")
	err := req.Validate()
	if err == nil {
		t.Error("expected validation error for empty prompt with no context")
	}

	req.Context = []Message{{Role: "user", Content: "hello"}}
	err = req.Validate()
	if err != nil {
		t.Errorf("expected validation to pass with context, got %v", err)
	}
}

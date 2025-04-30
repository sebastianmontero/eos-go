package eos_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/sebastianmontero/eos-go"
)

// mockRoundTripper allows us to simulate different endpoint behaviors
// by returning different responses or errors based on the request URL.
type mockRoundTripper struct {
	failURLs     map[string]bool
	responseBody [][]byte
	statusCode   int
	calls        int
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.failURLs[req.URL.String()] {
		return nil, errors.New("simulated endpoint failure")
	}
	resp := &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(bytes.NewReader(m.responseBody[m.calls])),
		Header:     make(http.Header),
	}
	if m.calls < len(m.responseBody)-1 {
		m.calls++
	}
	return resp, nil
}

func getInfoMockRoundTripper(chainIDs []string) *mockRoundTripper {
	responseBodies := make([][]byte, len(chainIDs))
	for i, chainID := range chainIDs {
		responseBodies[i] = []byte(`{"head_block_num":123456789, "chain_id": "` + chainID + `"}`)
	}
	return &mockRoundTripper{
		failURLs:     map[string]bool{},
		responseBody: responseBodies,
		statusCode:   200,
	}
}

func TestAPIVerifyEndpoints_Success(t *testing.T) {
	endpoints := []string{"http://endpoint1", "http://endpoint2", "http://endpoint3"}
	// Simulate endpoint1 and endpoint2 failing, endpoint3 succeeds

	mockClient := &http.Client{
		Transport: getInfoMockRoundTripper([]string{"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906"}),
		Timeout:   2 * time.Second,
	}
	api := &eos.TestAPIWrapper{
		API: eos.API{
			HttpClient: mockClient,
			BaseURLs:   endpoints,
		},
	}

	chainID, err := api.VerifyEndpoints()
	require.NoError(t, err, "VerifyEndpoints should succeed")
	require.Equal(t, "aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906", chainID.String())

}

func TestAPIVerifyEndpoints_Fail(t *testing.T) {
	endpoints := []string{"http://endpoint1", "http://endpoint2", "http://endpoint3"}
	// Simulate endpoint1 and endpoint2 failing, endpoint3 succeeds

	mockClient := &http.Client{
		Transport: getInfoMockRoundTripper([]string{"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906", "8a34ec7df1b8cd06ff4a8abbaa7cc50300823350cadc59ab296cb00d104d2b8f"}),
		Timeout:   2 * time.Second,
	}
	api := &eos.TestAPIWrapper{
		API: eos.API{
			HttpClient: mockClient,
			BaseURLs:   endpoints,
		},
	}

	_, err := api.VerifyEndpoints()
	require.Error(t, err, "VerifyEndpoints should fail")
	require.Equal(t, "all endpoints must be of the same chain ID", err.Error())

}

func TestAPICall_Failover(t *testing.T) {
	endpoints := []string{"http://endpoint1", "http://endpoint2", "http://endpoint3"}
	// Simulate endpoint1 and endpoint2 failing, endpoint3 succeeds
	failURLs := map[string]bool{
		"http://endpoint1/v1/testapi/test": true,
		"http://endpoint2/v1/testapi/test": true,
	}
	respData := map[string]interface{}{"success": true}
	respBytes, _ := json.Marshal(respData)

	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			failURLs:     failURLs,
			responseBody: [][]byte{respBytes},
			statusCode:   200,
		},
		Timeout: 2 * time.Second,
	}
	api := &eos.API{
		HttpClient: mockClient,
		BaseURLs:   endpoints,
	}

	var out map[string]interface{}
	err := api.Call(context.Background(), "testapi", "test", nil, &out)
	require.NoError(t, err, "Call should succeed via failover")
	require.Equal(t, true, out["success"])
}

func TestAPICall_AllEndpointsFail(t *testing.T) {
	endpoints := []string{"http://endpoint1", "http://endpoint2"}
	failURLs := map[string]bool{
		"http://endpoint1/v1/testapi/test": true,
		"http://endpoint2/v1/testapi/test": true,
	}
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			failURLs:     failURLs,
			responseBody: nil,
			statusCode:   500,
		},
		Timeout: 2 * time.Second,
	}
	api := &eos.API{
		HttpClient: mockClient,
		BaseURLs:   endpoints,
	}

	var out map[string]interface{}
	err := api.Call(context.Background(), "testapi", "test", nil, &out)
	require.Error(t, err, "Call should fail if all endpoints fail")
}

func TestAPICall_SingleEndpointSuccess(t *testing.T) {
	endpoints := []string{"http://endpoint1"}
	failURLs := map[string]bool{}
	respData := map[string]interface{}{"ok": 1}
	respBytes, _ := json.Marshal(respData)
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			failURLs:     failURLs,
			responseBody: [][]byte{respBytes},
			statusCode:   200,
		},
		Timeout: 2 * time.Second,
	}
	api := &eos.API{
		HttpClient: mockClient,
		BaseURLs:   endpoints,
	}

	var out map[string]interface{}
	err := api.Call(context.Background(), "testapi", "test", nil, &out)
	require.NoError(t, err, "Call should succeed on single endpoint")
	require.Equal(t, float64(1), out["ok"])
}

func TestAPICall_ReturnsAPIErrorImmediately(t *testing.T) {
	endpoints := []string{"http://endpoint1", "http://endpoint2"}
	// Provide a complete error structure matching APIError expectations
	apiErr := map[string]interface{}{
		"code":    400,
		"message": "Bad Request",
		"error": map[string]interface{}{
			"code":    123,
			"name":    "bad_request",
			"what":    "bad request",
			"details": []interface{}{},
		},
	}
	apiErrBytes, _ := json.Marshal(apiErr)
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			failURLs:     map[string]bool{},
			responseBody: [][]byte{apiErrBytes},
			statusCode:   400,
		},
		Timeout: 2 * time.Second,
	}
	api := &eos.API{
		HttpClient: mockClient,
		BaseURLs:   endpoints,
	}
	var out map[string]interface{}
	err := api.Call(context.Background(), "testapi", "test", nil, &out)
	require.Error(t, err, "Call should return an error immediately if APIError is returned")
	if errTyped, ok := err.(eos.APIError); ok {
		require.Equal(t, 400, errTyped.Code)
		require.Equal(t, "Bad Request", errTyped.Message)
	} else {
		t.Fatalf("Expected APIError, got %T: %v", err, err)
	}
}

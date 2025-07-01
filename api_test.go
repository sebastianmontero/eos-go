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

func TestAPIVerifyEndpoints_AllEndpointsValid(t *testing.T) {
	endpoints := []string{"http://endpoint1", "http://endpoint2", "http://endpoint3"}
	
	api := &eos.TestAPIWrapper{
		API: eos.API{
			HttpClient: &http.Client{
				Transport: getInfoMockRoundTripper([]string{
					"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
					"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
					"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
				}),
				Timeout: 2 * time.Second,
			},
			BaseURLs: endpoints,
		},
	}

	t.Run("strict mode", func(t *testing.T) {
		chainID, err := api.VerifyEndpoints(true)
		require.NoError(t, err, "VerifyEndpoints should succeed with all endpoints valid in strict mode")
		require.Equal(t, "aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906", chainID.String())
	})

	t.Run("non-strict mode", func(t *testing.T) {
		chainID, err := api.VerifyEndpoints(false)
		require.NoError(t, err, "VerifyEndpoints should succeed with all endpoints valid in non-strict mode")
		require.Equal(t, "aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906", chainID.String())
	})
}

func TestAPIVerifyEndpoints_SomeEndpointsFail(t *testing.T) {
	endpoints := []string{"http://endpoint1", "http://endpoint2", "http://endpoint3"}
	
	t.Run("strict mode - should fail", func(t *testing.T) {
		mockTransport := getInfoMockRoundTripper([]string{
			"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
			"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
		})
		mockTransport.failURLs = map[string]bool{
			"http://endpoint3/v1/chain/get_info": true,
		}

		api := &eos.TestAPIWrapper{
			API: eos.API{
				HttpClient: &http.Client{
					Transport: mockTransport,
					Timeout:   2 * time.Second,
				},
				BaseURLs: endpoints,
			},
		}

		_, err := api.VerifyEndpoints(true)
		require.Error(t, err, "VerifyEndpoints should fail in strict mode when some endpoints fail")
		require.Contains(t, err.Error(), "http://endpoint3/v1/chain/get_info: Post \"http://endpoint3/v1/chain/get_info\": simulated endpoint failure")
	})

	t.Run("non-strict mode - should succeed", func(t *testing.T) {
		mockTransport := getInfoMockRoundTripper([]string{
			"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
			"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
		})
		mockTransport.failURLs = map[string]bool{
			"http://endpoint3/v1/chain/get_info": true,
		}

		api := &eos.TestAPIWrapper{
			API: eos.API{
				HttpClient: &http.Client{
					Transport: mockTransport,
					Timeout:   2 * time.Second,
				},
				BaseURLs: endpoints,
			},
		}

		chainID, err := api.VerifyEndpoints(false)
		require.NoError(t, err, "VerifyEndpoints should succeed in non-strict mode with some endpoints failing")
		require.Equal(t, "aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906", chainID.String())
	})
}

func TestAPIVerifyEndpoints_AllEndpointsFail(t *testing.T) {
	endpoints := []string{"http://endpoint1", "http://endpoint2", "http://endpoint3"}
	
	t.Run("strict mode - should fail", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			failURLs: map[string]bool{
				"http://endpoint1/v1/chain/get_info": true,
			},
			responseBody: [][]byte{},
			statusCode:   200,
		}

		api := &eos.TestAPIWrapper{
			API: eos.API{
				HttpClient: &http.Client{
					Transport: mockTransport,
					Timeout:   2 * time.Second,
				},
				BaseURLs: endpoints,
			},
		}

		_, err := api.VerifyEndpoints(true)
		require.Error(t, err, "VerifyEndpoints should fail in strict mode when first endpoint fails")
		require.Contains(t, err.Error(), "http://endpoint1/v1/chain/get_info: Post \"http://endpoint1/v1/chain/get_info\": simulated endpoint failure")
	})

	t.Run("non-strict mode - should fail with last error", func(t *testing.T) {
		mockTransport := &mockRoundTripper{
			failURLs: map[string]bool{
				"http://endpoint1/v1/chain/get_info": true,
				"http://endpoint2/v1/chain/get_info": true,
				"http://endpoint3/v1/chain/get_info": true,
			},
			responseBody: [][]byte{},
			statusCode:   200,
		}

		api := &eos.TestAPIWrapper{
			API: eos.API{
				HttpClient: &http.Client{
					Transport: mockTransport,
					Timeout:   2 * time.Second,
				},
				BaseURLs: endpoints,
			},
		}

		_, err := api.VerifyEndpoints(false)
		require.Error(t, err, "VerifyEndpoints should fail when all endpoints fail")
		require.Contains(t, err.Error(), "no valid endpoints could be reached, last error: endpoint http://endpoint3: http://endpoint3/v1/chain/get_info: Post \"http://endpoint3/v1/chain/get_info\": simulated endpoint failure")
	})
}

func TestAPIVerifyEndpoints_ChainIDMismatch(t *testing.T) {
	endpoints := []string{"http://endpoint1", "http://endpoint2"}
	
	t.Run("strict mode - should fail on first mismatch", func(t *testing.T) {
		mockTransport := getInfoMockRoundTripper([]string{
			"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
			"8a34ec7df1b8cd06ff4a8abbaa7cc50300823350cadc59ab296cb00d104d2b8f", // Different chain ID
		})

		api := &eos.TestAPIWrapper{
			API: eos.API{
				HttpClient: &http.Client{
					Transport: mockTransport,
					Timeout:   2 * time.Second,
				},
				BaseURLs: endpoints,
			},
		}

		_, err := api.VerifyEndpoints(true)
		require.Error(t, err, "VerifyEndpoints should fail in strict mode when chain IDs don't match")
		require.Equal(t, "all endpoints must be of the same chain ID", err.Error())
	})

	t.Run("non-strict mode - should fail on first mismatch", func(t *testing.T) {
		mockTransport := getInfoMockRoundTripper([]string{
			"aca376f206b8fc25a6ed44dbdc66547c36c6c33e3a119ffbeaef943642f0e906",
			"8a34ec7df1b8cd06ff4a8abbaa7cc50300823350cadc59ab296cb00d104d2b8f", // Different chain ID
		})

		api := &eos.TestAPIWrapper{
			API: eos.API{
				HttpClient: &http.Client{
					Transport: mockTransport,
					Timeout:   2 * time.Second,
				},
				BaseURLs: endpoints,
			},
		}

		_, err := api.VerifyEndpoints(false)
		require.Error(t, err, "VerifyEndpoints should fail in non-strict mode when chain IDs don't match")
		require.Equal(t, "all endpoints must be of the same chain ID", err.Error())
	})
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

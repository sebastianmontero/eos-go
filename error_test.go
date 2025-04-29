package eos

import (
	"encoding/json"
	"testing"
)

func TestAPIErrorUnmarshal(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		expected APIError
	}{
		{
			name: "full error with two details",
			jsonData: `{
				"code": 500,
				"message": "Internal Server Error",
				"error": {
					"code": 123456,
					"name": "some_error_name",
					"what": "Something went wrong",
					"details": [
						{
							"message": "Detail message 1",
							"file": "file1.go",
							"line_number": 10,
							"method": "MethodA"
						},
						{
							"message": "Detail message 2",
							"file": "file2.go",
							"line_number": 20,
							"method": "MethodB"
						}
					]
				}
			}`,
			expected: APIError{
				Code:    500,
				Message: "Internal Server Error",
				ErrorStruct: apiErrorStruct{
					Code: 123456,
					Name: "some_error_name",
					What: "Something went wrong",
					Details: []APIErrorDetail{
						{Message: "Detail message 1", File: "file1.go", LineNumber: 10, Method: "MethodA"},
						{Message: "Detail message 2", File: "file2.go", LineNumber: 20, Method: "MethodB"},
					},
				},
			},
		},
		{
			name: "unknown key error",
			jsonData: `{
				"code": 400,
				"message": "Account lookup",
				"error": {
					"code": 3060002,
					"name": "account_query_exception",
					"what": "Account Query Exception",
					"details": [
						{
							"message": "unable to retrieve account info (unknown key (boost::tuples::tuple<bool, eosio::chain::name, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type>): (0 nonexistant))",
							"file": "chain_plugin.cpp",
							"line_number": 2594,
							"method": "get_account"
						}
					]
				}
			}`,
			expected: APIError{
				Code:    400,
				Message: "Account lookup",
				ErrorStruct: apiErrorStruct{
					Code: 3060002,
					Name: "account_query_exception",
					What: "Account Query Exception",
					Details: []APIErrorDetail{
						{Message: "unable to retrieve account info (unknown key (boost::tuples::tuple<bool, eosio::chain::name, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type, boost::tuples::null_type>): (0 nonexistant))", File: "chain_plugin.cpp", LineNumber: 2594, Method: "get_account"},
					},
				},
			},
		},
		{
			name: "unknown key error",
			jsonData: `{
				"code": 500,
				"message": "Internal Service Error",
				"error": {
					"code": 3080004,
					"name": "tx_cpu_usage_exceeded",
					"what": "Transaction exceeded the current CPU usage limit imposed on the transaction",
					"details": [
						{
							"message": "transaction 60f0fab2544e0cb946e673caf95e1d7056fd8bbae89903db6167346571bcb404 was executing for too long 150101us reached on chain max_transaction_cpu_usage 150000us",
							"file": "transaction_context.cpp",
							"line_number": 482,
							"method": "checktime"
						},
						{
							"message": "testcontract <= testcontract::toolong pending console output: ",
							"file": "apply_context.cpp",
							"line_number": 134,
							"method": "exec_one"
						}
					]
				}
			}`,
			expected: APIError{
				Code:    500,
				Message: "Internal Service Error",
				ErrorStruct: apiErrorStruct{
					Code: 3080004,
					Name: "tx_cpu_usage_exceeded",
					What: "Transaction exceeded the current CPU usage limit imposed on the transaction",
					Details: []APIErrorDetail{
						{Message: "transaction 60f0fab2544e0cb946e673caf95e1d7056fd8bbae89903db6167346571bcb404 was executing for too long 150101us reached on chain max_transaction_cpu_usage 150000us", File: "transaction_context.cpp", LineNumber: 482, Method: "checktime"},
						{Message: "testcontract <= testcontract::toolong pending console output: ", File: "apply_context.cpp", LineNumber: 134, Method: "exec_one"},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var actual APIError
			err := json.Unmarshal([]byte(tt.jsonData), &actual)
			if err != nil {
				t.Fatalf("Failed to unmarshal: %v", err)
			}
			compareAPIError(t, actual, tt.expected)
		})
	}
}

func TestAPIErrorUnmarshal_Invalid(t *testing.T) {
	cases := []struct {
		name     string
		jsonData string
	}{
		{
			name:     "invalid json syntax",
			jsonData: `{`,
		},
		{
			name:     "wrong type for code",
			jsonData: `{"code": "not_an_int", "message": "msg", "error": {"code": 1, "name": "n", "what": "w", "details": []}}`,
		},
		{
			name:     "wrong type for error.details",
			jsonData: `{"code": 1, "message": "msg", "error": {"code": 1, "name": "n", "what": "w", "details": {}}}`,
		},
		{
			name:     "empty object",
			jsonData: `{}`,
		},
		{
			name:     "missing message field",
			jsonData: `{"code": 1, "error": {"code": 1, "name": "n", "what": "w", "details": []}}`,
		},
		{
			name:     "missing error field",
			jsonData: `{"code": 1, "message": "msg"}`,
		},
		{
			name:     "missing error.name field",
			jsonData: `{"code": 1, "message": "msg", "error": {"code": 1, "what": "w", "details": []}}`,
		},
		{
			name:     "missing error.what field",
			jsonData: `{"code": 1, "message": "msg", "error": {"code": 1, "name": "n", "details": []}}`,
		},
		{
			name:     "extra field at root",
			jsonData: `{"code": 1, "message": "msg", "extra": 42, "error": {"code": 1, "name": "n", "what": "w", "details": []}}`,
		},
		{
			name:     "extra field in error struct",
			jsonData: `{"code": 1, "message": "msg", "error": {"code": 1, "name": "n", "what": "w", "details": [], "extra": 42}}`,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var apiErr APIError
			err := json.Unmarshal([]byte(c.jsonData), &apiErr)
			if err == nil {
				t.Errorf("Expected unmarshal to fail for case '%s', but it succeeded", c.name)
			}
		})
	}
}

func TestAPIErrorUnmarshal_ValidMinimal(t *testing.T) {
	jsonData := `{"code": 1, "message": "msg", "error": {"code": 2, "name": "n", "what": "w", "details": []}}`
	var apiErr APIError
	err := json.Unmarshal([]byte(jsonData), &apiErr)
	if err != nil {
		t.Fatalf("Expected minimal valid JSON to unmarshal, got error: %v", err)
	}
	if apiErr.Message != "msg" || apiErr.ErrorStruct.Name != "n" || apiErr.ErrorStruct.What != "w" {
		t.Errorf("Fields not set as expected: got %+v", apiErr)
	}
}

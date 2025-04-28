package eos

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestAPIErrorUnmarshal(t *testing.T) {
	jsonData := `{
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
	}`

	var apiErr APIError
	err := json.Unmarshal([]byte(jsonData), &apiErr)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if apiErr.Code != 500 {
		t.Errorf("Expected code 500, got %d", apiErr.Code)
	}
	if apiErr.Message != "Internal Server Error" {
		t.Errorf("Expected message 'Internal Server Error', got '%s'", apiErr.Message)
	}
	if apiErr.ErrorStruct.Code != 123456 {
		t.Errorf("Expected error.code 123456, got %d", apiErr.ErrorStruct.Code)
	}
	if apiErr.ErrorStruct.Name != "some_error_name" {
		t.Errorf("Expected error.name 'some_error_name', got '%s'", apiErr.ErrorStruct.Name)
	}
	if apiErr.ErrorStruct.What != "Something went wrong" {
		t.Errorf("Expected error.what 'Something went wrong', got '%s'", apiErr.ErrorStruct.What)
	}

	expectedDetails := []APIErrorDetail{
		{Message: "Detail message 1", File: "file1.go", LineNumber: 10, Method: "MethodA"},
		{Message: "Detail message 2", File: "file2.go", LineNumber: 20, Method: "MethodB"},
	}

	if !reflect.DeepEqual(apiErr.ErrorStruct.Details, expectedDetails) {
		t.Errorf("Expected error.details %+v, got %+v", expectedDetails, apiErr.ErrorStruct.Details)
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

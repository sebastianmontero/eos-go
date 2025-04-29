package eos

import (
	"reflect"
	"testing"
)

// compareAPIError asserts that actual and expected APIError fields match, and reports errors via t.
func compareAPIError(t *testing.T, actual, expected APIError) {
	t.Helper()
	if actual.Code != expected.Code {
		t.Errorf("Expected code %d, got %d", expected.Code, actual.Code)
	}
	if actual.Message != expected.Message {
		t.Errorf("Expected message '%s', got '%s'", expected.Message, actual.Message)
	}
	if actual.ErrorStruct.Code != expected.ErrorStruct.Code {
		t.Errorf("Expected error.code %d, got %d", expected.ErrorStruct.Code, actual.ErrorStruct.Code)
	}
	if actual.ErrorStruct.Name != expected.ErrorStruct.Name {
		t.Errorf("Expected error.name '%s', got '%s'", expected.ErrorStruct.Name, actual.ErrorStruct.Name)
	}
	if actual.ErrorStruct.What != expected.ErrorStruct.What {
		t.Errorf("Expected error.what '%s', got '%s'", expected.ErrorStruct.What, actual.ErrorStruct.What)
	}
	if !reflect.DeepEqual(actual.ErrorStruct.Details, expected.ErrorStruct.Details) {
		t.Errorf("Expected error.details %+v, got %+v", expected.ErrorStruct.Details, actual.ErrorStruct.Details)
	}
}


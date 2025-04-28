package eos

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sebastianmontero/eos-go/eoserr"
)

// APIError represents the errors as reported by the server
type APIError struct {
	Code        int    `json:"code,omitempty"` // http code
	Message     string `json:"message"`
	ErrorStruct apiErrorStruct `json:"error"`
}

type apiErrorStruct struct {
	Code    int              `json:"code"`
	Name    string           `json:"name"`
	What    string           `json:"what"`
	Details []APIErrorDetail `json:"details"`
}

func (a *APIError) UnmarshalJSON(data []byte) error {
	type Alias APIError // to avoid recursion
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	var alias Alias
	if err := dec.Decode(&alias); err != nil {
		return err
	}
	// Validate required fields (Message, ErrorStruct fields)
	if alias.Message == "" {
		return fmt.Errorf("missing required field: message")
	}
	if alias.ErrorStruct.Code == 0 && alias.ErrorStruct.Name == "" && alias.ErrorStruct.What == "" && len(alias.ErrorStruct.Details) == 0 {
		return fmt.Errorf("missing required field: error")
	}
	// Validate ErrorStruct fields
	if alias.ErrorStruct.Name == "" {
		return fmt.Errorf("missing required field: error.name")
	}
	if alias.ErrorStruct.What == "" {
		return fmt.Errorf("missing required field: error.what")
	}
	// You may want to check Details, but empty details may be valid
	*a = APIError(alias)
	return nil
}

func (e *apiErrorStruct) UnmarshalJSON(data []byte) error {
	type Alias apiErrorStruct
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	var alias Alias
	if err := dec.Decode(&alias); err != nil {
		return err
	}
	if alias.Name == "" {
		return fmt.Errorf("missing required field: name")
	}
	if alias.What == "" {
		return fmt.Errorf("missing required field: what")
	}
	*e = apiErrorStruct(alias)
	return nil
}

func NewAPIError(httpCode int, msg string, e eoserr.Error) *APIError {
	newError := &APIError{
		Code:    httpCode,
		Message: msg,
	}
	newError.ErrorStruct.Code = e.Code
	newError.ErrorStruct.Name = e.Name
	newError.ErrorStruct.What = msg
	newError.ErrorStruct.Details = []APIErrorDetail{
		APIErrorDetail{
			File:       "",
			LineNumber: 0,
			Message:    msg,
			Method:     e.Name,
		},
	}

	return newError
}

type APIErrorDetail struct {
	Message    string `json:"message"`
	File       string `json:"file"`
	LineNumber int    `json:"line_number"`
	Method     string `json:"method"`
}

func (e APIError) Error() string {
	msg := e.Message
	msg = fmt.Sprintf("%s: %s", msg, e.ErrorStruct.What)

	for _, detail := range e.ErrorStruct.Details {
		msg = fmt.Sprintf("%s: %s", msg, detail.Message)
	}

	return msg
}

// IsUnknowKeyError determines if the APIError is a 500 error
// with an `unknown key` message in at least one of the detail element.
// Some endpoint like `/v1/chain/get_account` returns a body in
// the form:
//
// ```
//
//	 {
//	 	"code": 500,
//	 	"message": "Internal Service Error",
//	 	"error": {
//	 		"code": 0,
//	 		"name": "exception",
//	 		"what": "unspecified",
//	 		"details": [
//			 		{
//			 			"message": "unknown key (<... redacted ...>): (0 eos.rex)",
//			 			"file": "http_plugin.cpp",
//			 			"line_number": 589,
//			 			"method": "handle_exception"
//			 		}
//	 		]
//	 	}
//	 }
//
// ```
//
// This will check if root code is a 500, that inner error code is 0 and there is
// a detail message starting with prefix `"unknown key"`.
func (e APIError) IsUnknownKeyError() bool {
	return e.Code == 500 &&
		e.ErrorStruct.Code == 0 &&
		e.hasDetailMessagePrefix("unknown key")
}

func (e APIError) IsContractError() bool {
	return e.errorNameIs("eosio_assert_message_exception")
}

func (e APIError) hasDetailMessagePrefix(prefix string) bool {
	for _, detail := range e.ErrorStruct.Details {
		if strings.HasPrefix(detail.Message, prefix) {
			return true
		}
	}

	return false
}

func (e APIError) errorNameIs(name string) bool {
	return strings.Contains(strings.ToLower(e.ErrorStruct.Name), strings.ToLower(name))
}

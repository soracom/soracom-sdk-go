package soracom

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// APIError represents an error ocurred while calling API
type APIError struct {
	HTTPStatusCode int
	ErrorCode      string
	Message        string
}

type apiErrorResponse struct {
	ErrorCode   string `json:"code"`
	Message     string `json:"message"`
	MessageArgs string `json:"messageArgs"`
}

func parseAPIErrorResponse(resp *http.Response) *apiErrorResponse {
	var aer apiErrorResponse
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&aer)
	return &aer
}

// NewAPIError creates an instance of APIError from http.Response
func NewAPIError(resp *http.Response) *APIError {
	var errorCode, message string
	ct := resp.Header.Get("Content-Type")

	if strings.Index(ct, "text/plain") == 0 {
		errorCode = "UNK0001"
		message = readAll(resp.Body)
	} else if strings.Index(ct, "application/json") == 0 {
		if resp.StatusCode >= http.StatusBadRequest &&
			resp.StatusCode < http.StatusInternalServerError {
			aer := parseAPIErrorResponse(resp)
			errorCode = aer.ErrorCode
			message = fmt.Sprintf(aer.Message, aer.MessageArgs)
		} else {
			errorCode = ""
			message = readAll(resp.Body)
		}
	} else {
		errorCode = "INT0001"
		message = "Content-Type: " + ct + " is not supported"
	}
	return &APIError{
		HTTPStatusCode: resp.StatusCode,
		ErrorCode:      errorCode,
		Message:        message,
	}
}

func (ae *APIError) Error() string {
	return ae.Message
}

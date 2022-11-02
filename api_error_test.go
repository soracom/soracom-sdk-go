package soracom

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestTextPlainApiError(t *testing.T) {
	r := io.NopCloser(strings.NewReader("test"))
	h := http.Header{}
	h.Set("Content-Type", "text/plain")
	res := &http.Response{
		StatusCode: 404,
		Body:       r,
		Header:     h,
	}

	apiError := NewAPIError(res)
	if apiError.HTTPStatusCode != 404 {
		t.Fatalf("wrong http status code: %v", apiError.HTTPStatusCode)
	}

	if apiError.ErrorCode != "UNK0001" {
		t.Fatalf("wrong error code: %v", apiError.ErrorCode)
	}
}

func TestJsonClientApiError(t *testing.T) {
	json := `{"code":"SEM0095"}`
	r := io.NopCloser(bytes.NewReader([]byte(json)))
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	res := &http.Response{
		StatusCode: 404,
		Body:       r,
		Header:     h,
	}

	apiError := NewAPIError(res)
	if apiError.HTTPStatusCode != 404 {
		t.Fatalf("wrong http status code: %v", apiError.HTTPStatusCode)
	}

	if apiError.ErrorCode != "SEM0095" {
		t.Fatalf("wrong error code: %v", apiError.ErrorCode)
	}
}

func TestJsonServerApiError(t *testing.T) {
	json := `{}`
	r := io.NopCloser(bytes.NewReader([]byte(json)))
	h := http.Header{}
	h.Set("Content-Type", "application/json")
	res := &http.Response{
		StatusCode: 500,
		Body:       r,
		Header:     h,
	}

	apiError := NewAPIError(res)
	if apiError.HTTPStatusCode != 500 {
		t.Fatalf("wrong http status code: %v", apiError.HTTPStatusCode)
	}

	if apiError.ErrorCode != "" {
		t.Fatalf("wrong error code: %v", apiError.ErrorCode)
	}
}

func TestNotSupportedContentTypeApiError(t *testing.T) {
	h := http.Header{}
	h.Set("Content-Type", "application/pdf")
	res := &http.Response{
		StatusCode: 415,
		Header:     h,
	}

	apiError := NewAPIError(res)
	if apiError.HTTPStatusCode != 415 {
		t.Fatalf("wrong http status code: %v", apiError.HTTPStatusCode)
	}

	if apiError.ErrorCode != "INT0001" {
		t.Fatalf("wrong error code: %v", apiError.ErrorCode)
	}
}

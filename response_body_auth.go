package soracom

import (
	"encoding/json"
	"net/http"
)

// ResponseBodyAuth contains all values returned from /auth API
type ResponseBodyAuth struct {
	APIKey     string `json:"apiKey"`
	OperatorID string `json:"operatorId"`
	Token      string `json:"token"`
}

func parseAuthResponse(resp *http.Response) *ResponseBodyAuth {
	var arb ResponseBodyAuth
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&arb)
	return &arb
}

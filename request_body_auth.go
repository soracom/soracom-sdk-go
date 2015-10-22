package soracom

import "encoding/json"

// RequestBodyAuth contains parameters for /auth API
type RequestBodyAuth struct {
	Email               string `json:"email"`
	Password            string `json:"password"`
	TokenTimeoutSeconds int    `json:"tokenTimeoutSeconds"`
}

// JSON converts RequestBodyAuth struct to JSON string
func (arb *RequestBodyAuth) JSON() string {
	bodyBytes, err := json.Marshal(arb)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

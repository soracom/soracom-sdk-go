package soracom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type SoracomApiClient struct {
	httpClient *http.Client
	Token      string
}

type AuthRequestBody struct {
	Email               string `json:"email"`
	Password            string `json:"password"`
	TokenTimeoutSeconds int    `json:"tokenTimeoutSeconds"`
}

type AuthResponseBody struct {
	ApiKey     string
	OperatorId string
	Token      string
}

type Subscriber struct {
	Apn        string            `json:"apn"`
	Expiry     *time.Time        `json:"expiredAt"`
	GroupId    string            `json:"groupId"`
	Imsi       string            `json:"imsi"`
	IpAddress  string            `json:"ipAddress"`
	Msisdn     string            `json:"msisdn"`
	Status     string            `json:"status"`
	SpeedClass string            `json:"type"`
	Tags       map[string]string `json:"tags"`
}

type ApiError struct {
	ErrorCode string
	Message   string
}

type ApiErrorResponse struct {
	ErrorCode   string `json:"code"`
	Message     string `json:"message"`
	MessageArgs string `json:"messageArgs"`
}

type apiParams struct {
	method      string
	path        string
	contentType string
	body        string
}

func readAll(r io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.String()
}

func NewApiError(resp *http.Response) *ApiError {
	var errorCode, message string
	ct := resp.Header.Get("Content-Type")

	if strings.Index(ct, "text/plain") == 0 {
		errorCode = "UNK0001"
		message = readAll(resp.Body)
	} else if strings.Index(ct, "application/json") == 0 {
		if resp.StatusCode >= http.StatusBadRequest &&
			resp.StatusCode < http.StatusInternalServerError {
			dec := json.NewDecoder(resp.Body)
			var aer ApiErrorResponse
			dec.Decode(&aer)
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
	return &ApiError{
		ErrorCode: errorCode,
		Message:   message,
	}
}

func (ae *ApiError) Error() string {
	return ae.Message
}

func NewClient() *SoracomApiClient {
	hc := new(http.Client)
	return &SoracomApiClient{
		httpClient: hc,
		Token:      "",
	}
}

func (sac *SoracomApiClient) getEndpointBase() string {
	return "http://sora-prod.apigee.net"
}

func (sac *SoracomApiClient) callApi(params *apiParams) (*http.Response, error) {
	url := sac.getEndpointBase() + params.path
	req, err := http.NewRequest(params.method, url, strings.NewReader(params.body))
	if err != nil {
		return nil, err
	}

	if params.contentType != "" {
		req.Header.Set("Content-Type", params.contentType)
	}

	req.Header.Set("X-Soracom-Token", sac.Token)

	return sac.httpClient.Do(req)
}

func (sac *SoracomApiClient) Auth(email, password string) error {
	body := &AuthRequestBody{
		Email:               email,
		Password:            password,
		TokenTimeoutSeconds: 24 * 60 * 60,
	}
	bodyStr, err := json.Marshal(&body)
	if err != nil {
		return err
	}

	params := &apiParams{
		method:      "POST",
		path:        "/v1/operators/auth",
		contentType: "application/json",
		body:        string(bodyStr),
	}

	resp, err := sac.callApi(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return NewApiError(resp)
	}

	var respBody AuthResponseBody
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&respBody)
	sac.Token = respBody.Token
	return nil
}

func (sac *SoracomApiClient) ListSubscribers() ([]Subscriber, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/subscribers",
	}
	resp, err := sac.callApi(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, NewApiError(resp)
	}

	var subscribers []Subscriber
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&subscribers)

	return subscribers, nil
}

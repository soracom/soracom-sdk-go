package soracom

import (
	"net/http"
	"strings"
)

// APIClient provides an access to SORACOM REST API
type APIClient struct {
	httpClient *http.Client
	APIKey     string
	Token      string
	endpoint   string
}

// APIClientOptions holds options for creating an APIClient
type APIClientOptions struct {
	Endpoint string
}

// NewAPIClient creates an instance of APIClient
func NewAPIClient(options *APIClientOptions) *APIClient {
	hc := new(http.Client)

	var endpoint = "https://api.soracom.io"
	if options != nil && options.Endpoint != "" {
		endpoint = options.Endpoint
	}

	return &APIClient{
		httpClient: hc,
		APIKey:     "",
		Token:      "",
		endpoint:   endpoint,
	}
}

type apiParams struct {
	method      string
	path        string
	query       string
	contentType string
	body        string
}

func (ac *APIClient) callAPI(params *apiParams) (*http.Response, error) {
	url := ac.endpoint + params.path
	if params.query != "" {
		url += "?" + params.query
	}
	//fmt.Printf("url == %v\n", url)
	req, err := http.NewRequest(params.method, url, strings.NewReader(params.body))
	if err != nil {
		return nil, err
	}

	if params.contentType != "" {
		req.Header.Set("Content-Type", params.contentType)
	}

	req.Header.Set("X-Soracom-API-Key", ac.APIKey)
	req.Header.Set("X-Soracom-Token", ac.Token)

	res, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode >= http.StatusBadRequest {
		defer res.Body.Close()
		return nil, NewAPIError(res)
	}

	return res, nil
}

// Auth does the authentication process. Gets an API key and an API Token
func (ac *APIClient) Auth(email, password string) error {
	body := &RequestBodyAuth{
		Email:               email,
		Password:            password,
		TokenTimeoutSeconds: 24 * 60 * 60,
	}
	params := &apiParams{
		method:      "POST",
		path:        "/v1/auth",
		contentType: "application/json",
		body:        body.JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody := parseAuthResponse(resp)
	ac.APIKey = respBody.APIKey
	ac.Token = respBody.Token

	return nil
}

// ListSubscribers lists subscribers for the operator
func (ac *APIClient) ListSubscribers(options *ListSubscribersOptions) ([]Subscriber, *PaginationKeys, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/subscribers",
	}
	if options != nil {
		params.query = options.String()
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	subscribers, paginationKeys, err := parseListSubscribersResponse(resp)
	if err != nil {
		return nil, nil, err
	}

	return subscribers, paginationKeys, nil
}

// GetSubscriber gets information about a subscriber specifed by imsi.
func (ac *APIClient) GetSubscriber(imsi string) (*Subscriber, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/subscribers/" + imsi,
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseGetSubscriberResponse(resp)

	return subscriber, nil
}

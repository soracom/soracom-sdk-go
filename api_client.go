package soracom

import (
	"fmt"
	"net/http"
	"strings"
	"time"
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
	body := &AuthRequest{
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

// RegisterSubscriber registers a subscriber.
func (ac *APIClient) RegisterSubscriber(imsi string, regOptions RegisterSubscriberOptions) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/register",
		contentType: "application/json",
		body:        regOptions.JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
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

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// UpdateSubscriberSpeedClass updates speed class of a subscriber.
func (ac *APIClient) UpdateSubscriberSpeedClass(imsi, speedClass string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/update_speed_class",
		contentType: "application/json",
		body:        (&updateSpeedClassRequest{SpeedClass: speedClass}).JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// ActivateSubscriber activates a subscriber.
func (ac *APIClient) ActivateSubscriber(imsi string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/activate",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// DeactivateSubscriber deactivates a subscriber.
func (ac *APIClient) DeactivateSubscriber(imsi string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/deactivate",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// TerminateSubscriber terminates a subscriber.
func (ac *APIClient) TerminateSubscriber(imsi string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/terminate",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// EnableSubscriberTermination enables termination of a subscriber.
func (ac *APIClient) EnableSubscriberTermination(imsi string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/enable_termination",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// DisableSubscriberTermination disables termination of a subscriber.
func (ac *APIClient) DisableSubscriberTermination(imsi string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/disable_termination",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// SetSubscriberExpiryTime sets expiration time of a subscriber.
func (ac *APIClient) SetSubscriberExpiryTime(imsi string, expiryTime time.Time) (*Subscriber, error) {
	ts := &Timestamp{Time: expiryTime}
	req := &setExpiryTimeRequest{
		ExpiryTime: fmt.Sprint(ts.UnixMilli()),
	}
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/set_expiry_time",
		contentType: "application/json",
		body:        req.JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// UnsetSubscriberExpiryTime unsets expiration time of a subscriber.
func (ac *APIClient) UnsetSubscriberExpiryTime(imsi string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/unset_expiry_time",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// SetSubscriberGroup sets a group of a subscriber.
func (ac *APIClient) SetSubscriberGroup(imsi, groupID string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/set_group",
		contentType: "application/json",
		body:        (&setSubscriberGroupRequest{GroupID: groupID}).JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// UnsetSubscriberGroup unsets group of a subscriber.
func (ac *APIClient) UnsetSubscriberGroup(imsi string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscribers/" + imsi + "/unset_group",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// PutSubscriberTags puts tags on a subscriber
func (ac *APIClient) PutSubscriberTags(imsi string, tags []Tag) (*Subscriber, error) {
	params := &apiParams{
		method:      "PUT",
		path:        "/v1/subscribers/" + imsi + "/tags",
		contentType: "application/json",
		body:        tagsToJSON(tags),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// DeleteSubscriberTag deletes a tag on a subscriber
func (ac *APIClient) DeleteSubscriberTag(imsi string, tagName string) error {
	params := &apiParams{
		method: "DELETE",
		path:   "/v1/subscribers/" + imsi + "/tags/" + percentEncoding(tagName),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

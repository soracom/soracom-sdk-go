package soracom

import (
	"fmt"
	"net/http"
	"strings"
	"time"
)

// MetadataClient provides an access to SORACOM Metadata Service APIs
type MetadataClient struct {
	httpClient *http.Client
	endpoint   string
	verbose    bool
}

// MetadataClientOptions holds options for creating an MetadataClient
type MetadataClientOptions struct {
	Endpoint string
}

// NewMetadataClient creates an instance of MetadataClient
func NewMetadataClient(options *MetadataClientOptions) *MetadataClient {
	hc := http.DefaultClient

	var endpoint = "http://metadata.soracom.io"
	if options != nil && options.Endpoint != "" {
		endpoint = options.Endpoint
	}

	return &MetadataClient{
		httpClient: hc,
		endpoint:   endpoint,
	}
}

// SetVerbose sets if verbose output is enabled or not
func (mc *MetadataClient) SetVerbose(verbose bool) {
	mc.verbose = verbose
}

func (mc *MetadataClient) callAPI(params *apiParams) (*http.Response, error) {
	url := mc.endpoint + params.path
	if params.query != "" {
		url += "?" + params.query
	}

	req, err := http.NewRequest(params.method, url, strings.NewReader(params.body))
	if err != nil {
		return nil, err
	}

	if params.contentType != "" {
		req.Header.Set("Content-Type", params.contentType)
	}

	if mc.verbose {
		dumpHTTPRequest(req)
	}

	res, err := mc.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if mc.verbose {
		dumpHTTPResponse(res)
		fmt.Println("==========")
	}

	if res.StatusCode >= http.StatusBadRequest {
		defer res.Body.Close()
		return nil, NewAPIError(res)
	}

	return res, nil
}

// GetSubscriber gets metadata for the calling subscriber
func (mc *MetadataClient) GetSubscriber() (*Subscriber, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/subscriber",
	}
	resp, err := mc.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// UpdateSpeedClass updates speed class of the calling subscriber.
func (mc *MetadataClient) UpdateSpeedClass(speedClass string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscriber/update_speed_class",
		contentType: "application/json",
		body:        (&updateSpeedClassRequest{SpeedClass: speedClass}).JSON(),
	}

	resp, err := mc.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// EnableTermination enables termination of the calling subscriber.
func (mc *MetadataClient) EnableTermination() (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscriber/enable_termination",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := mc.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// DisableTermination disables termination of the calling subscriber.
func (mc *MetadataClient) DisableTermination() (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscriber/disable_termination",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := mc.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// SetExpiredAt sets expiration time of the calling subscriber.
func (mc *MetadataClient) SetExpiredAt(expiryTime time.Time) (*Subscriber, error) {
	ts := &TimestampMilli{Time: expiryTime}
	req := &setExpiredAtRequest{
		ExpiredAt: fmt.Sprint(ts.UnixMilli()),
	}
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscriber/set_expiry_time",
		contentType: "application/json",
		body:        req.JSON(),
	}

	resp, err := mc.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// UnsetExpiredAt unsets expiration time of the calling subscriber.
func (mc *MetadataClient) UnsetExpiredAt() (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscriber/unset_expiry_time",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := mc.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// SetGroup sets a group of the calling subscriber.
func (mc *MetadataClient) SetGroup(groupID string) (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscriber/set_group",
		contentType: "application/json",
		body:        (&setSubscriberGroupRequest{GroupID: groupID}).JSON(),
	}

	resp, err := mc.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// UnsetGroup unsets group of the calling subscriber.
func (mc *MetadataClient) UnsetGroup() (*Subscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/subscriber/unset_group",
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := mc.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// PutTags puts tags on the calling subscriber
func (mc *MetadataClient) PutTags(tags []Tag) (*Subscriber, error) {
	params := &apiParams{
		method:      "PUT",
		path:        "/v1/subscriber/tags",
		contentType: "application/json",
		body:        tagsToJSON(tags),
	}

	resp, err := mc.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	subscriber := parseSubscriber(resp)

	return subscriber, nil
}

// DeleteTag deletes a tag on the calling subscriber
func (mc *MetadataClient) DeleteTag(tagName string) error {
	params := &apiParams{
		method: "DELETE",
		path:   "/v1/subscriber/tags/" + percentEncoding(tagName),
	}
	resp, err := mc.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetUserdata gets userdata for the calling subscriber's group
func (mc *MetadataClient) GetUserdata() (string, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/userdata",
	}
	resp, err := mc.callAPI(params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	return readAll(resp.Body), nil
}

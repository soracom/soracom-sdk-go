package soracom

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// APIClient provides an access to SORACOM REST API
type APIClient struct {
	httpClient *http.Client
	APIKey     string
	Token      string
	OperatorID string
	endpoint   string
	verbose    bool
}

// APIClientOptions holds options for creating an APIClient
type APIClientOptions struct {
	Endpoint string
}

// NewAPIClient creates an instance of APIClient
func NewAPIClient(options *APIClientOptions) *APIClient {
	hc := http.DefaultClient

	var endpoint = "https://api.soracom.io"
	if options != nil && options.Endpoint != "" {
		endpoint = options.Endpoint
	}

	return &APIClient{
		httpClient: hc,
		APIKey:     "",
		Token:      "",
		OperatorID: "",
		endpoint:   endpoint,
		verbose:    false,
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

	if ac.verbose {
		dumpHTTPRequest(req)
	}

	res, err := ac.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if ac.verbose {
		dumpHTTPResponse(res)
		fmt.Println("==========")
	}

	if res.StatusCode >= http.StatusBadRequest {
		defer res.Body.Close()
		return nil, NewAPIError(res)
	}

	return res, nil
}

// SetVerbose sets if verbose output is enabled or not
func (ac *APIClient) SetVerbose(verbose bool) {
	ac.verbose = verbose
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
	ac.OperatorID = respBody.OperatorID

	return nil
}

// GenerateAPIToken generates an API token
func (ac *APIClient) GenerateAPIToken(timeout int) (string, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/operators/" + ac.OperatorID + "/token",
		contentType: "application/json",
		body:        (&generateAPITokenRequest{Timeout: timeout}).JSON(),
	}
	resp, err := ac.callAPI(params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody := parseGenerateAPITokenResponse(resp)
	return respBody.Token, nil
}

// UpdatePassword updates operator's password
func (ac *APIClient) UpdatePassword(currentPassword, newPassword string) error {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/operators/" + ac.OperatorID + "/password",
		contentType: "application/json",
		body:        (&updatePasswordRequest{CurrentPassword: currentPassword, NewPassword: newPassword}).JSON(),
	}
	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetSupportToken retrieves a token for accessing to the support site
func (ac *APIClient) GetSupportToken() (string, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/operators/" + ac.OperatorID + "/support/token",
		contentType: "application/json",
		body:        "{}",
	}
	resp, err := ac.callAPI(params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody := parseGetSupportTokenResponse(resp)
	return respBody.Token, nil
}

// CreateOperator sends a request to create an operator with the specified email & password.
func (ac *APIClient) CreateOperator(email, password string) error {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/operators",
		contentType: "application/json",
		body:        (&createOperatorRequest{Email: email, Password: password}).JSON(),
	}
	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// VerifyOperator sends a token to complete an operator creation process.
func (ac *APIClient) VerifyOperator(token string) error {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/operators/verify",
		contentType: "application/json",
		body:        (&verifyOperatorRequest{Token: token}).JSON(),
	}
	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetOperator gets information about an operator specifed by operatorID.
func (ac *APIClient) GetOperator(operatorID string) (*Operator, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/operators/" + operatorID,
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	operator := parseOperator(resp)

	return operator, nil
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
	ts := &TimestampMilli{Time: expiryTime}
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

// GetAirStats gets stats of Air for a subscriber for a specified period
func (ac *APIClient) GetAirStats(imsi string, from, to time.Time, period StatsPeriod) ([]AirStats, error) {
	params := &apiParams{
		method: "GET",
		path:   fmt.Sprintf("/v1/stats/air/subscribers/%s?from=%d&to=%d&period=%s", imsi, from.Unix(), to.Unix(), period.String()),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	airStats := parseAirStats(resp)

	return airStats, nil
}

// GetBeamStats gets stats of Beam for a subscriber for a specified period
func (ac *APIClient) GetBeamStats(imsi string, from, to time.Time, period StatsPeriod) ([]BeamStats, error) {
	params := &apiParams{
		method: "GET",
		path:   fmt.Sprintf("/v1/stats/beam/subscribers/%s?from=%d&to=%d&period=%s", imsi, from.Unix(), to.Unix(), period.String()),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	beamStats := parseBeamStats(resp)

	return beamStats, nil
}

// ExportAirStats gets a URL to download a CSV file which contains stats of all Air SIMs for the operator for a specified period
func (ac *APIClient) ExportAirStats(from, to time.Time, period StatsPeriod) (*url.URL, error) {
	params := &apiParams{
		method:      "POST",
		path:        fmt.Sprintf("/v1/stats/air/operators/%s/export", ac.OperatorID),
		contentType: "application/json",
		body: (&exportAirStatsRequest{
			From:   from.Unix(),
			To:     to.Unix(),
			Period: period.String(),
		}).JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody := parseExportAirStatsResponse(resp)
	url, err := url.Parse(respBody.URL)
	if err != nil {
		return nil, err
	}

	return url, nil
}

// ExportBeamStats gets a URL to download a CSV file which contains all stats of Beam for the operator for a specified period
func (ac *APIClient) ExportBeamStats(from, to time.Time, period StatsPeriod) (*url.URL, error) {
	params := &apiParams{
		method:      "POST",
		path:        fmt.Sprintf("/v1/stats/beam/operators/%s/export", ac.OperatorID),
		contentType: "application/json",
		body: (&exportBeamStatsRequest{
			From:   from.Unix(),
			To:     to.Unix(),
			Period: period.String(),
		}).JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody := parseExportBeamStatsResponse(resp)
	url, err := url.Parse(respBody.URL)
	if err != nil {
		return nil, err
	}

	return url, nil
}

// ListGroups lists groups for the operator
func (ac *APIClient) ListGroups(options *ListGroupsOptions) ([]Group, *PaginationKeys, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/groups",
	}
	if options != nil {
		params.query = options.String()
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	groups, paginationKeys, err := parseListGroupsResponse(resp)
	if err != nil {
		return nil, nil, err
	}

	return groups, paginationKeys, nil
}

// CreateGroup creates a group
func (ac *APIClient) CreateGroup(tags Tags) (*Group, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/groups",
		contentType: "application/json",
		body: (&createGroupRequest{
			Tags: tags,
		}).JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	group := parseGroup(resp)

	return group, nil
}

// CreateGroupWithName creates a group with name
func (ac *APIClient) CreateGroupWithName(name string) (*Group, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/groups",
		contentType: "application/json",
		body: (&createGroupRequest{
			Tags: Tags{"name": name},
		}).JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	group := parseGroup(resp)

	return group, nil
}

// DeleteGroup deletes a group
func (ac *APIClient) DeleteGroup(groupID string) error {
	params := &apiParams{
		method:      "DELETE",
		path:        "/v1/groups/" + groupID,
		contentType: "application/json",
		body:        "{}",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetGroup gets detailed info about a group
func (ac *APIClient) GetGroup(groupID string) (*Group, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/groups/" + groupID,
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	group := parseGroup(resp)

	return group, nil
}

// ListSubscribersInGroup lists subscribers in a group
func (ac *APIClient) ListSubscribersInGroup(groupID string, options *ListSubscribersInGroupOptions) ([]Subscriber, *PaginationKeys, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/groups/" + groupID + "/subscribers",
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

// UpdateGroupConfigurations updates configurations for a group
func (ac *APIClient) UpdateGroupConfigurations(groupID, namespace string, configurations []GroupConfig) (*Group, error) {
	params := &apiParams{
		method:      "PUT",
		path:        "/v1/groups/" + groupID + "/configuration/" + namespace,
		contentType: "application/json",
		body:        toJSON(configurations),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	group := parseGroup(resp)

	return group, nil
}

// UpdateAirConfig updates SORACOM Air configurations for a group
func (ac *APIClient) UpdateAirConfig(groupID string, airConfig *AirConfig) (*Group, error) {
	params := &apiParams{
		method:      "PUT",
		path:        "/v1/groups/" + groupID + "/configuration/SoracomAir",
		contentType: "application/json",
		body:        airConfig.JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	group := parseGroup(resp)

	return group, nil
}

// UpdateBeamTCPConfig updates SORACOM Beam configurations for a group
func (ac *APIClient) UpdateBeamTCPConfig(groupID, entryPoint string, beamTCPConfig *BeamTCPConfig) (*Group, error) {
	params := &apiParams{
		method:      "PUT",
		path:        "/v1/groups/" + groupID + "/configuration/SoracomBeam",
		contentType: "application/json",
		body: toJSON([]GroupConfig{
			GroupConfig{Key: entryPoint, Value: beamTCPConfig},
		}),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	group := parseGroup(resp)

	return group, nil
}

// DeleteGroupConfiguration deletes a configuration for a group
func (ac *APIClient) DeleteGroupConfiguration(groupID, namespace, name string) (*Group, error) {
	params := &apiParams{
		method: "DELETE",
		path:   "/v1/groups/" + groupID + "/configuration/" + namespace + "/" + percentEncoding(name),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	group := parseGroup(resp)

	return group, nil
}

// UpdateGroupTags updates tags a group
func (ac *APIClient) UpdateGroupTags(groupID string, tags []Tag) (*Group, error) {
	params := &apiParams{
		method:      "PUT",
		path:        "/v1/groups/" + groupID + "/tags",
		contentType: "application/json",
		body:        toJSON(tags),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	group := parseGroup(resp)

	return group, nil
}

// DeleteGroupTag deletes a tag for a group
func (ac *APIClient) DeleteGroupTag(groupID, tagName string) error {
	params := &apiParams{
		method: "DELETE",
		path:   "/v1/groups/" + groupID + "/tags/" + percentEncoding(tagName),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// ListEventHandlers lists event handlers for the operator
func (ac *APIClient) ListEventHandlers(options *ListEventHandlersOptions) ([]EventHandler, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/event_handlers",
	}
	if options != nil {
		params.query = options.String()
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	eventHandlers, err := parseListEventHandlersResponse(resp)
	if err != nil {
		return nil, err
	}

	return eventHandlers, nil
}

// CreateEventHandler creates an event handler
func (ac *APIClient) CreateEventHandler(options *CreateEventHandlerOptions) error {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/event_handlers",
		contentType: "application/json",
		body:        options.JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	/*
		eventHandler, err := parseEventHandler(resp)
		if err != nil {
			return nil, err
		}
		return eventHandler, nil
	*/

	return nil
}

// ListEventHandlersForSubscriber creates an event handler with the specified options
func (ac *APIClient) ListEventHandlersForSubscriber(imsi string) ([]EventHandler, error) {
	params := &apiParams{
		method:      "GET",
		path:        "/v1/event_handlers/subscribers/" + imsi,
		contentType: "application/json",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	eventHandlers, err := parseListEventHandlersResponse(resp)
	if err != nil {
		return nil, err
	}

	return eventHandlers, nil
}

// DeleteEventHandler deletes the specified event handler
func (ac *APIClient) DeleteEventHandler(handlerID string) error {
	params := &apiParams{
		method: "DELETE",
		path:   "/v1/event_handlers/" + handlerID,
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetEventHandler gets the specified event handler
func (ac *APIClient) GetEventHandler(handlerID string) (*EventHandler, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/event_handlers/" + handlerID,
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	eventHandler, err := parseEventHandler(resp)
	if err != nil {
		return nil, err
	}
	return eventHandler, nil
}

// UpdateEventHandler updates the specified event handler
func (ac *APIClient) UpdateEventHandler(eh *EventHandler) error {
	params := &apiParams{
		method:      "PUT",
		path:        "/v1/event_handlers/" + eh.HandlerID,
		contentType: "application/json",
		body:        eh.JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// RegisterPaymentMethodWebPay registers the specified WebPay information as an active payment method
func (ac *APIClient) RegisterPaymentMethodWebPay(wp *PaymentMethodInfoWebPay) error {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/payment_methods/webpay",
		contentType: "application/json",
		body:        wp.JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetSignupToken retrieves token to complete signup (sandbox environment only)
func (ac *APIClient) GetSignupToken(email, authKeyID, authKey string) (string, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/sandbox/operators/token/" + email,
		contentType: "application/json",
		body: (&AuthKey{
			AuthKeyID:     authKeyID,
			AuthKeySecret: authKey,
		}).JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	token, err := parseSignupToken(resp)
	if err != nil {
		return "", err
	}

	return token, nil
}

// CreateSubscriber sends a request to create a brand-new subscriber
func (ac *APIClient) CreateSubscriber() (*CreatedSubscriber, error) {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/sandbox/subscribers/create",
		contentType: "application/json",
		body:        "",
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	cs, err := parseCreatedSubscriber(resp)
	if err != nil {
		return nil, err
	}

	return cs, nil
}

// InsertAirStats inserts a set of data communication stats with a timestamp for a subscriber
func (ac *APIClient) InsertAirStats(imsi string, stats AirStats) error {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/sandbox/stats/air/subscribers/" + imsi,
		contentType: "application/json",
		body:        stats.JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// InsertBeamStats inserts a set of beam stats with a timestamp for a subscriber
func (ac *APIClient) InsertBeamStats(imsi string, stats BeamStats) error {
	params := &apiParams{
		method:      "POST",
		path:        "/v1/sandbox/stats/beam/subscribers/" + imsi,
		contentType: "application/json",
		body:        stats.JSON(),
	}

	resp, err := ac.callAPI(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// Package: SORACOM SDK for Go.

package soracom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Tag is a pair of Name and Value.
type Tag struct {
	TagName  string `json:"tagName"`
	TagValue string `json:"tagValue"`
}

// Tags is a map of tag name and tag value
type Tags map[string]string

// Properties is a map of property name and propaty value
type Properties map[string]string

// TimestampMilli is ...
type TimestampMilli struct {
	time.Time
}

// MarshalJSON is ...
func (t *TimestampMilli) MarshalJSON() ([]byte, error) {
	ts := t.Time.UnixNano() / (1000 * 1000)
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

// UnmarshalJSON is ...
func (t *TimestampMilli) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	ms := int64(ts)
	s := ms / 1000
	ns := (ms % 1000) * 1000 * 1000
	t.Time = time.Unix(s, ns)

	return nil
}

// UnixMilli returns t as a Unix time, the number of milliseconds elapsed since January 1, 1970 UTC.
func (t *TimestampMilli) UnixMilli() int64 {
	ns := t.Time.UnixNano()
	return ns / (1000 * 1000)
}

// AuthRequest contains parameters for /auth API
type AuthRequest struct {
	Email               string `json:"email"`
	Password            string `json:"password"`
	TokenTimeoutSeconds int    `json:"tokenTimeoutSeconds"`
}

// JSON returns JSON representing AuthRequest
func (ar *AuthRequest) JSON() string {
	bodyBytes, err := json.Marshal(ar)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

// AuthResponse contains all values returned from /auth API
type AuthResponse struct {
	APIKey     string `json:"apiKey"`
	OperatorID string `json:"operatorId"`
	Token      string `json:"token"`
}

func parseAuthResponse(resp *http.Response) *AuthResponse {
	var ar AuthResponse
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&ar)
	return &ar
}

type generateAPITokenRequest struct {
	Timeout int `json:"timeout_seconds"`
}

// JSON retunrs a JSON representing updateSpeedClassRequest object
func (r *generateAPITokenRequest) JSON() string {
	return toJSON(r)
}

// GenerateAPITokenResponse contains all values returned from /operators/{operator_id}/token API
type GenerateAPITokenResponse struct {
	Token string `json:"token"`
}

func parseGenerateAPITokenResponse(resp *http.Response) *GenerateAPITokenResponse {
	var r GenerateAPITokenResponse
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&r)
	return &r
}

type updatePasswordRequest struct {
	CurrentPassword string `json:"currentPassword"`
	NewPassword     string `json:"newPassword"`
}

func (r *updatePasswordRequest) JSON() string {
	return toJSON(r)
}

// GetSupportTokenResponse contains all values returned from /operators/{operator_id}/support/token API.
type GetSupportTokenResponse struct {
	Token string `json:"token"`
}

func parseGetSupportTokenResponse(resp *http.Response) *GetSupportTokenResponse {
	var r GetSupportTokenResponse
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&r)
	return &r
}

type createOperatorRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (r *createOperatorRequest) JSON() string {
	return toJSON(r)
}

type verifyOperatorRequest struct {
	Token string `json:"token"`
}

func (r *verifyOperatorRequest) JSON() string {
	return toJSON(r)
}

// Operator keeps information about an operator
type Operator struct {
	OperatorID     string     `json:"operatorId"`
	RootOperatorID *string    `json:"rootOperatorId"`
	Email          string     `json:"email"`
	Description    *string    `json:"description"`
	CreateDate     *time.Time `json:"createDate"`
	UpdateDate     *time.Time `json:"updateDate"`
}

func parseOperator(resp *http.Response) *Operator {
	var o Operator
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&o)
	return &o
}

// TagValueMatchMode is one of MatchModeUnspecified, MatchModeExact or MatchModePrefix
type TagValueMatchMode int

const (
	// MatchModeUnspecified is a value of TagValueMatchMode.
	// For list functions, they don't match tag values (i.e. list all items regardless of values of tags) if this value is specified.
	MatchModeUnspecified TagValueMatchMode = iota

	// MatchModeExact is a value of TagValueMatchMode.
	// For list functions, they do exact match for tag values if this value is specified.
	MatchModeExact

	// MatchModePrefix is a value of TagValueMatchMode.
	// For list functions, they do prefix match for tag values if this value is specified.
	MatchModePrefix
)

func (m TagValueMatchMode) String() string {
	switch m {
	case MatchModeExact:
		return "exact"
	case MatchModePrefix:
		return "prefix"
	}
	return ""
}

// Parse parses a provided string and returns TagValueMatchMode
func (m TagValueMatchMode) Parse(s string) TagValueMatchMode {
	switch s {
	case "exact":
		return MatchModeExact
	case "prefix":
		return MatchModePrefix
	default:
		return MatchModeUnspecified
	}
}

// ListSubscribersOptions holds options for APIClient.ListSubscribers()
type ListSubscribersOptions struct {
	TagName           string
	TagValue          string
	TagValueMatchMode TagValueMatchMode
	StatusFilter      string
	TypeFilter        string
	Limit             int
	LastEvaluatedKey  string
}

func (lso *ListSubscribersOptions) String() string {
	var s = make([]string, 0, 10)
	if lso.TagName != "" {
		s = append(s, "tag_name="+lso.TagName)
	}
	if lso.TagValue != "" {
		s = append(s, "tag_value="+lso.TagValue)
	}
	if lso.TagValueMatchMode != MatchModeUnspecified {
		s = append(s, "tag_value_match_mode="+lso.TagValueMatchMode.String())
	}
	if lso.StatusFilter != "" {
		s = append(s, "status_filter="+lso.StatusFilter)
	}
	if lso.TypeFilter != "" {
		s = append(s, "type_filter="+lso.TypeFilter)
	}
	if lso.Limit != 0 {
		s = append(s, "limit="+strconv.Itoa(lso.Limit))
	}
	if lso.LastEvaluatedKey != "" {
		s = append(s, "last_evaluated_key="+lso.LastEvaluatedKey)
	}
	return strings.Join(s, "&")
}

// RegisterSubscriberOptions keeps information for registering a subscriber
type RegisterSubscriberOptions struct {
	RegistrationSecret string `json:"registrationSecret"`
	GroupID            string `json:"groupId"`
	Tags               Tags   `json:"tags"`
}

// JSON retunrs a JSON representing RegisterSubscriberOptions object
func (rso *RegisterSubscriberOptions) JSON() string {
	return toJSON(rso)
}

// SessionStatus keeps information about a session
type SessionStatus struct {
	DNSServers    []string        `json:"dnsServers"`
	Imei          string          `json:"imei"`
	LastUpdatedAt *TimestampMilli `json:"lastUpdatedAt"`
	Location      *string         `json:"location"`
	Online        bool            `json:"online"`
	UEIPAddress   string          `json:"ueIpAddress"`
}

// Subscriber keeps information about a subscriber
type Subscriber struct {
	Apn                string          `json:"apn"`
	CreatedTime        *TimestampMilli `json:"createdTime"`
	ExpiryTime         *TimestampMilli `json:"expiryTime"`
	GroupID            *string         `json:"groupId"`
	Imsi               string          `json:"imsi"`
	IPAddress          *string         `json:"ipAddress"`
	LastModifiedTime   *TimestampMilli `json:"lastModifiedTime"`
	ModuleType         string          `json:"ModuleType"`
	Msisdn             string          `json:"msisdn"`
	OperatorID         string          `json:"operatorId"`
	Plan               int             `json:"plan"`
	SessionStatus      *SessionStatus  `json:"sessionStatus"`
	Status             string          `json:"status"`
	SpeedClass         string          `json:"speedClass"`
	Tags               Tags            `json:"tags"`
	TerminationEnabled bool            `json:"terminationEnabled"`
}

// PaginationKeys holds keys for pagination
type PaginationKeys struct {
	Prev string
	Next string
}

func parseLinkHeader(linkHeader string) *PaginationKeys {
	var pk *PaginationKeys
	if linkHeader != "" {
		pk = &PaginationKeys{}
		links := strings.Split(linkHeader, ",")
		for _, link := range links {
			s := strings.Split(link, ";")
			url, err := url.Parse(strings.Trim(s[0], "<>"))
			if err != nil {
				continue
			}
			lek := url.Query()["last_evaluated_key"][0]
			rel := strings.Split(s[1], "=")[1]
			if rel == "prev" {
				pk.Prev = lek
			} else if rel == "next" {
				pk.Next = lek
			}
		}
	}
	return pk
}

func parseListSubscribersResponse(resp *http.Response) ([]Subscriber, *PaginationKeys, error) {
	subs := make([]Subscriber, 0, 10)
	dec := json.NewDecoder(resp.Body)

	// read open bracket
	_, err := dec.Token()
	if err != nil {
		return nil, nil, err
	}

	for dec.More() {
		var s Subscriber
		err = dec.Decode(&s)
		if err != nil {
			continue
		}
		subs = append(subs, s)
	}

	// read close bracket
	_, err = dec.Token()
	if err != nil {
		return nil, nil, err
	}

	linkHeader := resp.Header.Get(http.CanonicalHeaderKey("Link"))
	pk := parseLinkHeader(linkHeader)

	return subs, pk, nil
}

func parseSubscriber(resp *http.Response) *Subscriber {
	var sub Subscriber
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&sub)
	return &sub
}

type updateSpeedClassRequest struct {
	SpeedClass string `json:"speedClass"`
}

// JSON retunrs a JSON representing updateSpeedClassRequest object
func (r *updateSpeedClassRequest) JSON() string {
	bodyBytes, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

type setExpiryTimeRequest struct {
	ExpiryTime string `json:"expiryTime"`
}

// JSON retunrs a JSON representing setExpiryTimeRequest object
func (r *setExpiryTimeRequest) JSON() string {
	bodyBytes, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

type setSubscriberGroupRequest struct {
	GroupID string `json:"groupId"`
}

// JSON retunrs a JSON representing setSubscriberGroupRequest object
func (r *setSubscriberGroupRequest) JSON() string {
	bodyBytes, err := json.Marshal(r)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

// StatsPeriod is a period to gather stats
type StatsPeriod int

const (
	// StatsPeriodUnspecified means no StatsPeriod is specified
	StatsPeriodUnspecified StatsPeriod = iota

	// StatsPeriodMonth means the period of gathering stats is 'month'
	StatsPeriodMonth

	// StatsPeriodDay means that the period of gathering stats is 'day'
	StatsPeriodDay

	// StatsPeriodMinutes means that the period of gathering stats is 'minutes'
	StatsPeriodMinutes
)

func (p StatsPeriod) String() string {
	switch p {
	case StatsPeriodMonth:
		return "month"
	case StatsPeriodDay:
		return "day"
	case StatsPeriodMinutes:
		return "minutes"
	}
	return ""
}

// Parse parses the specified string and returns a StatsPeriod value represented by the string
func (p StatsPeriod) Parse(s string) StatsPeriod {
	switch s {
	case "month":
		return StatsPeriodMonth
	case "day":
		return StatsPeriodDay
	case "minutes":
		return StatsPeriodMinutes
	default:
		return StatsPeriodUnspecified
	}
}

// SpeedClass represents one of speed classes
type SpeedClass string

const (
	// SpeedClassS1Minimum is s1.minimum
	SpeedClassS1Minimum SpeedClass = "s1.minimum"

	// SpeedClassS1Slow is s1.slow
	SpeedClassS1Slow SpeedClass = "s1.slow"

	// SpeedClassS1Standard is s1.standard
	SpeedClassS1Standard SpeedClass = "s1.standard"

	// SpeedClassS1Fast is s1.fast
	SpeedClassS1Fast SpeedClass = "s1.fast"
)

// AirStatsForSpeedClass holds Upload/Download Bytes/Packets for a speed class
type AirStatsForSpeedClass struct {
	UploadBytes     uint64 `json:"uploadByteSizeTotal"`
	UploadPackets   uint64 `json:"uploadPacketSizeTotal"`
	DownloadBytes   uint64 `json:"downloadByteSizeTotal"`
	DownloadPackets uint64 `json:"downloadPacketSizeTotal"`
}

// AirStats holds a set of traffic information for each speed class
type AirStats struct {
	Date     string                               `json:"date"`
	Unixtime uint64                               `json:"unixtime"`
	Traffic  map[SpeedClass]AirStatsForSpeedClass `json:"dataTrafficStatsMap"`
}

func parseAirStats(resp *http.Response) []AirStats {
	var v []AirStats
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&v)
	return v
}

// BeamType represents one of in/out protocols for Beam
type BeamType string

const (
	// BeamTypeInHTTP is ...
	BeamTypeInHTTP = "inHttp"
	// BeamTypeInMQTT is ...
	BeamTypeInMQTT = "inMqtt"
	// BeamTypeInTCP is ...
	BeamTypeInTCP = "inTcp"
	// BeamTypeOutHTTP is ...
	BeamTypeOutHTTP = "outHttp"
	// BeamTypeOutHTTPS is ...
	BeamTypeOutHTTPS = "outHttps"
	// BeamTypeOutMQTT is ...
	BeamTypeOutMQTT = "outMqtt"
	// BeamTypeOutMQTTS is ...
	BeamTypeOutMQTTS = "outMqtts"
	// BeamTypeOutTCP is ...
	BeamTypeOutTCP = "outTcp"
	// BeamTypeOutTCPS is ...
	BeamTypeOutTCPS = "outTcps"
)

// BeamStatsForType holds Upload/Download Bytes/Packets for a speed class
type BeamStatsForType struct {
	Count uint64 `json:"count"`
}

// BeamStats holds a set of traffic information for each speed class
type BeamStats struct {
	Date     string                        `json:"date"`
	Unixtime uint64                        `json:"unixtime"`
	Traffic  map[BeamType]BeamStatsForType `json:"beamStatsMap"`
}

func parseBeamStats(resp *http.Response) []BeamStats {
	var v []BeamStats
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&v)
	return v
}

type exportAirStatsRequest struct {
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	Period string `json:"period"`
}

func (r *exportAirStatsRequest) JSON() string {
	return toJSON(r)
}

type exportAirStatsResponse struct {
	URL string `json:"url"`
}

func parseExportAirStatsResponse(resp *http.Response) *exportAirStatsResponse {
	var r exportAirStatsResponse
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&r)
	return &r
}

type exportBeamStatsRequest struct {
	From   int64  `json:"from"`
	To     int64  `json:"to"`
	Period string `json:"period"`
}

func (r *exportBeamStatsRequest) JSON() string {
	return toJSON(r)
}

type exportBeamStatsResponse struct {
	URL string `json:"url"`
}

func parseExportBeamStatsResponse(resp *http.Response) *exportBeamStatsResponse {
	var r exportBeamStatsResponse
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&r)
	return &r
}

// ConfigNamespace is a type of namespace of a configuration
type ConfigNamespace string

// Group keeps information about a group
type Group struct {
	Configuration    map[ConfigNamespace]interface{} `json:"configuration"`
	CreatedTime      *TimestampMilli                 `json:"createdTime"`
	GroupID          string                          `json:"groupId"`
	LastModifiedTime *TimestampMilli                 `json:"lastModifiedTime"`
	OperatorID       string                          `json:"operatorId"`
	Tags             Tags                            `json:"tags"`
}

// ListGroupsOptions holds options for APIClient.ListGroups()
type ListGroupsOptions struct {
	TagName           string
	TagValue          string
	TagValueMatchMode TagValueMatchMode
	Limit             int
	LastEvaluatedKey  string
}

func (lso *ListGroupsOptions) String() string {
	var s = make([]string, 0, 10)
	if lso.TagName != "" {
		s = append(s, "tag_name="+lso.TagName)
	}
	if lso.TagValue != "" {
		s = append(s, "tag_value="+lso.TagValue)
	}
	if lso.TagValueMatchMode != MatchModeUnspecified {
		s = append(s, "tag_value_match_mode="+lso.TagValueMatchMode.String())
	}
	if lso.Limit != 0 {
		s = append(s, "limit="+strconv.Itoa(lso.Limit))
	}
	if lso.LastEvaluatedKey != "" {
		s = append(s, "last_evaluated_key="+lso.LastEvaluatedKey)
	}
	return strings.Join(s, "&")
}

func parseListGroupsResponse(resp *http.Response) ([]Group, *PaginationKeys, error) {
	groups := make([]Group, 0, 10)
	dec := json.NewDecoder(resp.Body)

	// read open bracket
	_, err := dec.Token()
	if err != nil {
		return nil, nil, err
	}

	for dec.More() {
		var g Group
		err = dec.Decode(&g)
		if err != nil {
			continue
		}
		groups = append(groups, g)
	}

	// read close bracket
	_, err = dec.Token()
	if err != nil {
		return nil, nil, err
	}

	linkHeader := resp.Header.Get(http.CanonicalHeaderKey("Link"))
	pk := parseLinkHeader(linkHeader)

	return groups, pk, nil
}

type createGroupRequest struct {
	Tags Tags `json:"tags"`
}

func (r *createGroupRequest) JSON() string {
	return toJSON(r)
}

func parseGroup(resp *http.Response) *Group {
	var g Group
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&g)
	return &g
}

// ListSubscribersInGroupOptions holds options for APIClient.ListSubscribersInGroup()
type ListSubscribersInGroupOptions struct {
	Limit            int
	LastEvaluatedKey string
}

func (lso *ListSubscribersInGroupOptions) String() string {
	var s = make([]string, 0, 10)
	if lso.Limit != 0 {
		s = append(s, "limit="+strconv.Itoa(lso.Limit))
	}
	if lso.LastEvaluatedKey != "" {
		s = append(s, "last_evaluated_key="+lso.LastEvaluatedKey)
	}
	return strings.Join(s, "&")
}

// GroupConfig holds a pair of a key and a value
type GroupConfig struct {
	Key   string      `json:"key"`
	Value interface{} `json:"value"`
}

// MetaData holds configuration for SORACOM Air Metadata
type MetaData struct {
	Enabled     bool   `json:"enabled"`
	ReadOnly    bool   `json:"readonly"`
	AllowOrigin string `json:"allowOrigin"`
}

// AirConfig holds configuration parameters for SORACOM Air
type AirConfig struct {
	UseCustomDNS bool     `json:"useCustomDns"`
	DNSServers   []string `json:"dnsServers"`
	MetaData     MetaData `json:"metadata"`
	UserData     string   `json:"userdata"`
}

// JSON converts AirConfig into JSON string
func (ac *AirConfig) JSON() string {
	return toJSON([]GroupConfig{
		GroupConfig{Key: "useCustomDns", Value: ac.UseCustomDNS},
		GroupConfig{Key: "dnsServers", Value: ac.DNSServers},
	})
}

// CustomHeader holds Action, Key and Value for a custom header
type CustomHeader struct {
	Action string `json:"action"`
	Key    string `json:"headerKey"`
	Value  string `json:"headerValue"`
}

// BeamTCPConfig holds SORACOM Beam TCP entry point configurations
type BeamTCPConfig struct {
	Name                string `json:"name"`
	Destination         string `json:"destination"`
	Enabled             bool   `json:"enabled"`
	AddSubscriberHeader bool   `json:"addSubscriberHeader"`
	AddSignature        bool   `json:"addSignature"`
	PSK                 string `json:"psk"`
}

// BeamUDPConfig holds SORACOM Beam UDP entry point configurations
type BeamUDPConfig struct {
	Name                string `json:"name"`
	Destination         string `json:"destination"`
	Enabled             bool   `json:"enabled"`
	AddSubscriberHeader bool   `json:"addSubscriberHeader"`
	AddSignature        bool   `json:"addSignature"`
	PSK                 string `json:"psk"`
}

// ClientCerts consists of a CA certificate,
type ClientCerts struct {
	CA         string `json:"ca"`
	Cert       string `json:"cert"`
	PrivateKey string `json:"key"`
}

// BeamMQTTConfig holds SORACOM Beam MQTT entry point configurations
type BeamMQTTConfig struct {
	Name                  string                 `json:"name"`
	Destination           string                 `json:"destination"`
	Enabled               bool                   `json:"enabled"`
	AddSubscriberHeader   bool                   `json:"addSubscriberHeader"`
	Username              string                 `json:"username"`
	Password              string                 `json:"password"`
	UseClientCertificates string                 `json:"useClientCert"`
	ClientCertificates    map[string]ClientCerts `json:"clientCerts"`
}

// BeamHTTPConfig holds SORACOM Beam HTTP entry point configurations
type BeamHTTPConfig struct {
	Name                string                  `json:"name"`
	Destination         string                  `json:"destination"`
	Enabled             bool                    `json:"enabled"`
	AddSubscriberHeader bool                    `json:"addSubscriberHeader"`
	AddSignature        bool                    `json:"addSignature"`
	CustomHeaders       map[string]CustomHeader `json:"customHeaders"`
	PSK                 string                  `json:"psk"`
}

// EventHandlerRuleType is a type of event hander's rule
type EventHandlerRuleType string

const (
	// EventHandlerRuleTypeUnspecified means that the type field in RuleConfig has not been specified
	EventHandlerRuleTypeUnspecified EventHandlerRuleType = ""

	// EventHandlerRuleTypeDailyTraffic is a rule type to invoke actions when data traffic for a day for a subscriber exceeds the specified limit
	EventHandlerRuleTypeDailyTraffic EventHandlerRuleType = "DailyTrafficRule"

	// EventHandlerRuleTypeMonthlyTraffic is a rule type to invoke actions when data traffic for a month for a subscriber exceeds the specified limit
	EventHandlerRuleTypeMonthlyTraffic EventHandlerRuleType = "MonthlyTraffic"

	// EventHandlerRuleTypeCumulativeTraffic is a rule type to invoke actions when cumulative data traffic for a subscriber exceeds the specified limit
	EventHandlerRuleTypeCumulativeTraffic EventHandlerRuleType = "CumulativeTraffic"

	// EventHandlerRuleTypeDailyTotalTraffic is a rule type to invoke actions when total data traffic for a day for all subscribers exceeds the specified limit
	EventHandlerRuleTypeDailyTotalTraffic EventHandlerRuleType = "DailyTotalTraffic"

	// EventHandlerRuleTypeMonthlyTotalTraffic is a rule type to invoke actions when total data traffic for a month for all subscribers exceeds the specified limit
	EventHandlerRuleTypeMonthlyTotalTraffic EventHandlerRuleType = "MonthlyTotalTraffic"

	// EventHandlerRuleTypeSubscriberStatusChanged is a rule type to invoke actions when status of a subscriber has been changed
	EventHandlerRuleTypeSubscriberStatusChanged EventHandlerRuleType = "SubscriberStatusChanged"

	// EventHandlerRuleTypeSubscriberSpeedClassChanged is a rule type to invoke actions when speed class of a subscriber has been changed
	EventHandlerRuleTypeSubscriberSpeedClassChanged EventHandlerRuleType = "SubscriberSpeedClassChanged"
)

// RuleConfig contains a condition to invoke actions
type RuleConfig struct {
	Type       EventHandlerRuleType `json:"type"`
	Properties Properties           `json:"properties"`
}

// EventHandlerActionType is a type of event hander's action
type EventHandlerActionType string

const (
	// EventHandlerActionTypeUnspecified means that the type field in ActionConfigList has not been specified
	EventHandlerActionTypeUnspecified EventHandlerActionType = ""

	// EventHandlerActionTypeChangeSpeedClass indicates a type of action to be invoked to change speed class for a subscriber once a condition is satisfied
	EventHandlerActionTypeChangeSpeedClass EventHandlerActionType = "ChangeSpeedClassAction"

	// EventHandlerActionTypeSendMail indicates a type of action to be invoked to send an email once a condition is satisfied
	EventHandlerActionTypeSendMail EventHandlerActionType = "SendMailAction"

	// EventHandlerActionTypeInvokeAWSLambda indicates a type of action to be invoked to invoke AWS Lambda function once a condition is satisfied
	EventHandlerActionTypeInvokeAWSLambda EventHandlerActionType = "InvokeAWSLambdaAction"
)

// ActionConfig contains an action to be invoked when a condition is satisfied
type ActionConfig struct {
	Type       EventHandlerActionType `json:"type"`
	Properties Properties             `json:"properties"`
}

// EventHandler keeps information about an event handler
type EventHandler struct {
	HandlerID        string         `json:"handlerId"`
	TargetImsi       *string        `json:"targetImsi"`
	TargetOperatorID *string        `json:"targetOperatorId"`
	TargetTag        *Tags          `json:"targetTag"`
	TargetGroupID    *string        `json:"targetGroupId"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	RuleConfig       RuleConfig     `json:"ruleConfig"`
	Status           string         `json:"status"`
	ActionConfigList []ActionConfig `json:"actionConfigList"`
}

// JSON converts Eventhandler into a JSON string
func (o *EventHandler) JSON() string {
	return toJSON(o)
}

// ListEventHandlersOptions holds options for APIClient.ListEventHandlers()
type ListEventHandlersOptions struct {
	Target string
}

func (leho *ListEventHandlersOptions) String() string {
	return leho.Target
}

func parseListEventHandlersResponse(resp *http.Response) ([]EventHandler, error) {
	eventHandlers := make([]EventHandler, 0, 10)
	dec := json.NewDecoder(resp.Body)

	// read open bracket
	_, err := dec.Token()
	if err != nil {
		return nil, err
	}

	for dec.More() {
		var eh EventHandler
		err = dec.Decode(&eh)
		if err != nil {
			continue
		}
		eventHandlers = append(eventHandlers, eh)
	}

	// read close bracket
	_, err = dec.Token()
	if err != nil {
		return nil, err
	}

	return eventHandlers, nil
}

func parseEventHandler(resp *http.Response) (*EventHandler, error) {
	dec := json.NewDecoder(resp.Body)
	var eh EventHandler
	err := dec.Decode(&eh)
	if err != nil {
		return nil, err
	}
	return &eh, nil
}

// CreateEventHandlerOptions keeps information to create an event handler
type CreateEventHandlerOptions struct {
	TargetImsi       *string        `json:"targetImsi"`
	TargetOperatorID *string        `json:"targetOperatorId"`
	TargetTag        *Tags          `json:"targetTag"`
	TargetGroupID    *string        `json:"targetGroupId"`
	Name             string         `json:"name"`
	Description      string         `json:"description"`
	RuleConfig       RuleConfig     `json:"ruleConfig"`
	Status           string         `json:"status"`
	ActionConfigList []ActionConfig `json:"actionConfigList"`
}

// JSON converts CreateEventhandlerOptions into a JSON string
func (o *CreateEventHandlerOptions) JSON() string {
	return toJSON(o)
}

func tagsToJSON(tags []Tag) string {
	bodyBytes, err := json.Marshal(tags)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

func readAll(r io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.String()
}

func percentEncoding(s string) string {
	return url.QueryEscape(s)
}

func toJSON(x interface{}) string {
	bodyBytes, err := json.Marshal(x)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

func dumpHTTPResponse(resp *http.Response) {
	dump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(string(dump))
}

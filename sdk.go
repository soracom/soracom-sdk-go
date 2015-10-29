// Package: SORACOM SDK for Go.

package soracom

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
	SpeedClassS1Minimum = "s1.minimum"

	// SpeedClassS1Slow is s1.slow
	SpeedClassS1Slow = "s1.slow"

	// SpeedClassS1Standard is s1.standard
	SpeedClassS1Standard = "s1.standard"

	// SpeedClassS1Fast is s1.fast
	SpeedClassS1Fast = "s1.fast"
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
	RegistrationSecret string            `json:"registrationSecret"`
	GroupID            string            `json:"groupId"`
	Tags               map[string]string `json:"tags"`
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
	Apn                string            `json:"apn"`
	CreatedTime        *TimestampMilli   `json:"createdTime"`
	ExpiryTime         *TimestampMilli   `json:"expiryTime"`
	GroupID            *string           `json:"groupId"`
	Imsi               string            `json:"imsi"`
	IPAddress          *string           `json:"ipAddress"`
	LastModifiedTime   *TimestampMilli   `json:"lastModifiedTime"`
	ModuleType         string            `json:"ModuleType"`
	Msisdn             string            `json:"msisdn"`
	OperatorID         string            `json:"operatorId"`
	Plan               int               `json:"plan"`
	SessionStatus      *SessionStatus    `json:"sessionStatus"`
	Status             string            `json:"status"`
	SpeedClass         string            `json:"type"`
	Tags               map[string]string `json:"tags"`
	TerminationEnabled bool              `json:"terminationEnabled"`
}

// PaginationKeys holds keys for pagination
type PaginationKeys struct {
	Prev string
	Next string
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

	var pk *PaginationKeys
	linkHeader := resp.Header.Get(http.CanonicalHeaderKey("Link"))
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

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

// Timestamp is ...
type Timestamp struct {
	time.Time
}

// MarshalJSON is ...
func (t *Timestamp) MarshalJSON() ([]byte, error) {
	ts := t.Time.Unix()
	stamp := fmt.Sprint(ts)

	return []byte(stamp), nil
}

// UnmarshalJSON is ...
func (t *Timestamp) UnmarshalJSON(b []byte) error {
	ts, err := strconv.Atoi(string(b))
	if err != nil {
		return err
	}

	t.Time = time.Unix(int64(ts), 0)

	return nil
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

// SessionStatus keeps information about a session
type SessionStatus struct {
	DNSServers    []string   `json:"dnsServers"`
	Imei          string     `json:"imei"`
	LastUpdatedAt *Timestamp `json:"lastUpdatedAt"`
	Location      *string    `json:"location"`
	Online        bool       `json:"online"`
	UEIPAddress   string     `json:"ueIpAddress"`
}

// Subscriber keeps information about a subscriber
type Subscriber struct {
	Apn                string            `json:"apn"`
	CreatedTime        *Timestamp        `json:"createdTime"`
	ExpiryTime         *Timestamp        `json:"expiryTime"`
	GroupID            *string           `json:"groupId"`
	Imsi               string            `json:"imsi"`
	IPAddress          *string           `json:"ipAddress"`
	LastModifiedTime   *Timestamp        `json:"lastModifiedTime"`
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

func parseGetSubscriberResponse(resp *http.Response) *Subscriber {
	var sub Subscriber
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&sub)
	return &sub
}

func readAll(r io.Reader) string {
	buf := new(bytes.Buffer)
	buf.ReadFrom(r)
	return buf.String()
}

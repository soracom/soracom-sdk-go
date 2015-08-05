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

type ApiClient struct {
	httpClient *http.Client
	Token      string
}

type AuthRequestBody struct {
	Email               string `json:"email"`
	Password            string `json:"password"`
	TokenTimeoutSeconds int    `json:"tokenTimeoutSeconds"`
}

func (arb *AuthRequestBody) Json() string {
	bodyBytes, err := json.Marshal(arb)
	if err != nil {
		return ""
	}
	return string(bodyBytes)
}

type AuthResponseBody struct {
	ApiKey     string
	OperatorId string
	Token      string
}

func ParseAuthResponse(resp *http.Response) *AuthResponseBody {
	var arb AuthResponseBody
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&arb)
	return &arb
}

type TagValueMatchMode int

const (
	MatchExact TagValueMatchMode = iota
	MatchPrefix
)

func (m TagValueMatchMode) String() string {
	switch m {
	case MatchExact:
		return "exact"
	case MatchPrefix:
		return "prefix"
	}
	return ""
}

type ListSubscribersOptions struct {
	TagName           string
	TagValue          string
	TagValueMatchMode *TagValueMatchMode
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
	if lso.TagValueMatchMode != nil {
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

type PaginationKeys struct {
	Prev string
	Next string
}

func ParseListSubscribersResponse(resp *http.Response) (subs []Subscriber, pk *PaginationKeys) {
	dec := json.NewDecoder(resp.Body)
	dec.Decode(&subs)

	linkHeader := resp.Header.Get(http.CanonicalHeaderKey("Link"))
	if linkHeader == "" {
		return
	}

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

	return
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
	query       string
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

func NewApiClient() *ApiClient {
	hc := new(http.Client)
	return &ApiClient{
		httpClient: hc,
		Token:      "",
	}
}

func (ac *ApiClient) getEndpointBase() string {
	return "http://sora-prod.apigee.net"
}

func (ac *ApiClient) callApi(params *apiParams) (*http.Response, error) {
	url := ac.getEndpointBase() + params.path
	if params.query != "" {
		url += "?" + params.query
	}
	fmt.Printf("url == %v\n", url)
	req, err := http.NewRequest(
		params.method, url, strings.NewReader(params.body))
	if err != nil {
		return nil, err
	}

	if params.contentType != "" {
		req.Header.Set("Content-Type", params.contentType)
	}

	req.Header.Set("X-Soracom-Token", ac.Token)

	return ac.httpClient.Do(req)
}

func (ac *ApiClient) Auth(email, password string) error {
	body := &AuthRequestBody{
		Email:               email,
		Password:            password,
		TokenTimeoutSeconds: 24 * 60 * 60,
	}
	params := &apiParams{
		method:      "POST",
		path:        "/v1/operators/auth",
		contentType: "application/json",
		body:        body.Json(),
	}

	resp, err := ac.callApi(params)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return NewApiError(resp)
	}

	respBody := ParseAuthResponse(resp)
	ac.Token = respBody.Token

	return nil
}

func (ac *ApiClient) ListSubscribers(options *ListSubscribersOptions) ([]Subscriber, *PaginationKeys, error) {
	params := &apiParams{
		method: "GET",
		path:   "/v1/subscribers",
	}
	if options != nil {
		params.query = options.String()
	}
	resp, err := ac.callApi(params)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= http.StatusBadRequest {
		return nil, nil, NewApiError(resp)
	}

	subscribers, paginationKeys := ParseListSubscribersResponse(resp)

	return subscribers, paginationKeys, nil
}

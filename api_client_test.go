package soracom

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	apiClient          *APIClient
	email              string
	password           string
	createdSubscribers []CreatedSubscriber
)

const (
	defaultEndpointForTest = "https://api-sandbox.soracom.io"
	nSIM                   = 25
)

func TestMain(m *testing.M) {
	err := setup()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func setup() error {
	rand.Seed(time.Now().Unix())
	apiClient = setupAPIClient()
	if os.Getenv("SORACOM_VERBOSE") != "" {
		apiClient.SetVerbose(true)
	}

	email = os.Getenv("SORACOM_EMAIL_FOR_TEST")
	if email == "" {
		return errors.New("SORACOM_EMAIL_FOR_TEST env var is required")
	}
	email = randomizeEmail(email)
	if email == "" {
		return errors.New("SORACOM_EMAIL_FOR_TEST might be in invalid format")
	}

	password = os.Getenv("SORACOM_PASSWORD_FOR_TEST")
	if password == "" {
		return errors.New("SORACOM_PASSWORD_FOR_TEST env var is required")
	}

	err := signup()
	if err != nil {
		return err
	}

	err = auth()
	if err != nil {
		return err
	}

	// auth again to update token
	err = auth()
	if err != nil {
		return err
	}

	createdSubscribers = make([]CreatedSubscriber, 0, nSIM)
	for i := 0; i < nSIM; i++ {
		s, err := apiClient.CreateSubscriber()
		if err != nil {
			return err
		}
		createdSubscribers = append(createdSubscribers, *s)
	}

	err = registerSubscribers()
	if err != nil {
		return err
	}

	return nil
}

func setupAPIClient() *APIClient {
	endpoint := os.Getenv("SORACOM_ENDPOINT")
	if endpoint == "" {
		endpoint = defaultEndpointForTest
	}

	options := &APIClientOptions{
		Endpoint: endpoint,
	}

	return NewAPIClient(options)
}

func randomizeEmail(email string) string {
	s := strings.Split(email, "@")
	if len(s) != 2 {
		return ""
	}

	return s[0] + "+" + getRandomString(20) + "@" + s[1]
}

func getRandomString(size uint) string {
	p := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l",
		"m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z"}
	s := ""
	for i := uint(0); i < size; i++ {
		s += p[rand.Intn(len(p))]
	}
	return s
}

func signup() error {
	authKeyID := os.Getenv("SORACOM_AUTHKEY_ID_FOR_TEST")
	if authKeyID == "" {
		return errors.New("SORACOM_AUTHKEY_ID_FOR_TEST env var is required")
	}
	authKey := os.Getenv("SORACOM_AUTHKEY_FOR_TEST")
	if authKey == "" {
		return errors.New("SORACOM_AUTHKEY_FOR_TEST env var is required")
	}

	_, err := apiClient.InitOperatorForSandbox(email, password, authKeyID, authKey, true, []string{"jp"})

	if err != nil {
		return err
	}

	return nil
}

func auth() error {
	err := apiClient.Auth(email, password)
	if err != nil {
		return err
	}

	return nil
}

func registerSubscribers() error {
	for i, cs := range createdSubscribers {
		o := RegisterSubscriberOptions{
			RegistrationSecret: cs.RegistrationSecret,
			Tags:               Tags{},
		}
		if i%3 == 0 {
			o.Tags["soracom-sdk-go-test"] = "foo"
		}
		if i%3 == 1 {
			o.Tags["soracom-sdk-go-test"] = "hoge"
		}
		if i%3 == 2 {
			o.Tags["soracom-sdk-go-test"] = "beam-stats"
		}
		_, err := apiClient.RegisterSubscriber(cs.IMSI, o)
		if err != nil {
			return err
		}
		if i%4 == 0 {
			_, err := apiClient.ActivateSubscriber(cs.IMSI)
			if err != nil {
				return err
			}
		}
		if i%4 == 1 {
			_, err := apiClient.DeactivateSubscriber(cs.IMSI)
			if err != nil {
				return err
			}
		}
		if i%5 == 0 {
			_, err := apiClient.UpdateSubscriberSpeedClass(cs.IMSI, "s1.minimum")
			if err != nil {
				return err
			}
		}
		if i%5 == 1 {
			_, err := apiClient.UpdateSubscriberSpeedClass(cs.IMSI, "s1.slow")
			if err != nil {
				return err
			}
		}
		if i%5 == 2 {
			_, err := apiClient.UpdateSubscriberSpeedClass(cs.IMSI, "s1.standard")
			if err != nil {
				return err
			}
		}
		if i%5 == 3 {
			_, err := apiClient.UpdateSubscriberSpeedClass(cs.IMSI, "s1.fast")
			if err != nil {
				return err
			}
		}

		baseTime := time.Now()
		for j := 0; j < 10; j++ {
			t := baseTime.AddDate(0, 0, -10*j).Add(time.Duration(j+1) * time.Second)
			ts := t.Unix()
			err := apiClient.InsertAirStats(cs.IMSI, generateDummyAirStats(ts))
			if err != nil {
				return err
			}
		}

		for k := 0; k < 10; k++ {
			t := baseTime.AddDate(0, 0, -10*k).Add(2 * time.Duration(k+1) * time.Second)
			ts := t.Unix()
			err := apiClient.InsertBeamStats(cs.IMSI, generateDummyBeamStats(ts))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func getRandomSpeedClass() SpeedClass {
	switch rand.Intn(4) {
	case 0:
		return SpeedClassS1Minimum
	case 1:
		return SpeedClassS1Slow
	case 2:
		return SpeedClassS1Standard
	case 3:
		return SpeedClassS1Fast
	}
	return SpeedClassS1Standard
}

func generateDummyAirStats(ts int64) AirStats {
	ub := rand.Intn(1000000)
	up := ub / (rand.Intn(100) + 1)
	db := rand.Intn(1000000)
	dp := db / (rand.Intn(100) + 1)
	t := make(map[SpeedClass]AirStatsForSpeedClass)
	t[getRandomSpeedClass()] = AirStatsForSpeedClass{
		UploadBytes:     uint64(ub),
		UploadPackets:   uint64(up),
		DownloadBytes:   uint64(db),
		DownloadPackets: uint64(dp),
	}
	return AirStats{
		Unixtime: uint64(ts),
		Traffic:  t,
	}
}

func getRandomBeamType() BeamType {
	switch rand.Intn(11) {
	case 0:
		return BeamTypeInHTTP
	case 1:
		return BeamTypeInMQTT
	case 2:
		return BeamTypeInTCP
	case 3:
		return BeamTypeInUDP
	case 4:
		return BeamTypeOutHTTP
	case 5:
		return BeamTypeOutHTTPS
	case 6:
		return BeamTypeOutMQTT
	case 7:
		return BeamTypeOutMQTTS
	case 8:
		return BeamTypeOutTCP
	case 9:
		return BeamTypeOutTCPS
	case 10:
		return BeamTypeOutUDP
	}
	return BeamTypeInHTTP
}

func generateDummyBeamStats(ts int64) BeamStats {
	t := make(map[BeamType]BeamStatsForType)
	t[getRandomBeamType()] = BeamStatsForType{
		Count: uint64(rand.Intn(1000)),
	}
	return BeamStats{
		Unixtime: uint64(ts),
		Traffic:  t,
	}
}

func TestGenerateAPIToken(t *testing.T) {
	_, err := apiClient.GenerateAPIToken(0)
	if err != nil {
		t.Fatalf("GenerateAPIToken(0) failed")
	}

	_, err = apiClient.GenerateAPIToken(1)
	if err != nil {
		t.Fatalf("GenerateAPIToken(1) failed")
	}

	_, err = apiClient.GenerateAPIToken(48 * 60 * 60)
	if err != nil {
		t.Fatalf("GenerateAPIToken(MAX) failed")
	}

	/*
		_, err = apiClient.GenerateAPIToken(48*60*60 + 1)
		if err == nil {
			t.Fatalf("GenerateAPIToken(MAX + 1) should fail")
		}

		_, err = apiClient.GenerateAPIToken(-1)
		if err == nil {
			t.Fatalf("GenerateAPIToken(-1) should fail")
		}

	*/
}

func TestGetSupportToken(t *testing.T) {
	token, err := apiClient.GetSupportToken()
	if err != nil {
		t.Fatalf("GetSupportToken() failed")
	}
	if token == "" {
		t.Fatalf("Unable to get a support token")
	}
}

func TestGetOperator(t *testing.T) {
	o, err := apiClient.GetOperator(apiClient.OperatorID)
	if err != nil {
		t.Fatalf("GetOperator() failed")
	}
	if o.OperatorID != apiClient.OperatorID {
		t.Fatalf("Got an unexpected operator")
	}
}

func TestListSubscribers(t *testing.T) {
	subs, pk, err := apiClient.ListSubscribers(nil)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers(nil): %v", err.Error())
	}
	if len(subs) < 10 {
		t.Fatalf("10+ subscribers are required to proceed the test")
	}
	if pk != nil {
		t.Fatalf("Pagination keys must not be returned if Limit option is not specified")
	}

	options := &ListSubscribersOptions{
		Limit: 10,
	}
	subs, pk, err = apiClient.ListSubscribers(options)
	if err != nil {
		t.Fatalf("Error ocurred on ListSubscribers({Limit: 10}): %v", err.Error())
	}
	if len(subs) != 10 {
		t.Fatalf("Limit option does not have any effect. Length of subscribers = %v", len(subs))
	}
	if pk == nil {
		t.Fatalf("Pagination keys must be returned if Limit option is specified and there are subscribers more than the limit")
	}
	//t.Logf("subs == %v", subs)

	options2 := &ListSubscribersOptions{
		Limit:            10,
		LastEvaluatedKey: pk.Next,
	}
	subs2, _, err := apiClient.ListSubscribers(options2)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers({Limit:10,LastEvaluatedKey:xxx})")
	}
	if sameSubscribers(subs, subs2) {
		t.Fatalf("Pagination seems not working")
	}
	//t.Logf("subs2 == %v", subs2)
}

func TestListSubscribersByTag(t *testing.T) {
	options := &ListSubscribersOptions{
		TagName:  "soracom-sdk-go-test",
		TagValue: "foo",
	}
	subs, _, err := apiClient.ListSubscribers(options)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers({TagName: 'soracom-sdk-go-test', TagValue: 'foo'})")
	}
	if len(subs) < 5 {
		t.Fatalf("5+ subscribers tagged with 'soracom-sdk-go-test' (tag-name) / 'foo' (tag-value) are required")
	}
	//t.Logf("subs == %v", subs)

	options2 := &ListSubscribersOptions{
		TagName:  "soracom-sdk-go-test",
		TagValue: "foo",
		Limit:    3,
	}
	subs2, pk2, err := apiClient.ListSubscribers(options2)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers({TagName: 'soracom-sdk-go-test', TagValue: 'foo', Limit: 3})")
	}
	if len(subs2) != 3 {
		t.Fatalf("3+ subscribers tagged with 'soracom-sdk-go-test' (tag-name) / 'foo' (tag-value) are required")
	}
	if pk2 == nil {
		t.Fatalf("Pagination keys must be returned if Limit option is specified and there are subscribers more than the limit")
	}
	for i := range subs2 {
		if subs2[i].Tags["soracom-sdk-go-test"] != "foo" {
			t.Fatalf("Found a subscriber with the name which is not specified")
		}
	}
	//t.Logf("subs2 == %v", subs2)

	options3 := &ListSubscribersOptions{
		TagName:          "soracom-sdk-go-test",
		TagValue:         "foo",
		Limit:            3,
		LastEvaluatedKey: pk2.Next,
	}
	subs3, _, err := apiClient.ListSubscribers(options3)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs3) == 0 {
		t.Fatalf("5+ subscribers tagged with 'soracom-sdk-go-test' (tag-name) / 'foo' (tag-value) are required")
	}
	if sameSubscribers(subs2, subs3) {
		t.Fatalf("Pagination seems not working with tag search")
	}
	for i := range subs3 {
		if subs3[i].Tags["soracom-sdk-go-test"] != "foo" {
			t.Fatalf("Found a subscriber with the name which is not specified")
		}
	}
	//t.Logf("subs3 == %v", subs3)
}

func TestListSubscribersByTagPrefix(t *testing.T) {
	options := &ListSubscribersOptions{
		TagName:           "soracom-sdk-go-test",
		TagValue:          "fo",
		TagValueMatchMode: MatchModePrefix,
	}
	subs, _, err := apiClient.ListSubscribers(options)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs) < 5 {
		t.Fatalf("5+ subscribers tagged with 'soracom-sdk-go-test' (tag-name) / 'ho*' (tag-value) are required")
	}
	//t.Logf("subs == %v", subs)

	options2 := &ListSubscribersOptions{
		TagName:           "soracom-sdk-go-test",
		TagValue:          "fo",
		TagValueMatchMode: MatchModePrefix,
		Limit:             3,
	}
	subs2, pk2, err := apiClient.ListSubscribers(options2)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs2) != 3 {
		t.Fatalf("3+ subscribers tagged with 'soracom-sdk-go-test' (tag-name) / 'ho*' (tag-value) are required")
	}
	if pk2 == nil {
		t.Fatalf("Pagination keys must be returned if Limit option is specified and there are subscribers more than the limit")
	}
	for i := range subs2 {
		if !strings.HasPrefix(subs2[i].Tags["soracom-sdk-go-test"], "fo") {
			t.Fatalf("Found a subscriber with the name which is not specified")
		}
	}
	//t.Logf("subs2 == %v", subs2)

	options3 := &ListSubscribersOptions{
		TagName:           "soracom-sdk-go-test",
		TagValue:          "fo",
		TagValueMatchMode: MatchModePrefix,
		Limit:             3,
		LastEvaluatedKey:  pk2.Next,
	}
	subs3, _, err := apiClient.ListSubscribers(options3)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs3) == 0 {
		t.Fatalf("5+ subscribers tagged with 'soracom-sdk-go-test' (tag-name) / 'ho*' (tag-value) are required")
	}
	if sameSubscribers(subs2, subs3) {
		t.Fatalf("Pagination seems not working with tag search")
	}
	for i := range subs3 {
		if !strings.HasPrefix(subs3[i].Tags["soracom-sdk-go-test"], "fo") {
			t.Fatalf("Found a subscriber with the name which is not specified")
		}
	}
	//t.Logf("subs3 == %v", subs3)
}

func TestListSubscribersWithStatusFilter(t *testing.T) {
	options := &ListSubscribersOptions{
		StatusFilter: "inactive",
	}
	subs, _, err := apiClient.ListSubscribers(options)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers({StatusFilter: 'inactive'})")
	}
	if len(subs) < 1 {
		t.Fatalf("5+ subscribers with inactive status are required")
	}
	//t.Logf("subs == %v", subs)

	options2 := &ListSubscribersOptions{
		StatusFilter: "inactive",
		Limit:        3,
	}
	subs2, pk2, err := apiClient.ListSubscribers(options2)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers({StatusFilter: 'inactive', Limit: 3})")
	}
	if len(subs2) != 3 {
		t.Fatalf("5+ subscribers with inactive status are required")
	}
	if pk2 == nil {
		t.Fatalf("Pagination keys must be returned if Limit option is specified and there are subscribers more than the limit")
	}
	for i := range subs2 {
		if subs2[i].Status != "inactive" {
			t.Fatalf("Found a subscriber with the status which is not specified")
		}
	}
	//t.Logf("subs == %v", subs2)

	options3 := &ListSubscribersOptions{
		StatusFilter:     "inactive",
		Limit:            3,
		LastEvaluatedKey: pk2.Next,
	}
	subs3, _, err := apiClient.ListSubscribers(options3)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs3) == 0 {
		t.Fatalf("5+ subscribers with inactive status are required")
	}
	if sameSubscribers(subs2, subs3) {
		t.Fatalf("Pagination seems not working with status filter")
	}
	for i := range subs3 {
		if subs3[i].Status != "inactive" {
			t.Fatalf("Found a subscriber with the status which is not specified")
		}
	}
	//t.Logf("subs3 == %v", subs3)
}

func TestListSubscribersByNameWithStatusFilter(t *testing.T) {
	options := &ListSubscribersOptions{
		StatusFilter: "active|inactive",
		TagName:      "soracom-sdk-go-test",
		TagValue:     "foo",
		Limit:        3,
	}
	subs, pk, err := apiClient.ListSubscribers(options)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs) != 3 {
		t.Fatalf("5+ subscribers tagged with 'soracom-sdk-go-test' (tag-name) / 'foo' (tag-value) and with active or inactive status are required")
	}
	//t.Logf("subs == %v", subs)

	options2 := &ListSubscribersOptions{
		StatusFilter:     "active|inactive",
		TagName:          "soracom-sdk-go-test",
		TagValue:         "foo",
		Limit:            3,
		LastEvaluatedKey: pk.Next,
	}
	subs2, pk2, err := apiClient.ListSubscribers(options2)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs2) == 0 {
		t.Fatalf("5+ subscribers with inactive status are required")
	}
	if pk2 == nil {
		t.Fatalf("Pagination keys must be returned if Limit option is specified and there are subscribers more than the limit")
	}
	if sameSubscribers(subs, subs2) {
		t.Fatalf("Pagination seems not working")
	}
	for i := range subs2 {
		status := subs2[i].Status
		if status != "active" && status != "inactive" {
			t.Fatalf("Found a subscriber with the status which is not specified")
		}
	}
	//t.Logf("subs == %v", subs2)
}

func TestListSubscribersWithTypeFilter(t *testing.T) {
	options := &ListSubscribersOptions{
		TypeFilter: "s1.minimum",
	}
	subs, _, err := apiClient.ListSubscribers(options)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs) < 1 {
		t.Fatalf("5+ subscribers with s1.minimum are required")
	}
	//t.Logf("subs == %v", subs)

	options2 := &ListSubscribersOptions{
		TypeFilter: "s1.minimum",
		Limit:      3,
	}
	subs2, pk2, err := apiClient.ListSubscribers(options2)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs2) != 3 {
		t.Fatalf("5+ subscribers with s1.minimum are required")
	}
	if pk2 == nil {
		t.Fatalf("Pagination keys must be returned if Limit option is specified and there are subscribers more than the limit")
	}
	for i := range subs2 {
		if subs2[i].SpeedClass != "s1.minimum" {
			t.Fatalf("Found a subscriber with speed class which is not specified")
		}
	}
	//t.Logf("subs == %v", subs2)

	options3 := &ListSubscribersOptions{
		TypeFilter:       "s1.minimum",
		Limit:            3,
		LastEvaluatedKey: pk2.Next,
	}
	subs3, _, err := apiClient.ListSubscribers(options3)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs3) == 0 {
		t.Fatalf("5+ subscribers with s1.minimum are required")
	}
	if sameSubscribers(subs2, subs3) {
		t.Fatalf("Pagination seems not working with type filter")
	}
	for i := range subs3 {
		if subs3[i].SpeedClass != "s1.minimum" {
			t.Fatalf("Found a subscriber with speed class which is not specified")
		}
	}
	//t.Logf("subs3 == %v", subs3)
}

func TestGetSubscriber(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		Limit: 1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 subscriber is required")
	}
	imsi := subs[0].IMSI
	sub, err := apiClient.GetSubscriber(imsi)
	if err != nil {
		t.Fatalf("Error occurred on GetSubscriber(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
}

func TestUpdateSubscriberSpeedClass(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		Limit: 1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 subscriber is required")
	}
	imsi := subs[0].IMSI
	sub, err := apiClient.UpdateSubscriberSpeedClass(imsi, "s1.minimum")
	if err != nil {
		t.Fatalf("Error occurred on UpdateSubscriberSpeedClass(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.SpeedClass != "s1.minimum" {
		t.Fatalf("Found a subscriber speed class which is not specified")
	}

	sub, err = apiClient.UpdateSubscriberSpeedClass(imsi, "s1.fast")
	if err != nil {
		t.Fatalf("Error occurred on UpdateSubscriberSpeedClass(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.SpeedClass != "s1.fast" {
		t.Fatalf("Found a subscriber speed class which is not specified")
	}
}

func TestActivateSubscriber(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "inactive",
		Limit:        1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 inactive subscriber is required")
	}
	imsi := subs[0].IMSI
	sub, err := apiClient.ActivateSubscriber(imsi)
	if err != nil {
		t.Fatalf("Error occurred on ActivateSubscriber(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.Status != "active" {
		t.Fatalf("Found a subscriber status which is not specified")
	}
}

func TestDeactivateSubscriber(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "active",
		Limit:        1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 active subscriber is required")
	}
	imsi := subs[0].IMSI
	sub, err := apiClient.DeactivateSubscriber(imsi)
	if err != nil {
		t.Fatalf("Error occurred on DeactivateSubscriber(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.Status != "inactive" {
		t.Fatalf("Found a subscriber status which is not specified")
	}
}

func TestEnableTermination(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "active|inactive|ready",
		Limit:        1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 non-terminated subscriber is required")
	}
	imsi := subs[0].IMSI
	sub, err := apiClient.EnableSubscriberTermination(imsi)
	if err != nil {
		t.Fatalf("Error occurred on EnableSubscriberTermination(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if !sub.TerminationEnabled {
		t.Fatalf("Termination must be enabled")
	}

	sub, err = apiClient.DisableSubscriberTermination(imsi)
	if err != nil {
		t.Fatalf("Error occurred on DisableSubscriberTermination(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.TerminationEnabled {
		t.Fatalf("Termination must be disabled")
	}

	sub, err = apiClient.TerminateSubscriber(imsi)
	if err == nil {
		t.Fatalf("Termination must be failed when termination is disabled")
	}
	if sub != nil && sub.Status == "terminated" {
		t.Fatalf("Termination must not be done")
	}

	sub, err = apiClient.EnableSubscriberTermination(imsi)
	if err != nil {
		t.Fatalf("Error occurred on EnableSubscriberTermination(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if !sub.TerminationEnabled {
		t.Fatalf("Termination must be enabled")
	}

	sub, err = apiClient.DisableSubscriberTermination(imsi)
	if err != nil {
		t.Fatalf("Error occurred on DisableSubscriberTermination(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.TerminationEnabled {
		t.Fatalf("Termination must be disabled")
	}

	sub, err = apiClient.TerminateSubscriber(imsi)
	if err == nil {
		t.Fatalf("Termination must be failed when termination is disabled")
	}
	if sub != nil && sub.Status == "terminated" {
		t.Fatalf("Termination must not be done")
	}

	sub, err = apiClient.EnableSubscriberTermination(imsi)
	if err != nil {
		t.Fatalf("Error occurred on EnableSubscriberTermination(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if !sub.TerminationEnabled {
		t.Fatalf("Termination must be enabled")
	}

	sub, err = apiClient.TerminateSubscriber(imsi)
	if err != nil {
		t.Fatalf("Termination must be completed successfully when termination is enabled")
	}
	if sub.Status != "terminated" {
		t.Fatalf("Termination must be done")
	}
}

func TestSessionEvent(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "active|inactive|ready",
		Limit:        1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 non-terminated subscriber is required")
	}
	imsi := subs[0].IMSI

	opt := &ListSessionEventsOption{
		// From:  time.Now().Add(time.Hour * -24),
		// To:    time.Now(),
		Limit: 1,
	}
	result, pek, err := apiClient.ListSessionEvents(imsi, opt)
	if err != nil {
		t.Fatalf("Error occurred on ListSessionEvents: %s", err)
	}

	if pek != nil {
		t.Fatal("ListSessionEvents should not contain pagination keys")
	}

	if len(result) > 0 {
		t.Logf("found event %d", len(result))
	}
}

func TestSetSubscriberExpiredAt(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "active",
		Limit:        1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 active subscriber is required")
	}
	imsi := subs[0].IMSI
	tomorrow := time.Now().AddDate(0, 0, 1)
	sub, err := apiClient.SetSubscriberExpiredAt(imsi, tomorrow)
	if err != nil {
		t.Fatalf("Error occurred on SetSubscriberExpiredAt(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.ExpiredAt.UnixMilli() != (tomorrow.UnixNano() / 1000 / 1000) {
		fmt.Printf("sub.ExpiredAt.Time == %s, tomorrow == %s", sub.ExpiredAt.Time, tomorrow)
		t.Fatalf("Expiry time for a subscriber has not been updated")
	}
}

func TestUnsetSubscriberExpiredAt(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "active",
		Limit:        1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 active subscriber is required")
	}
	imsi := subs[0].IMSI
	tomorrow := time.Now().AddDate(0, 0, 1)
	sub, err := apiClient.SetSubscriberExpiredAt(imsi, tomorrow)
	if err != nil {
		t.Fatalf("Error occurred on SetSubscriberExpiredAt(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.ExpiredAt.UnixMilli() != (tomorrow.UnixNano() / 1000 / 1000) {
		fmt.Printf("sub.ExpiredAt.Time == %s, tomorrow == %s", sub.ExpiredAt.Time, tomorrow)
		t.Fatalf("Expiry time for a subscriber has not been updated")
	}

	sub, err = apiClient.UnsetSubscriberExpiredAt(imsi)
	if err != nil {
		t.Fatalf("Error occurred on UnsetSubscriberExpiredAt(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.ExpiredAt != nil {
		t.Fatalf("Expiry time for a subscriber must be nil")
	}
}

const (
	tagNameForTest1    = "soracom-sdk-go-test-tag-name-日本語\\$%25&&?*-_."
	tagValueForTest1   = "!@#$%^&*()_-=_+"
	tagValueForTest1_2 = "ABCDEFG"
	tagNameForTest2    = "soracom-sdk-go-test-tag-name-1"
	tagValueForTest2   = "XYZ"
	tagValueForTest2_2 = ""
)

func TestPutSubscriberTag(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "active",
		Limit:        1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 active subscriber is required")
	}
	imsi := subs[0].IMSI

	tags := make([]Tag, 0, 10)
	tags = append(tags, Tag{TagName: tagNameForTest1, TagValue: tagValueForTest1})
	sub, err := apiClient.PutSubscriberTags(imsi, tags)
	if err != nil {
		t.Fatalf("Error occurred on SetSubscriberExpiredAt(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.Tags[tagNameForTest1] != tagValueForTest1 {
		t.Fatalf("Tag value '%s' could not be put on the subscriber with tag name '%s'", tagValueForTest1, tagNameForTest1)
	}

	tags = make([]Tag, 0, 10)
	tags = append(tags, Tag{TagName: tagNameForTest1, TagValue: tagValueForTest1_2})
	tags = append(tags, Tag{TagName: tagNameForTest2, TagValue: tagValueForTest2})
	sub, err = apiClient.PutSubscriberTags(imsi, tags)
	if err != nil {
		t.Fatalf("Error occurred on SetSubscriberExpiredAt(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.Tags[tagNameForTest1] != tagValueForTest1_2 {
		t.Fatalf("Tag value '%s' could not be put on the subscriber with tag name '%s'", tagValueForTest1_2, tagNameForTest1)
	}
	if sub.Tags[tagNameForTest2] != tagValueForTest2 {
		t.Fatalf("Tag value '%s' could not be put on the subscriber with tag name '%s'", tagValueForTest2, tagNameForTest2)
	}

	tags = make([]Tag, 0, 10)
	tags = append(tags, Tag{TagName: tagNameForTest2, TagValue: tagValueForTest2_2})
	sub, err = apiClient.PutSubscriberTags(imsi, tags)
	if err != nil {
		t.Fatalf("Error occurred on PutSubscriberTag(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.Tags[tagNameForTest1] != tagValueForTest1_2 {
		t.Fatalf("Tag value '%s' with tag name '%s' has been unexpectedly changed", tagValueForTest1_2, tagNameForTest1)
	}
	if sub.Tags[tagNameForTest2] != tagValueForTest2_2 {
		t.Fatalf("Tag value '%s' could not be put on the subscriber with tag name '%s'", tagValueForTest2, tagNameForTest2)
	}

	err = apiClient.DeleteSubscriberTag(imsi, tagNameForTest1)
	if err != nil {
		t.Fatalf("Error occurred on DeteleSubscriberTag(): %v", err.Error())
	}

	sub, err = apiClient.GetSubscriber(imsi)
	if err != nil {
		t.Fatalf("Error occurred on GetSubscriber(): %v", err.Error())
	}
	if sub.IMSI != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.Tags[tagNameForTest1] != "" {
		t.Fatalf("Tag value '%s' could not be deleted on the subscriber with tag name '%s'", tagValueForTest1_2, tagNameForTest1)
	}
	if sub.Tags[tagNameForTest2] != tagValueForTest2_2 {
		t.Fatalf("Tag value '%s' with tag name '%s' has been unexpectedly changed", tagValueForTest2, tagNameForTest2)
	}
}

func sameSubscribers(subs1, subs2 []Subscriber) bool {
	if len(subs1) != len(subs2) {
		return false
	}

	for i := range subs1 {
		if !sameSubscriber(&subs1[i], &subs2[i]) {
			return false
		}
	}

	return true
}

func sameSubscriber(sub1, sub2 *Subscriber) bool {
	return sub1.IMSI == sub2.IMSI
}

func TestGetAirStats(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "active",
		Limit:        1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 active subscriber is required")
	}
	imsi := subs[0].IMSI

	from := time.Now().AddDate(0, -6, 0)
	to := time.Now()
	stats, err := apiClient.GetAirStats(imsi, from, to, StatsPeriodMonth)
	if err != nil {
		t.Fatalf("GetAirStats() failed: %v", err.Error())
	}

	if len(stats) == 0 {
		t.Fatalf("TODO: ensure to find a subscriber with real stats")
	}
}

func TestGetBeamStats(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		TagName:           "soracom-sdk-go-test",
		TagValue:          "beam-stats",
		TagValueMatchMode: MatchModePrefix,
		Limit:             1,
	})
	if err != nil || len(subs) == 0 {
		t.Fatalf("At least 1 subscriber which has soracom-sdk-go-test tag with value 'beam-stats' is required")
	}
	imsi := subs[0].IMSI

	from := time.Now().AddDate(0, -6, 0)
	to := time.Now()
	stats, err := apiClient.GetBeamStats(imsi, from, to, StatsPeriodMonth)
	if err != nil {
		t.Fatalf("GetBeamStats() failed: %v", err.Error())
	}

	if len(stats) == 0 {
		t.Fatalf("TODO: ensure to find a subscriber with real stats")
	}
}

func TestExportAirStats(t *testing.T) {
	from := time.Now().AddDate(0, -6, 0)
	to := time.Now()
	_, err := apiClient.ExportAirStats(from, to, StatsPeriodMonth)
	if err != nil {
		t.Fatalf("ExportAirStats() failed: %v", err.Error())
	}
}

func TestExportBeamStats(t *testing.T) {
	from := time.Now().AddDate(0, -6, 0)
	to := time.Now()
	_, err := apiClient.ExportBeamStats(from, to, StatsPeriodMonth)
	if err != nil {
		t.Fatalf("ExportBeamStats() failed: %v", err.Error())
	}
}

func TestCreateGroup(t *testing.T) {
	g1, err := apiClient.CreateGroup(Tags{})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(g1.GroupID)

	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	g2, err := apiClient.CreateGroup(Tags{"name": name, "test-tag": "test-value"})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(g2.GroupID)
}

func TestCreateGroupWithName(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	group, err := apiClient.CreateGroupWithName(name)
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(group.GroupID)
	if group.Tags["name"] != name {
		t.Fatalf("Created a group with wrong name")
	}
}

func TestDeleteGroup(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	group, err := apiClient.CreateGroupWithName(name)
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}

	err = apiClient.DeleteGroup(group.GroupID)
	if err != nil {
		t.Fatalf("DeleteGroup() failed: %v", err.Error())
	}
}

func TestGetGroup(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	groupCreated, err := apiClient.CreateGroup(Tags{"name": name, "test-tag": "test-value"})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(groupCreated.GroupID)

	groupFound, err := apiClient.GetGroup(groupCreated.GroupID)
	if err != nil {
		t.Fatalf("GetGroup() failed: %v", err.Error())
	}

	if groupCreated.GroupID != groupFound.GroupID {
		t.Fatalf("Wrong group found")
	}
}

func TestListSubscribersInGroup(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	groupCreated, err := apiClient.CreateGroup(Tags{"name": name, "test-tag": "test-value"})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(groupCreated.GroupID)

	for _, cs := range createdSubscribers {
		_, err := apiClient.SetSubscriberGroup(cs.IMSI, groupCreated.GroupID)
		if err != nil {
			t.Fatalf("SetSubscriberGroup() failed: %v", err.Error())
		}
	}

	subs, _, err := apiClient.ListSubscribersInGroup(groupCreated.GroupID, nil)
	if err != nil {
		t.Fatalf("ListSubscribersInGroup() failed: %v", err.Error())
	}
	if len(subs) != len(createdSubscribers) {
		t.Fatalf("All subscribers should be in group %s: %v", groupCreated.GroupID, err.Error())
	}

	for _, cs := range createdSubscribers {
		_, err := apiClient.UnsetSubscriberGroup(cs.IMSI)
		if err != nil {
			t.Fatalf("SetSubscriberGroup() failed: %v", err.Error())
		}
	}

	subs, _, err = apiClient.ListSubscribersInGroup(groupCreated.GroupID, nil)
	if err != nil {
		t.Fatalf("ListSubscribersInGroup() failed: %v", err.Error())
	}
	if len(subs) != 0 {
		t.Fatalf("No subscribers should be in group %s: %v", groupCreated.GroupID, err.Error())
	}
}

func TestUpdateGroupConfigurations(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	group, err := apiClient.CreateGroup(Tags{"name": name, "test-tag": "test-value"})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(group.GroupID)

	beamTCPConfig := &BeamTCPConfig{
		Name:                "TCP Config Name 1",
		Destination:         "tcps://tcp.example.com:1234",
		Enabled:             true,
		AddSubscriberHeader: true,
		AddSignature:        true,
		PSK:                 "Pre-Shared Key",
	}

	g1, err := apiClient.UpdateGroupConfigurations(group.GroupID, "SoracomAir", []GroupConfig{
		{Key: "useCustomDns", Value: true},
		{Key: "dnsServers", Value: []string{"8.8.8.8", "8.8.4.4"}},
	})
	if err != nil {
		t.Fatalf("UpdateGroupConfigurations() failed: %v", err.Error())
	}
	airCfg := g1.Configuration["SoracomAir"].(map[string]interface{})
	if airCfg["useCustomDns"].(bool) != true {
		t.Fatalf("Unexpected value found")
	}
	if len(airCfg["dnsServers"].([]interface{})) != 2 {
		t.Fatalf("Unexpected value found")
	}

	g2, err := apiClient.UpdateGroupConfigurations(group.GroupID, "SoracomBeam", []GroupConfig{
		{Key: "tcp://beam.soracom.io:8023", Value: beamTCPConfig},
	})
	if err != nil {
		t.Fatalf("UpdateGroupConfigurations() failed: %v", err.Error())
	}
	beamCfg := g2.Configuration["SoracomBeam"].(map[string]interface{})
	cfg := beamCfg["tcp://beam.soracom.io:8023"].(map[string]interface{})
	cfgName := cfg["name"].(string)
	if cfgName != "TCP Config Name 1" {
		t.Fatalf("Unexpected value found")
	}
}

func TestUpdateAirConfig(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	group, err := apiClient.CreateGroup(Tags{"name": name, "test-tag": "test-value"})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(group.GroupID)

	airConfig1 := &AirConfig{
		UseCustomDNS: true,
		DNSServers: []string{
			"8.8.8.8",
			"8.8.4.4",
		},
	}

	g1, err := apiClient.UpdateAirConfig(group.GroupID, airConfig1)
	if err != nil {
		t.Fatalf("UpdateAirConfig() failed: %v", err.Error())
	}
	air1 := g1.Configuration["SoracomAir"].(map[string]interface{})
	if air1["useCustomDns"].(bool) != true {
		t.Fatalf("Unexpected value found")
	}
	if len(air1["dnsServers"].([]interface{})) != 2 {
		t.Fatalf("Unexpected value found")
	}

	airConfig2 := &AirConfig{
		UseCustomDNS: false,
		DNSServers: []string{
			"0.0.0.0",
		},
	}

	g2, err := apiClient.UpdateAirConfig(group.GroupID, airConfig2)
	if err != nil {
		t.Fatalf("UpdateAirConfig() failed: %v", err.Error())
	}
	air2 := g2.Configuration["SoracomAir"].(map[string]interface{})
	if air2["useCustomDns"].(bool) != false {
		t.Fatalf("Unexpected value found")
	}
	if len(air2["dnsServers"].([]interface{})) != 1 {
		t.Fatalf("Unexpected value found")
	}
}

func TestUpdateBeamTCPConfig(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	group, err := apiClient.CreateGroup(Tags{"name": name, "test-tag": "test-value"})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(group.GroupID)

	entryPoint1 := "tcp://beam.soracom.io:8023"
	beamTCPConfig1 := &BeamTCPConfig{
		Name:                "TCP Config Name 1",
		Destination:         "tcps://tcp.example.com:1234",
		Enabled:             true,
		AddSubscriberHeader: true,
		AddSignature:        true,
		PSK:                 "Pre-Shared Key",
	}

	g1, err := apiClient.UpdateBeamTCPConfig(group.GroupID, entryPoint1, beamTCPConfig1)
	if err != nil {
		t.Fatalf("UpdateBeamTCPConfig() failed: %v", err.Error())
	}
	beam1 := g1.Configuration["SoracomBeam"].(map[string]interface{})
	cfg1 := beam1["tcp://beam.soracom.io:8023"].(map[string]interface{})
	cfgName := cfg1["name"].(string)
	if cfgName != "TCP Config Name 1" {
		t.Fatalf("Unexpected value found")
	}
}

func TestDeleteGroupConfiguration(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	group, err := apiClient.CreateGroup(Tags{"name": name, "test-tag": "test-value"})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(group.GroupID)

	entryPoint1 := "tcp://beam.soracom.io:8023"
	beamTCPConfig1 := &BeamTCPConfig{
		Name:                "TCP Config Name 1",
		Destination:         "tcps://tcp.example.com:1234",
		Enabled:             true,
		AddSubscriberHeader: true,
		AddSignature:        true,
		PSK:                 "Pre-Shared Key",
	}

	g1, err := apiClient.UpdateBeamTCPConfig(group.GroupID, entryPoint1, beamTCPConfig1)
	if err != nil {
		t.Fatalf("UpdateBeamTCPConfig() failed: %v", err.Error())
	}
	beam1 := g1.Configuration["SoracomBeam"].(map[string]interface{})
	cfg1 := beam1[entryPoint1].(map[string]interface{})
	cfgName := cfg1["name"].(string)
	if cfgName != "TCP Config Name 1" {
		t.Fatalf("Unexpected value found")
	}

	g2, err := apiClient.DeleteGroupConfiguration(group.GroupID, "SoracomBeam", entryPoint1)
	if err != nil {
		t.Fatalf("DeleteGroupConfiguration() failed: %v", err.Error())
	}
	if g2.Configuration["SoracomBeam"] != nil {
		t.Fatalf("Group configuration was not deleted unexpectedly")
	}
}

func TestUpdateGroupTags(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	group, err := apiClient.CreateGroup(Tags{"name": name, "test-tag": "test-value"})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(group.GroupID)

	g, err := apiClient.UpdateGroupTags(group.GroupID, []Tag{
		{TagName: "name1", TagValue: "value1"},
		{TagName: "name2", TagValue: "value2"},
	})
	if err != nil {
		t.Fatalf("UpdateGroupTags() failed: %v", err.Error())
	}

	if len(g.Tags) != 4 {
		t.Fatalf("The group should have 4 tags (name, test-tag, name1, name2)")
	}
}

func TestDeleteGroupTag(t *testing.T) {
	name := fmt.Sprintf("group-name-for-test-%d", time.Now().Unix())
	group, err := apiClient.CreateGroup(Tags{"name": name, "test-tag": "test-value"})
	if err != nil {
		t.Fatalf("CreateGroup() failed: %v", err.Error())
	}
	defer apiClient.DeleteGroup(group.GroupID)

	g1, err := apiClient.UpdateGroupTags(group.GroupID, []Tag{
		{TagName: "name1", TagValue: "value1"},
		{TagName: "name2", TagValue: "value2"},
	})
	if err != nil {
		t.Fatalf("UpdateGroupTags() failed: %v", err.Error())
	}

	if len(g1.Tags) != 4 {
		t.Fatalf("The group should have 4 tags (name, test-tag, name1, name2)")
	}

	err = apiClient.DeleteGroupTag(group.GroupID, "name1")
	if err != nil {
		t.Fatalf("DeleteGroupTag() failed: %v", err.Error())
	}

	g2, err := apiClient.GetGroup(g1.GroupID)
	if err != nil {
		t.Fatalf("GetGroup() failed: %v", err.Error())
	}

	if len(g2.Tags) != 3 {
		t.Fatalf("The group should now have 3 tag")
	}
}

func TestCreateEventHandler(t *testing.T) {
	imsi := createdSubscribers[0].IMSI

	o := &CreateEventHandlerOptions{
		TargetIMSI:  &imsi,
		Status:      EventStatusActive,
		Name:        "Test Event handler Name",
		Description: "Test Event Handler Description",
		RuleConfig:  RuleDailyTraffic(1000, EventDateTimeBeginningOfNextMonth),
		ActionConfigList: []ActionConfig{
			ActionChangeSpeed(EventDateTimeImmediately, SpeedClassS1Minimum),
			ActionWebHook(EventDateTimeImmediately, ActionWebhookProperty{
				URL:         "https://example.com/my/api",
				Method:      http.MethodPost,
				ContentType: "application/json",
				Body:        `{"message":"Hello world"}`,
			}),
			ActionWebHook(EventDateTimeImmediately, ActionWebhookProperty{
				URL:         "https://example.com/my/api",
				Method:      http.MethodGet,
				ContentType: "application/json",
			}),
			ActionDeactivate(EventDateTimeImmediately),
			ActionActivate(EventDateTimeBeginningOfNextMonth),
		},
	}
	eh, err := apiClient.CreateEventHandler(o)
	if err != nil {
		t.Fatalf("CreateEventHandler() failed: %v", err.Error())
	}

	if eh.HandlerID == "" {
		t.Fatalf("CreateEventHandler() failed: has not HandlerID")
	}
	if eh.Name != o.Name {
		t.Fatalf("CreateEventHandler() failed: unmatch handler name want: %s, got:%s", o.Name, eh.Name)
	}

	if eh.Description != o.Description {
		t.Fatalf("CreateEventHandler() failed: unmatch handler Description want: %s, got:%s", o.Description, eh.Description)
	}
}

func TestListEventHandlers(t *testing.T) {
	eventHandlers, err := apiClient.ListEventHandlers(nil)
	if err != nil {
		t.Fatalf("ListEventHandlers() failed: %v", err.Error())
	}
	if len(eventHandlers) == 0 {
		t.Fatalf("At least 1 event handler is required.")
	}
}

func TestListEventHandlersForSubscriber(t *testing.T) {
	imsi := createdSubscribers[0].IMSI

	eventHandlers, err := apiClient.ListEventHandlersForSubscriber(imsi)
	if err != nil {
		t.Fatalf("ListEventHandlersForSubscriber() failed: %v", err.Error())
	}

	if len(eventHandlers) == 0 {
		t.Fatalf("At least 1 event handler is required for the subscriber.")
	}
}

func TestGetEventHandler(t *testing.T) {
	imsi := createdSubscribers[0].IMSI

	eventHandlers, err := apiClient.ListEventHandlersForSubscriber(imsi)
	if err != nil {
		t.Fatalf("ListEventHandlersForSubscriber() failed: %v", err.Error())
	}

	id := eventHandlers[len(eventHandlers)-1].HandlerID
	_, err = apiClient.GetEventHandler(id)
	if err != nil {
		t.Fatalf("GetEventHandler() failed: %v", err.Error())
	}
}

func TestUpdateEventHandler(t *testing.T) {
	imsi := createdSubscribers[0].IMSI

	eventHandlers, err := apiClient.ListEventHandlersForSubscriber(imsi)
	if err != nil {
		t.Fatalf("ListEventHandlersForSubscriber() failed: %v", err.Error())
	}

	id := eventHandlers[len(eventHandlers)-1].HandlerID
	eh, err := apiClient.GetEventHandler(id)
	if err != nil {
		t.Fatalf("GetEventHandler() failed: %v", err.Error())
	}

	updatedName := fmt.Sprintf("Updated Name %d", time.Now().Unix())
	eh.Name = updatedName
	err = apiClient.UpdateEventHandler(eh)
	if err != nil {
		t.Fatalf("UpdateEventHandler() failed: %v", err.Error())
	}

	eh2, err := apiClient.GetEventHandler(id)
	if err != nil {
		t.Fatalf("GetEventHandler() failed: %v", err.Error())
	}

	if eh2.Name != updatedName {
		t.Fatalf("Event handler has not been updated correctly")
	}
}

func TestDeleteEventHandler(t *testing.T) {
	imsi := createdSubscribers[0].IMSI

	eventHandlers, err := apiClient.ListEventHandlersForSubscriber(imsi)
	if err != nil {
		t.Fatalf("ListEventHandlersForSubscriber() failed: %v", err.Error())
	}

	id := eventHandlers[len(eventHandlers)-1].HandlerID
	err = apiClient.DeleteEventHandler(id)
	if err != nil {
		t.Fatalf("DeleteEventHandler() failed: %v", err.Error())
	}
}

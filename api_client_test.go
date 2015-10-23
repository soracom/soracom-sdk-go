package soracom

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

var (
	apiClient *APIClient
)

func TestAuth(t *testing.T) {
	email := os.Getenv("SORACOM_EMAIL")
	password := os.Getenv("SORACOM_PASSWORD")
	endpoint := os.Getenv("SORACOM_ENDPOINT")

	if email == "" {
		t.Fatal("SORACOM_EMAIL env var is required")
	}

	if password == "" {
		t.Fatal("SORACOM_PASSWORD env var is required")
	}

	options := &APIClientOptions{
		Endpoint: endpoint,
	}
	apiClient = NewAPIClient(options)

	err := apiClient.Auth(email, password)
	if err != nil {
		t.Fatalf("Auth() failed: %v", err.Error())
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
	if err != nil {
		t.Fatalf("At least 1 subscriber is required")
	}
	imsi := subs[0].Imsi
	sub, err := apiClient.GetSubscriber(imsi)
	if err != nil {
		t.Fatalf("Error occurred on GetSubscriber(): %v", err.Error())
	}
	if sub.Imsi != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
}

func TestUpdateSubscriberSpeedClass(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		Limit: 1,
	})
	if err != nil {
		t.Fatalf("At least 1 subscriber is required")
	}
	imsi := subs[0].Imsi
	sub, err := apiClient.UpdateSubscriberSpeedClass(imsi, "s1.minimum")
	if err != nil {
		t.Fatalf("Error occurred on UpdateSubscriberSpeedClass(): %v", err.Error())
	}
	if sub.Imsi != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.SpeedClass != "s1.minimum" {
		t.Fatalf("Found a subscriber speed class which is not specified")
	}

	sub, err = apiClient.UpdateSubscriberSpeedClass(imsi, "s1.fast")
	if err != nil {
		t.Fatalf("Error occurred on UpdateSubscriberSpeedClass(): %v", err.Error())
	}
	if sub.Imsi != imsi {
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
	if err != nil {
		t.Fatalf("At least 1 inactive subscriber is required")
	}
	imsi := subs[0].Imsi
	sub, err := apiClient.ActivateSubscriber(imsi)
	if err != nil {
		t.Fatalf("Error occurred on ActivateSubscriber(): %v", err.Error())
	}
	if sub.Imsi != imsi {
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
	if err != nil {
		t.Fatalf("At least 1 active subscriber is required")
	}
	imsi := subs[0].Imsi
	sub, err := apiClient.DeactivateSubscriber(imsi)
	if err != nil {
		t.Fatalf("Error occurred on DeactivateSubscriber(): %v", err.Error())
	}
	if sub.Imsi != imsi {
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
	if err != nil {
		t.Fatalf("At least 1 non-terminated subscriber is required")
	}
	imsi := subs[0].Imsi
	sub, err := apiClient.EnableSubscriberTermination(imsi)
	if err != nil {
		t.Fatalf("Error occurred on EnableSubscriberTermination(): %v", err.Error())
	}
	if sub.Imsi != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if !sub.TerminationEnabled {
		t.Fatalf("Termination must be enabled")
	}

	sub, err = apiClient.DisableSubscriberTermination(imsi)
	if err != nil {
		t.Fatalf("Error occurred on DisableSubscriberTermination(): %v", err.Error())
	}
	if sub.Imsi != imsi {
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
	if sub.Imsi != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if !sub.TerminationEnabled {
		t.Fatalf("Termination must be enabled")
	}

	sub, err = apiClient.DisableSubscriberTermination(imsi)
	if err != nil {
		t.Fatalf("Error occurred on DisableSubscriberTermination(): %v", err.Error())
	}
	if sub.Imsi != imsi {
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
}

func TestSetSubscriberExpiryTime(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "active",
		Limit:        1,
	})
	if err != nil {
		t.Fatalf("At least 1 active subscriber is required")
	}
	imsi := subs[0].Imsi
	tomorrow := time.Now().AddDate(0, 0, 1)
	sub, err := apiClient.SetSubscriberExpiryTime(imsi, tomorrow)
	if err != nil {
		t.Fatalf("Error occurred on SetSubscriberExpiryTime(): %v", err.Error())
	}
	if sub.Imsi != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.ExpiryTime.UnixMilli() != (tomorrow.UnixNano() / 1000 / 1000) {
		fmt.Printf("sub.ExpiryTime.Time == %s, tomorrow == %s", sub.ExpiryTime.Time, tomorrow)
		t.Fatalf("Expiry time for a subscriber has not been updated")
	}
}

func TestUnsetSubscriberExpiryTime(t *testing.T) {
	subs, _, err := apiClient.ListSubscribers(&ListSubscribersOptions{
		StatusFilter: "active",
		Limit:        1,
	})
	if err != nil {
		t.Fatalf("At least 1 active subscriber is required")
	}
	imsi := subs[0].Imsi
	tomorrow := time.Now().AddDate(0, 0, 1)
	sub, err := apiClient.SetSubscriberExpiryTime(imsi, tomorrow)
	if err != nil {
		t.Fatalf("Error occurred on SetSubscriberExpiryTime(): %v", err.Error())
	}
	if sub.Imsi != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.ExpiryTime.UnixMilli() != (tomorrow.UnixNano() / 1000 / 1000) {
		fmt.Printf("sub.ExpiryTime.Time == %s, tomorrow == %s", sub.ExpiryTime.Time, tomorrow)
		t.Fatalf("Expiry time for a subscriber has not been updated")
	}

	sub, err = apiClient.UnsetSubscriberExpiryTime(imsi)
	if err != nil {
		t.Fatalf("Error occurred on UnsetSubscriberExpiryTime(): %v", err.Error())
	}
	if sub.Imsi != imsi {
		t.Fatalf("Found a subscriber which is not specified")
	}
	if sub.ExpiryTime != nil {
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
	if err != nil {
		t.Fatalf("At least 1 active subscriber is required")
	}
	imsi := subs[0].Imsi

	tags := make([]Tag, 0, 10)
	tags = append(tags, Tag{TagName: tagNameForTest1, TagValue: tagValueForTest1})
	sub, err := apiClient.PutSubscriberTags(imsi, tags)
	if err != nil {
		t.Fatalf("Error occurred on SetSubscriberExpiryTime(): %v", err.Error())
	}
	if sub.Imsi != imsi {
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
		t.Fatalf("Error occurred on SetSubscriberExpiryTime(): %v", err.Error())
	}
	if sub.Imsi != imsi {
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
	if sub.Imsi != imsi {
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
	if sub.Imsi != imsi {
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
	return sub1.Imsi == sub2.Imsi
}

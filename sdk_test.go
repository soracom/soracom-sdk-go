package soracom

import (
	"github.com/soracom/soracom-sdk-go"
	"os"
	"testing"
)

var (
	apiClient *soracom.ApiClient
)

func TestAuth(t *testing.T) {
	email := os.Getenv("SORACOM_EMAIL")
	password := os.Getenv("SORACOM_PASSWORD")

	if email == "" {
		t.Fatal("SORACOM_EMAIL env var is required")
	}

	if password == "" {
		t.Fatal("SORACOM_PASSWORD env var is required")
	}

	apiClient = soracom.NewApiClient()

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

	options := &soracom.ListSubscribersOptions{
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

	options2 := &soracom.ListSubscribersOptions{
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
	options := &soracom.ListSubscribersOptions{
		TagName:  "name",
		TagValue: "hoge",
	}
	subs, _, err := apiClient.ListSubscribers(options)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers({TagName: 'name', TagValue: 'hoge'})")
	}
	if len(subs) < 1 {
		t.Fatalf("5+ subscribers with name 'hoge' are required")
	}
	//t.Logf("subs == %v", subs)

	options2 := &soracom.ListSubscribersOptions{
		TagName:  "name",
		TagValue: "hoge",
		Limit:    3,
	}
	subs2, pk2, err := apiClient.ListSubscribers(options2)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers({TagName: 'name', TagValue: 'hoge', Limit: 3})")
	}
	if len(subs2) != 3 {
		t.Fatalf("5+ subscribers with name 'hoge' are required")
	}
	if pk2 == nil {
		t.Fatalf("Pagination keys must be returned if Limit option is specified and there are subscribers more than the limit")
	}
	for i := range subs2 {
		if subs2[i].Tags["name"] != "hoge" {
			t.Fatalf("Found a subscriber with the name which is not specified")
		}
	}
	//t.Logf("subs2 == %v", subs2)
	t.Logf("pk2 == %v", pk2)

	options3 := &soracom.ListSubscribersOptions{
		TagName:          "name",
		TagValue:         "hoge",
		Limit:            3,
		LastEvaluatedKey: pk2.Next,
	}
	subs3, _, err := apiClient.ListSubscribers(options3)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs3) == 0 {
		t.Fatalf("5+ subscribers with name 'hoge' are required")
	}
	if sameSubscribers(subs2, subs3) {
		t.Fatalf("Pagination seems not working with tag search")
	}
	for i := range subs3 {
		if subs3[i].Tags["name"] != "hoge" {
			t.Fatalf("Found a subscriber with the name which is not specified")
		}
	}
	//t.Logf("subs3 == %v", subs3)
}

func TestListSubscribersWithStatusFilter(t *testing.T) {
	options := &soracom.ListSubscribersOptions{
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

	options2 := &soracom.ListSubscribersOptions{
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

	options3 := &soracom.ListSubscribersOptions{
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
	options := &soracom.ListSubscribersOptions{
		StatusFilter: "active|inactive",
		TagName:      "name",
		TagValue:     "hoge",
		Limit:        3,
	}
	subs, pk, err := apiClient.ListSubscribers(options)
	if err != nil {
		t.Fatalf("Error occurred on ListSubscribers()")
	}
	if len(subs) != 3 {
		t.Fatalf("5+ subscribers have name 'hoge' and with active or inactive status are required")
	}
	//t.Logf("subs == %v", subs)

	options2 := &soracom.ListSubscribersOptions{
		StatusFilter:     "active|inactive",
		TagName:          "name",
		TagValue:         "hoge",
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

func sameSubscribers(subs1, subs2 []soracom.Subscriber) bool {
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

func sameSubscriber(sub1, sub2 *soracom.Subscriber) bool {
	return sub1.Imsi == sub2.Imsi
}

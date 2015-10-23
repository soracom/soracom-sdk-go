package main

import (
	"fmt"
	"os"

	"github.com/k0kubun/pp"
	"github.com/soracom/soracom-sdk-go"
)

func main() {
	email := os.Getenv("SORACOM_EMAIL")
	password := os.Getenv("SORACOM_PASSWORD")

	if email == "" {
		fmt.Println("SORACOM_EMAIL env var is required")
		os.Exit(1)
		return
	}

	if password == "" {
		fmt.Println("SORACOM_PASSWORD env var is required")
		os.Exit(1)
		return
	}

	ac := soracom.NewAPIClient(nil)

	err := ac.Auth(email, password)
	if err != nil {
		fmt.Printf("auth err: %v\n", err.Error())
		return
	}

	subscribers, lek, err := ac.ListSubscribers(nil)
	if err != nil {
		fmt.Printf("err: %v\n", err.Error())
		return
	}

	pp.Print(subscribers)
	pp.Print(lek)
}

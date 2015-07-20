package main

import (
	"fmt"
	"os"

	"github.com/k0kubun/pp"
	"github.com/soracom/soracom-sdk-go"
)

func main() {
	client := soracom.NewClient()

	email := os.Getenv("SORACOM_EMAIL")
	password := os.Getenv("SORACOM_PASSWORD")

	err := client.Auth(email, password)
	if err != nil {
		fmt.Printf("auth err: %v\n", err.Error())
		return
	}

	subscribers, err := client.ListSubscribers()
	if err != nil {
		fmt.Printf("err: %v\n", err.Error())
		return
	}

	pp.Print(subscribers)
}

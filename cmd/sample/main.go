package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/codematic/opencdp-go/cdp"
)

func main() {
	apiKey := os.Getenv("CDP_API_KEY")
	if apiKey == "" {
		fmt.Println("Please set CDP_API_KEY environment variable")
		return
	}

	// Initialize Client
	config := cdp.CDPConfig{
		CDPAPIKey:             apiKey,
		Debug:                 true,
		FailOnException:       true,
		MaxConcurrentRequests: 5,
		Timeout:               5000,
	}

	client := cdp.NewClient(config)
	defer client.Close()

	ctx := context.Background()

	// 1. Ping
	fmt.Println("Pinging service...")
	if err := client.Ping(ctx); err != nil {
		fmt.Printf("Ping failed: %v\n", err)
		return
	}
	fmt.Println("Ping successful")

	// 2. Identify
	fmt.Println("Identifying user...")
	userID := "user_12345"
	traits := map[string]interface{}{
		"name":  "John Doe",
		"email": "john.doe@example.com",
		"plan":  "premium",
	}
	if err := client.Identify(ctx, userID, traits); err != nil {
		fmt.Printf("Identify failed: %v\n", err)
	} else {
		fmt.Println("Identify successful")
	}

	// 3. Track
	fmt.Println("Tracking event...")
	props := map[string]interface{}{
		"item_id": "prod_999",
		"price":   49.99,
	}
	if err := client.Track(ctx, userID, "item_purchased", props); err != nil {
		fmt.Printf("Track failed: %v\n", err)
	} else {
		fmt.Println("Track successful")
	}

	// 4. Send Email
	fmt.Println("Sending email...")
	emailPayload := cdp.EmailPayload{
		To: "john.doe@example.com",
		Identifiers: cdp.Identifiers{
			ID: userID,
		},
		TransactionalMessageID: "WELCOME_EMAIL",
		Subject:                "Welcome to the platform!",
		Body:                   "Thanks for joining.",
	}
	if err := client.SendEmail(ctx, emailPayload); err != nil {
		fmt.Printf("SendEmail failed: %v\n", err)
	} else {
		fmt.Println("SendEmail successful")
	}

	// Wait a bit to ensure logs flush if async (though this SDK is blocking)
	time.Sleep(1 * time.Second)
}

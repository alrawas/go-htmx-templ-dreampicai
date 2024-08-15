package main

import (
	"context"
	"fmt"
	"log"

	"github.com/joho/godotenv"
	"github.com/replicate/replicate-go"
)

func main() {

	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	ctx := context.Background()

	// You can also provide a token directly with
	// `replicate.NewClient(replicate.WithToken("r8_..."))`
	r8, err := replicate.NewClient(replicate.WithTokenFromEnv())
	if err != nil {
		// handle error
		log.Fatal("r8 client error: ", err)
	}

	model := "stability-ai/sdxl"
	version := "7762fd07cf82c948538e41f63f77d685e02b063e37e496e96eefd46c929f9bdc"
	var _ = model

	// Create a prediction input
	input := replicate.PredictionInput{
		"prompt": "An astronaut riding a rainbow unicorn",
	}

	webhook := replicate.Webhook{
		URL:    "https://webhook.site/ba692ef2-9819-4f8c-a531-696b8174e8d9",
		Events: []replicate.WebhookEventType{"start", "completed"},
	}

	// Run a model and wait for its output
	// output, err := r8.Run(ctx, fmt.Sprintf("%s:%s", model, version), input, &webhook)
	output, err := r8.CreatePrediction(ctx, version, input, &webhook, false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output: ", output)
}

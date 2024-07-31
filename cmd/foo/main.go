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
		log.Fatal(err)
	}
	ctx := context.TODO()

	// You can also provide a token directly with
	// `replicate.NewClient(replicate.WithToken("r8_..."))`
	r8, err := replicate.NewClient(replicate.WithTokenFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	// model := "fofr/sticker-maker"
	version := "4acb778eb059772225ec213948f0660867b2e03f277448f18cf1800b96a65a1a"

	input := replicate.PredictionInput{
		"prompt": "An astronaut riding a rainbow unicorn",
	}

	webhook := replicate.Webhook{
		URL:    "https://webhook.site/c214eb53-4909-4cd0-ab2e-5ffad90cdbce",
		Events: []replicate.WebhookEventType{"start", "completed"},
	}
	// Run a model and wait for its output
	output, err := r8.CreatePrediction(ctx, version, input, &webhook, false)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("output", output)
}

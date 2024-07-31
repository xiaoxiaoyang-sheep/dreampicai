package handler

import (
	"context"
	"database/sql"
	"dreampicai/db"
	"dreampicai/pkg/kit/validate"
	"dreampicai/types"
	"dreampicai/view/generate"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/replicate/replicate-go"
	"github.com/uptrace/bun"
)

const creditsPerImage = 2

func HandleGenerateIndex(w http.ResponseWriter, r *http.Request) error {
	user := getAuthenticatedUser(r)
	images, err := db.GetImagesByUserID(user.ID)
	if err != nil {
		return err
	}
	data := generate.ViewData{
		Images: images,
	}
	return render(r, w, generate.Index(data))
}

func HandleGenerateCreate(w http.ResponseWriter, r *http.Request) error {
	user := getAuthenticatedUser(r)
	amount, _ := strconv.Atoi(r.FormValue("amount"))
	params := generate.FormParams{
		Prompt: r.FormValue("prompt"),
		Amount: amount,
	}
	var errors generate.FormErrors
	if amount <= 0 || amount > 8 {
		errors.Amount = "Please enter a valid amount"
		return render(r, w, generate.Form(params, errors))
	}

	ok := validate.New(params, validate.Fields{
		"Prompt": validate.Rules(validate.Min(10), validate.Max(100)),
	}).Validate(&errors)
	if !ok {
		return render(r, w, generate.Form(params, errors))
	}

	creditsNeeded := params.Amount * creditsPerImage
	if user.Account.Credits < creditsNeeded {
		errors.CreditsNeeded = creditsNeeded
		errors.UserCredits = user.Account.Credits
		errors.Credits = true
		return render(r, w, generate.Form(params, errors))
	}

	user.Account.Credits -= creditsNeeded
	if err := db.UpdateAccount(&user.Account); err != nil {
		return err
	}

	batchID := uuid.New()
	genParams := GenerateImageParams{
		Amount:  params.Amount,
		UserID:  user.ID,
		BatchID: batchID,
	}
	if err := generateImages(r.Context(), genParams); err != nil {
		return err
	}

	err := db.Bun.RunInTx(r.Context(), &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		for i := 0; i < params.Amount; i++ {
			img := types.Image{
				Prompt:  params.Prompt,
				UserID:  user.ID,
				Status:  types.ImageStatusPending,
				BatchID: batchID,
			}
			if err := db.CreateImage(&img); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	return hxRedirect(w, r, "/generate")
}

func HandleGenerateImageStatus(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		return err
	}
	image, err := db.GetImageByID(id)
	if err != nil {
		return err
	}

	slog.Info("checking image status", "id", id)
	return render(r, w, generate.GalleryImage(image))
}

type GenerateImageParams struct {
	Prompt  string
	Amount  int
	BatchID uuid.UUID
	UserID  uuid.UUID
}

func generateImages(ctx context.Context, params GenerateImageParams) error {
	r8, err := replicate.NewClient(replicate.WithTokenFromEnv())
	if err != nil {
		log.Fatal(err)
	}

	// model := "fofr/sticker-maker"
	version := "4acb778eb059772225ec213948f0660867b2e03f277448f18cf1800b96a65a1a"

	input := replicate.PredictionInput{
		"prompt": params.Prompt,
	}

	webhook := replicate.Webhook{
		URL:    fmt.Sprintf("https://webhook.site/c214eb53-4909-4cd0-ab2e-5ffad90cdbce/%s/%s", params.UserID, params.BatchID),
		Events: []replicate.WebhookEventType{"completed"},
	}
	// Run a model and wait for its output
	_, err = r8.CreatePrediction(ctx, version, input, &webhook, false)
	return err
}

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

func HandleGeneratePost(w http.ResponseWriter, r *http.Request) error {
	user := getAuthenticatedUser(r)
	amount, _ := strconv.Atoi(r.FormValue("amount"))
	params := generate.FormParams{
		Prompt: r.FormValue("prompt"),
		Amount: amount,
	}
	var errors generate.FormErrors

	if amount <= 0 || amount > 8 {
		errors.Amount = "Amount must be between 1 and 10"
		return render(r, w, generate.Form(params, errors))
	}

	ok := validate.New(params, validate.Fields{
		"Prompt": validate.Rules(validate.Min(10), validate.Max(100)),
	}).Validate(&errors)

	if !ok {
		return render(r, w, generate.Form(params, errors))
	}

	// credits check

	creditsNeeded := params.Amount * creditsPerImage
	if user.Account.Credits < creditsNeeded {
		errors.CreditsNeeded = creditsNeeded
		errors.UserCredits = user.Credits
		errors.Credits = true
		return render(r, w, generate.Form(params, errors))
	}

	user.Account.Credits -= creditsNeeded
	if err := db.UpdateAccount(&user.Account); err != nil {
		return err
	}

	batchID := uuid.New()
	genParams := GenerateImageParams{
		Prompt:  params.Prompt,
		Amount:  params.Amount,
		UserID:  user.ID,
		BatchID: batchID,
	}

	if err := generateImages(r.Context(), genParams); err != nil {
		return err
	}

	// ability to rollback in case of error
	err := db.Bun.RunInTx(r.Context(), &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		for i := 0; i < params.Amount; i++ {
			img := types.Image{
				UserID:  user.ID,
				Status:  types.ImageStatusPending,
				BatchID: batchID,
			}
			if err := db.CreateImage(tx, &img); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	// trick because old swap to #gallery in case successfull but in case of error swaps in the form
	// check response-targets htmx extension. can specify target but not the swap way
	return hxRedirect(w, r, "/generate")
}
func HandleGenerateImageDelete(w http.ResponseWriter, r *http.Request) error {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))

	if err != nil {
		slog.Error("failed to parse id", err)
		hxRedirect(w, r, "/generate")
		return err
	}

	// delete image from db
	if err := db.SoftDeleteImage(id); err != nil {
		slog.Error("failed to soft delete image", err)
		hxRedirect(w, r, "/generate")
		return err
	}
	// redirect to generate page
	hxRedirect(w, r, "/generate")
	return nil
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
	Amount  int
	Prompt  string
	BatchID uuid.UUID
	UserID  uuid.UUID
}

func generateImages(ctx context.Context, params GenerateImageParams) error {

	r8, err := replicate.NewClient(replicate.WithTokenFromEnv())
	if err != nil {
		log.Fatal("r8 client error: ", err)
	}

	model := "stability-ai/sdxl"
	version := "7762fd07cf82c948538e41f63f77d685e02b063e37e496e96eefd46c929f9bdc"
	var _ = model

	// Create a prediction input
	input := replicate.PredictionInput{
		"prompt":      params.Prompt,
		"num_outputs": params.Amount,
	}

	webhook := replicate.Webhook{
		URL:    fmt.Sprintf("https://webhook.site/ba692ef2-9819-4f8c-a531-696b8174e8d9/%s/%s", params.UserID, params.BatchID), // take this from .env file replicate_callback_url
		Events: []replicate.WebhookEventType{"completed"},
	}

	// Run a model and wait for its output
	_, err = r8.CreatePrediction(ctx, version, input, &webhook, false)
	return err
}

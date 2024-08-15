package handler

import (
	"context"
	"database/sql"
	"dreampicai/db"
	"dreampicai/types"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

const (
	succeeded  = "succeeded"
	processing = "processing"
)

type ReplicateResp struct {
	Input struct {
		Prompt string `json:"prompt"`
	} `json:"input"`
	Status string   `json:"status"`
	Output []string `json:"output"`
}

func HandleReplicateCallback(w http.ResponseWriter, r *http.Request) error {
	var resp ReplicateResp
	if err := json.NewDecoder(r.Body).Decode(&resp); err != nil {
		return err
	}

	if resp.Status == processing {
		return nil
	}

	if resp.Status != succeeded {
		return fmt.Errorf("replicate callback failed with status %s", resp.Status)
	}

	batchID, err := uuid.Parse(chi.URLParam(r, "batchID"))
	if err != nil {
		return fmt.Errorf("replicate callback failed to parse batchID: %w", err)
	}
	_ = batchID

	images, err := db.GetImagesByBatchID(batchID)
	if err != nil {
		return fmt.Errorf("replicate callback failed to get images with batchID: %s: %w", batchID, err)
	}

	if len(images) != len(resp.Output) {
		return fmt.Errorf("replicate callback failed, expected %d images, got %d", len(images), len(resp.Output))
	}

	err = db.Bun.RunInTx(r.Context(), &sql.TxOptions{}, func(ctx context.Context, tx bun.Tx) error {
		for i, imageURL := range resp.Output {
			images[i].Status = types.ImageStatusCompleted
			images[i].ImageLocation = imageURL
			images[i].Prompt = resp.Input.Prompt
			if err := db.UpdateImage(tx, &images[i]); err != nil {
				// return err
				return fmt.Errorf("replicate callback failed to update image: %w", err)
			}
		}
		return nil
	})
	// do something with the output
	// fmt.Println(resp)
	return err
}

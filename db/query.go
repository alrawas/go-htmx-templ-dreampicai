package db

import (
	"context"
	"dreampicai/types"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

// pointer only for creation
func CreateImage(tx bun.Tx, image *types.Image) error {
	_, err := tx.NewInsert().
		Model(image).
		Exec(context.Background())
	return err
}

func GetImageByID(id int) (types.Image, error) {
	var image types.Image
	err := Bun.NewSelect().
		Model(&image).
		Where("id = ?", id).
		Scan(context.Background())
	return image, err
}

func GetImagesByBatchID(batchID uuid.UUID) ([]types.Image, error) {
	var images []types.Image
	err := Bun.NewSelect().
		Model(&images).
		Where("batch_id = ?", batchID).
		Scan(context.Background())
	return images, err
}

func GetImagesByUserID(userID uuid.UUID) ([]types.Image, error) {
	var images []types.Image
	err := Bun.NewSelect().
		Model(&images).
		Where("deleted = ?", false).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Scan(context.Background())
	return images, err
}

func SoftDeleteImage(id int) error {
	_, err := Bun.NewUpdate().
		Model((*types.Image)(nil)).
		Where("id = ?", id).
		// Where("deleted = ?", false).
		Set("deleted = ?", true).
		Exec(context.Background())
	return err
}

func UpdateImage(tx bun.Tx, image *types.Image) error {
	_, err := tx.NewUpdate().
		Model(image).
		WherePK().
		Exec(context.Background())
	return err
}

// we accept pointers but we don't return them to avoid pointer dereferencing and we cannot recover from that
func UpdateAccount(account *types.Account) error {
	_, err := Bun.NewUpdate().
		Model(account).
		WherePK().
		Exec(context.Background())
	return err
}
func GetAccountByUserID(userID uuid.UUID) (types.Account, error) {
	var account types.Account
	err := Bun.NewSelect().
		Model(&account).
		Where("user_id = ?", userID).
		Scan(context.Background())
	return account, err
}

// create account pass pointer is ok but get account no
// usually you pass the context
// in db not used usually this is why just use contxt background
// unless slow query and you really need to pass a context
// - gg
func CreateAccount(account *types.Account) error {
	_, err := Bun.NewInsert().
		Model(account).
		Exec(context.Background())
	return err
}

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/chocological13/kittykeeper/cat-service/internal/database/repository"
	"github.com/chocological13/kittykeeper/cat-service/internal/models"
	"github.com/chocological13/kittykeeper/cat-service/internal/utils"
	"github.com/google/uuid"
)

var (
	ErrCatNotFound    = errors.New("cat not found")
	ErrNotCatOwner    = errors.New("user is not the owner of the cat")
	ErrInvalidCatData = errors.New("invalid cat data")
)

type CatService struct {
	db repository.Querier
}

func NewCatService(db repository.Querier) *CatService {
	return &CatService{db: db}
}

func (s *CatService) VerifyCatOwnership(ctx context.Context, catID, userID uuid.UUID) (bool, error) {
	// * Get the owner of the cat
	ownerID, err := s.db.GetCatOwner(ctx, catID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, ErrCatNotFound
		}
		return false, fmt.Errorf("failed to get cat owner: %w", err)
	}

	// * Compare with ID in context
	return ownerID == userID, nil
}

func (s *CatService) CreateCat(ctx context.Context, ownerID uuid.UUID, req models.CatRequestParams) (models.CatResponse,
	error) {
	pgWeight, err := utils.PtrToPgNumeric(req.Weight)
	if err != nil {
		return models.CatResponse{}, ErrInvalidCatData
	}

	cat, err := s.db.CreateCat(ctx, repository.CreateCatParams{
		OwnerID:             ownerID,
		Name:                req.Name,
		Breed:               req.Breed,
		DateOfBirth:         utils.PtrToPgDate(req.DateOfBirth),
		Weight:              pgWeight,
		Color:               req.Color,
		Gender:              req.Gender,
		PhotoUrl:            req.PhotoUrl,
		MedicalNotes:        req.MedicalNotes,
		DietaryRequirements: req.DietaryRequirements,
	})
	if err != nil {
		return models.CatResponse{}, err
	}

	return utils.FromDBCat(cat), nil
}

func (s *CatService) GetCatByID(ctx context.Context, catID, userID uuid.UUID) (models.CatResponse, error) {
	cat, err := s.db.GetCatByID(ctx, catID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.CatResponse{}, ErrCatNotFound
		}
		return models.CatResponse{}, fmt.Errorf("failed to get cat: %w", err)
	}

	if cat.OwnerID != userID {
		return models.CatResponse{}, ErrNotCatOwner
	}

	return utils.FromDBCat(cat), nil
}

func (s *CatService) ListCatsByOwner(ctx context.Context, userID uuid.UUID) ([]models.CatResponse, error) {
	cats, err := s.db.ListCatsByOwner(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list cats by owner: %w", err)
	}

	responses := make([]models.CatResponse, len(cats))
	for i, cat := range cats {
		responses[i] = utils.FromDBCat(cat)
	}

	return responses, nil
}

func (s *CatService) UpdateCat(ctx context.Context, catID, userID uuid.UUID, req models.CatRequestParams) (models.CatResponse,
	error) {
	pgWeight, err := utils.PtrToPgNumeric(req.Weight)
	if err != nil {
		return models.CatResponse{}, ErrInvalidCatData
	}

	isOwner, err := s.VerifyCatOwnership(ctx, catID, userID)
	if err != nil {
		return models.CatResponse{}, err
	}
	if !isOwner {
		return models.CatResponse{}, ErrNotCatOwner
	}

	params := repository.UpdateCatParams{
		Name:                req.Name,
		Breed:               req.Breed,
		DateOfBirth:         utils.PtrToPgDate(req.DateOfBirth),
		Weight:              pgWeight,
		Color:               req.Color,
		Gender:              req.Gender,
		PhotoUrl:            req.PhotoUrl,
		MedicalNotes:        req.MedicalNotes,
		DietaryRequirements: req.DietaryRequirements,
		ID:                  catID,
		OwnerID:             userID,
	}

	newCat, err := s.db.UpdateCat(ctx, params)
	if err != nil {
		return models.CatResponse{}, err
	}

	return utils.FromDBCat(newCat), nil
}

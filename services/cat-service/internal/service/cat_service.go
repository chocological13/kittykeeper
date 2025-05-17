package service

import (
	"context"
	"errors"
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

func (s *CatService) CreateCat(ctx context.Context, ownerID uuid.UUID, req models.CreateCatRequestParams) (models.CatResponse,
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

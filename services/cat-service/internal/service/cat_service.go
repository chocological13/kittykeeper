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

func (s *CatService) CatByOwnerExists(ctx context.Context, catID, userID uuid.UUID) error {
	exists, err := s.db.CatByOwnerExists(ctx, repository.CatByOwnerExistsParams{
		ID:      catID,
		OwnerID: userID,
	})
	if err != nil {
		return fmt.Errorf("failed to check if cat exists: %w", err)
	}
	if !exists {
		return ErrCatNotFound
	}
	return nil
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
		return models.CatResponse{}, fmt.Errorf("%w: failed to parse weight: %w", ErrInvalidCatData, err)
	}

	if req.DateOfDeath != nil {
		birthday, err := s.db.GetCatsBirthday(ctx, repository.GetCatsBirthdayParams{
			ID:      catID,
			OwnerID: userID,
		})
		if err != nil {
			return models.CatResponse{}, fmt.Errorf("failed to get cats death: %w", err)
		} else if birthday.Time.After(*req.DateOfDeath) {
			return models.CatResponse{}, fmt.Errorf("%w: date of death cannot be before date of birth", ErrInvalidCatData)
		}
	}

	var name string
	if req.Name == "" {
		initialName, err := s.db.GetCatsName(ctx, repository.GetCatsNameParams{
			ID:      catID,
			OwnerID: userID,
		})
		if err != nil {
			return models.CatResponse{}, fmt.Errorf("failed to get cats name: %w", err)
		}
		name = initialName
	} else {
		name = req.Name
	}

	params := repository.UpdateCatParams{
		Name:                name,
		Breed:               req.Breed,
		DateOfBirth:         utils.PtrToPgDate(req.DateOfBirth),
		Weight:              pgWeight,
		Color:               req.Color,
		Gender:              req.Gender,
		PhotoUrl:            req.PhotoUrl,
		MedicalNotes:        req.MedicalNotes,
		DietaryRequirements: req.DietaryRequirements,
		DateOfDeath:         utils.PtrToPgDate(req.DateOfDeath),
		ID:                  catID,
		OwnerID:             userID,
	}

	newCat, err := s.db.UpdateCat(ctx, params)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			/*
				If there's no result,
				it means either the cat doesn't exist,
				or the user is not the owner
			*/
			err = s.CatByOwnerExists(ctx, catID, userID)
			if err != nil {
				return models.CatResponse{}, err // --> either ErrCatNotFound or error
			}
			return models.CatResponse{}, ErrNotCatOwner
		}
		return models.CatResponse{}, err
	}

	return utils.FromDBCat(newCat), nil
}

func (s *CatService) ClearDateOfDeath(ctx context.Context, catID, userID uuid.UUID) error {
	result, err := s.db.ClearDateOfDeath(ctx, repository.ClearDateOfDeathParams{
		ID:      catID,
		OwnerID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = s.CatByOwnerExists(ctx, catID, userID)
			if err != nil {
				return err
			}
			return ErrNotCatOwner
		}
		return fmt.Errorf("failed to clear date of death: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete cat: no rows affected")
	}

	return nil
}

func (s *CatService) DeleteCat(ctx context.Context, catID, userID uuid.UUID) error {
	result, err := s.db.SoftDeleteCat(ctx, repository.SoftDeleteCatParams{
		ID:      catID,
		OwnerID: userID,
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = s.CatByOwnerExists(ctx, catID, userID)
			if err != nil {
				return err
			}
			return ErrNotCatOwner
		}
		return fmt.Errorf("failed to delete cat: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("failed to delete cat: no rows affected")
	}

	return nil
}

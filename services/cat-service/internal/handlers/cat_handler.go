package handlers

import (
	"errors"
	"github.com/chocological13/kittykeeper/cat-service/internal/models"
	"github.com/chocological13/kittykeeper/cat-service/internal/service"
	"github.com/chocological13/kittykeeper/cat-service/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"net/http"
	"time"
)

type CatHandler struct {
	catService *service.CatService
	log        *log.Entry
}

func NewCatHandler(catService *service.CatService, log *log.Entry) *CatHandler {
	log = log.WithField("handler", "cat-handler")
	return &CatHandler{
		catService: catService,
		log:        log,
	}
}

func (h *CatHandler) CreateCat(c *gin.Context) {
	var req models.CreateCatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("failed to bind request to submit cat data")
		errs := utils.FormatValidationError(err)
		c.JSON(400, gin.H{"error": errs})
		return
	}

	var dob *time.Time
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		d, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			h.log.WithError(err).Warn("failed to parse date of birth")
			c.JSON(400, gin.H{"error": "invalid date of birth"})
			return
		}
		dob = &d
	}

	catParams := models.CreateCatRequestParams{
		Name:                req.Name,
		Breed:               req.Breed,
		DateOfBirth:         dob,
		Weight:              req.Weight,
		Color:               req.Color,
		Gender:              req.Gender,
		PhotoUrl:            req.PhotoUrl,
		MedicalNotes:        req.MedicalNotes,
		DietaryRequirements: req.DietaryRequirements,
	}

	userIDValue, exists := c.Get("userID")
	if !exists {
		h.log.Warnf("user not authenticated")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		h.log.Error("invalid user id")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	h.log.Info("Creating cat")
	cat, err := h.catService.CreateCat(c.Request.Context(), userID, catParams)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrInvalidCatData) {
			h.log.WithError(err).Warn("failed to create cat")
			statusCode = http.StatusBadRequest
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusCreated, gin.H{"cat": cat})
}

func (h *CatHandler) GetCat(c *gin.Context) {
	// * Get cat ID from param
	catIDStr := c.Query("id")
	if catIDStr == "" {
		h.log.Warn("cat id not provided")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cat id not provided"})
		return
	}

	catID, err := uuid.Parse(catIDStr)
	if err != nil {
		h.log.WithError(err).Warnf("failed to parse cat id: %s", catIDStr)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid cat id"})
		return
	}

	userIDStr, exists := c.Get("userID")
	if !exists {
		h.log.Warnf("user not authenticated")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}
	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		h.log.Error("invalid user id")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return
	}

	cat, err := h.catService.GetCatByID(c.Request.Context(), catID, userID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, service.ErrCatNotFound) {
			h.log.WithError(err).Warnf("failed to get cat with id: %s", catIDStr)
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, gin.H{"cat": cat})
}

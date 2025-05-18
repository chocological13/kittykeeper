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

	params, err := req.ToParams()
	if err != nil {
		h.log.WithError(err).Warn("failed to convert request to cat params")
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	h.log.Info("Creating cat")
	cat, err := h.catService.CreateCat(c.Request.Context(), userID, params)
	if err != nil {
		h.errorHandler(c, err)
		return
	}

	c.JSON(http.StatusCreated, gin.H{"cat": cat})
}

func (h *CatHandler) GetCat(c *gin.Context) {
	// * Get cat ID from param
	catID, ok := h.getCatID(c)
	if !ok {
		return
	}

	// * Get user ID from context
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	cat, err := h.catService.GetCatByID(c.Request.Context(), catID, userID)
	if err != nil {
		h.errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"cat": cat})
}

func (h *CatHandler) ListCatsByOwner(c *gin.Context) {
	// * Get user ID from context
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	cats, err := h.catService.ListCatsByOwner(c.Request.Context(), userID)
	if err != nil {
		h.errorHandler(c, err)
		return
	} else if len(cats) == 0 {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "no cats found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cats": cats})
}

func (h *CatHandler) UpdateCat(c *gin.Context) {
	catID, ok := h.getCatID(c)
	if !ok {
		return
	}

	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	var req models.UpdateCatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.log.WithError(err).Warn("failed to bind request to update cat data")
		errs := utils.FormatValidationError(err)
		c.JSON(400, gin.H{"error": errs})
	}
	params, err := req.ToParams()
	if err != nil {
		h.log.WithError(err).Warn("failed to convert request to cat params")
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	h.log.Info("Updating cat")
	updatedCat, err := h.catService.UpdateCat(c.Request.Context(), catID, userID, params)
	if err != nil {
		h.errorHandler(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"cat": updatedCat})
}

func (h *CatHandler) ClearDateOfDeath(c *gin.Context) {
	catID, ok := h.getCatID(c)
	if !ok {
		return
	}
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	h.log.Info("Clearing date of death")
	err := h.catService.ClearDateOfDeath(c.Request.Context(), catID, userID)
	if err != nil {
		h.errorHandler(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *CatHandler) DeleteCat(c *gin.Context) {
	catID, ok := h.getCatID(c)
	if !ok {
		return
	}
	userID, ok := h.getUserID(c)
	if !ok {
		return
	}

	h.log.Info("Deleting cat")
	err := h.catService.DeleteCat(c.Request.Context(), catID, userID)
	if err != nil {
		h.errorHandler(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ? Helper
func (h *CatHandler) getCatID(c *gin.Context) (uuid.UUID, bool) {
	catIDStr := c.Param("id")
	if catIDStr == "" {
		h.log.Warn("cat id not provided")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "cat id not provided"})
		return uuid.UUID{}, false
	}

	catID, err := uuid.Parse(catIDStr)
	if err != nil {
		h.log.WithError(err).Warnf("failed to parse cat id: %s", catIDStr)
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid cat id"})
		return uuid.UUID{}, false
	}

	return catID, true
}

func (h *CatHandler) getUserID(c *gin.Context) (uuid.UUID, bool) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		h.log.Warnf("user not authenticated")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return uuid.UUID{}, false
	}
	userID, ok := userIDStr.(uuid.UUID)
	if !ok {
		h.log.Error("invalid user id")
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "invalid user id"})
		return uuid.UUID{}, false
	}

	return userID, true
}

func (h *CatHandler) errorHandler(c *gin.Context, err error) {
	statusCode := http.StatusInternalServerError
	if errors.Is(err, service.ErrCatNotFound) {
		h.log.WithError(err).Warn("cat not found")
		statusCode = http.StatusNotFound
	} else if errors.Is(err, service.ErrNotCatOwner) {
		h.log.WithError(err).Warn("user is not the owner of the cat")
		statusCode = http.StatusForbidden
	} else if errors.Is(err, service.ErrInvalidCatData) {
		h.log.WithError(err).Warn("invalid cat data")
	}
	h.log.WithError(err).Error("failed to handle request")
	c.JSON(statusCode, gin.H{"error": err.Error()})
}

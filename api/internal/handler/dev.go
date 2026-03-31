package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/model"
	"gorm.io/gorm"
)

type DevHandler struct {
	db *gorm.DB
}

func NewDevHandler(db *gorm.DB) *DevHandler {
	return &DevHandler{db: db}
}

func (h *DevHandler) ResetDB(c *gin.Context) {
	tables := []interface{}{
		&model.Message{},
		&model.Room{},
		&model.Profile{},
		&model.User{},
	}

	for _, table := range tables {
		if err := h.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(table).Error; err != nil {
			dto.Fail(c, http.StatusInternalServerError, "Failed to reset: "+err.Error())
			return
		}
	}

	dto.OKWithMessage(c, "Database cleared", nil)
}

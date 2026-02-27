package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

type UserHandler struct {
	userService *service.UserService
	roomService *service.RoomService
	roomRepo    repository.RoomRepository
	messageRepo repository.MessageRepository
	config      *config.Config
}

func NewUserHandler(
	userService *service.UserService,
	roomService *service.RoomService,
	roomRepo repository.RoomRepository,
	messageRepo repository.MessageRepository,
	cfg *config.Config,
) *UserHandler {
	return &UserHandler{
		userService: userService,
		roomService: roomService,
		roomRepo:    roomRepo,
		messageRepo: messageRepo,
		config:      cfg,
	}
}

func (h *UserHandler) GetUserState(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		dto.Fail(c, http.StatusInternalServerError, "Failed to get user")
		return
	}

	profile, err := h.userService.GetProfile(c.Request.Context(), userID)
	isNewUser := err != nil || profile == nil

	userDTO := dto.UserDTO{
		ID:        user.ID.String(),
		Email:     user.Email,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
	}
	if profile != nil {
		userDTO.Profile = &dto.ProfileDTO{
			Age:      profile.Age,
			City:     profile.City,
			IsMale:   profile.IsMale,
			IsHidden: profile.IsHidden,
		}
	}

	resp := dto.UserStateResponse{
		User:      userDTO,
		IsNewUser: isNewUser,
	}

	room, err := h.roomRepo.FindActiveByUserID(c.Request.Context(), userID)
	if err == nil && room != nil {
		resp.Room = &dto.RoomDTO{
			ID:       room.ID.String(),
			User1ID:  room.User1ID.String(),
			User2ID:  room.User2ID.String(),
			Category: room.Category,
		}

		messages, err := h.messageRepo.FindByRoomID(c.Request.Context(), room.ID)
		if err == nil {
			resp.Messages = make([]dto.MessageDTO, len(messages))
			for i, msg := range messages {
				resp.Messages[i] = dto.MessageDTO{
					ID:        msg.ID.String(),
					RoomID:    msg.RoomID.String(),
					SenderID:  msg.SenderID.String(),
					Content:   msg.Content,
					CreatedAt: msg.CreatedAt.Unix(),
				}
			}
		}
	}

	dto.OK(c, resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		dto.Fail(c, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.Fail(c, http.StatusBadRequest, "Invalid request body")
		return
	}

	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID, req.IsMale, req.Age, req.City, req.IsHidden)
	if err != nil {
		dto.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	dto.OKWithMessage(c, "Profile updated successfully", dto.ProfileDTO{
		Age:      profile.Age,
		City:     profile.City,
		IsMale:   profile.IsMale,
		IsHidden: profile.IsHidden,
	})
}

func (h *UserHandler) LeaveCurrentRoom(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	err := h.roomService.LeaveCurrentRoom(c.Request.Context(), userID)
	if err != nil {
		dto.Fail(c, http.StatusBadRequest, err.Error())
		return
	}

	dto.OKWithMessage(c, "Successfully left room", nil)
}

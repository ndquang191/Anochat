package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

type UserHandler struct {
	userService  *service.UserService
	roomService  *service.RoomService
	queueService *service.QueueService
	roomRepo     repository.RoomRepository
	messageRepo  repository.MessageRepository
	config       *config.Config
}

func NewUserHandler(
	userService *service.UserService,
	roomService *service.RoomService,
	queueService *service.QueueService,
	roomRepo repository.RoomRepository,
	messageRepo repository.MessageRepository,
	cfg *config.Config,
) *UserHandler {
	return &UserHandler{
		userService:  userService,
		roomService:  roomService,
		queueService: queueService,
		roomRepo:     roomRepo,
		messageRepo:  messageRepo,
		config:       cfg,
	}
}

func (h *UserHandler) GetUserState(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	profile, _ := h.userService.GetProfile(c.Request.Context(), userID)

	resp := dto.UserStateResponse{
		InQueue: h.queueService.IsInQueue(userID),
	}
	if profile != nil {
		resp.Profile = &dto.ProfileDTO{
			Nickname: profile.Nickname,
			Age:      profile.Age,
			IsMale:   profile.IsMale,
			IsHidden: profile.IsHidden,
		}
	}

	room, err := h.roomRepo.FindActiveByUserID(c.Request.Context(), userID)
	if err == nil && room != nil {
		roomDTO := &dto.RoomDTO{
			ID:      room.ID.String(),
			User1ID: room.User1ID.String(),
			User2ID: room.User2ID.String(),
		}

		// Determine partner ID and fetch their profile
		partnerID := room.User1ID
		if room.User1ID == userID {
			partnerID = room.User2ID
		}
		partnerUser, err := h.userService.GetUserByID(c.Request.Context(), partnerID)
		if err == nil && partnerUser != nil {
			partnerDTO := &dto.UserDTO{
				ID: partnerUser.ID.String(),
			}
			partnerProfile, err := h.userService.GetProfile(c.Request.Context(), partnerID)
			if err == nil && partnerProfile != nil {
				if !partnerProfile.IsHidden {
					partnerDTO.Name = partnerUser.Name
					partnerDTO.Nickname = partnerProfile.Nickname
					partnerDTO.Profile = &dto.ProfileDTO{
						Age:      partnerProfile.Age,
						IsMale:   partnerProfile.IsMale,
						IsHidden: false,
					}
				} else {
					partnerDTO.Profile = &dto.ProfileDTO{IsHidden: true}
				}
			} else {
				partnerDTO.Name = partnerUser.Name
			}
			roomDTO.Partner = partnerDTO
		}

		resp.Room = roomDTO

		messages, err := h.messageRepo.FindByRoomID(c.Request.Context(), room.ID)
		if err == nil {
			resp.Messages = make([]dto.MessageDTO, len(messages))
			for i, msg := range messages {
				resp.Messages[i] = dto.MessageDTO{
					ID:        msg.ID.String(),
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

	if req.Age != nil && (*req.Age < h.config.User.MinAge || *req.Age > h.config.User.MaxAge) {
		dto.Fail(c, http.StatusBadRequest, fmt.Sprintf("Age must be between %d and %d", h.config.User.MinAge, h.config.User.MaxAge))
		return
	}

	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID, req.Nickname, req.IsMale, req.Age, req.IsHidden)
	if err != nil {
		dto.Fail(c, http.StatusInternalServerError, err.Error())
		return
	}

	dto.OKWithMessage(c, "Profile updated successfully", dto.ProfileDTO{
		Nickname: profile.Nickname,
		Age:      profile.Age,
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

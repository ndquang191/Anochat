package handler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

type UserHandler struct {
	userService  *service.UserService
	roomService  *service.RoomService
	queueService *service.QueueService
	fakeService  *service.FakeMatchService
	roomRepo     repository.RoomRepository
	messageRepo  repository.MessageRepository
	config       *config.Config
}

func NewUserHandler(
	userService *service.UserService,
	roomService *service.RoomService,
	queueService *service.QueueService,
	fakeService *service.FakeMatchService,
	roomRepo repository.RoomRepository,
	messageRepo repository.MessageRepository,
	cfg *config.Config,
) *UserHandler {
	return &UserHandler{
		userService:  userService,
		roomService:  roomService,
		queueService: queueService,
		fakeService:  fakeService,
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

		partnerID := room.User1ID
		if room.User1ID == userID {
			partnerID = room.User2ID
		}
		partnerUser, err := h.userService.GetUserWithProfile(c.Request.Context(), partnerID)
		if err == nil && partnerUser != nil {
			partnerDTO := &dto.UserDTO{
				ID: partnerUser.ID.String(),
			}
			if partnerUser.Profile != nil {
				if !partnerUser.Profile.IsHidden {
					partnerDTO.Name = partnerUser.Name
					partnerDTO.Nickname = partnerUser.Profile.Nickname
					partnerDTO.Profile = &dto.ProfileDTO{
						Age:      partnerUser.Profile.Age,
						IsMale:   partnerUser.Profile.IsMale,
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
	if resp.Room == nil {
		session := h.fakeService.GetByUserID(userID)
		if session != nil {
			resp.Room = &dto.RoomDTO{
				ID:      session.RoomID.String(),
				User1ID: session.UserID.String(),
				User2ID: session.PartnerID.String(),
				Partner: &dto.UserDTO{
					ID:   session.PartnerID.String(),
					Name: &session.PartnerName,
					Profile: &dto.ProfileDTO{
						Age:      &session.PartnerAge,
						IsMale:   &session.PartnerIsMale,
						IsHidden: false,
					},
				},
			}
			resp.Messages = []dto.MessageDTO{
				{
					ID:        session.GreetingID.String(),
					SenderID:  session.PartnerID.String(),
					Content:   session.Greeting,
					CreatedAt: session.CreatedAt.Unix(),
				},
			}
		}
	}

	dto.OK(c, resp)
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		dto.FailErr(c, apperr.ErrUnauthenticated)
		return
	}

	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.FailErr(c, apperr.ErrInvalidBody)
		return
	}

	if req.Age != nil && (*req.Age < h.config.User.MinAge || *req.Age > h.config.User.MaxAge) {
		dto.Fail(c, http.StatusBadRequest, fmt.Sprintf("Tuoi phai nam trong khoang tu %d den %d", h.config.User.MinAge, h.config.User.MaxAge))
		return
	}

	profile, err := h.userService.UpdateProfile(c.Request.Context(), userID, req.Nickname, req.IsMale, req.Age, req.IsHidden)
	if err != nil {
		dto.FailErr(c, err)
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
		if h.fakeService.EndSession(userID) != nil {
			dto.OKWithMessage(c, "Successfully left room", nil)
			return
		}
		dto.FailErr(c, err)
		return
	}

	dto.OKWithMessage(c, "Successfully left room", nil)
}

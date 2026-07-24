package handler

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/chat"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/dto"
	"github.com/ndquang191/Anochat/api/internal/repository"
	"github.com/ndquang191/Anochat/api/internal/service"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/ndquang191/Anochat/api/pkg/config"
)

type UserHandler struct {
	userService    *service.UserService
	roomService    *service.RoomService
	messageService *service.MessageService
	queueService   *service.QueueService
	roomRepo       repository.RoomRepository
	config         *config.Config
}

const messageCursorSize = 24

func nicknameChangeAvailableAt(profile *identity.Profile) *int64 {
	next := service.DisplayNameChangeAvailableAt(profile, time.Now())
	if next == nil {
		return nil
	}
	unix := next.Unix()
	return &unix
}

func NewUserHandler(
	userService *service.UserService,
	roomService *service.RoomService,
	messageService *service.MessageService,
	queueService *service.QueueService,
	roomRepo repository.RoomRepository,
	cfg *config.Config,
) *UserHandler {
	return &UserHandler{
		userService:    userService,
		roomService:    roomService,
		messageService: messageService,
		queueService:   queueService,
		roomRepo:       roomRepo,
		config:         cfg,
	}
}

func (h *UserHandler) GetUserState(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		signOutAndRedirect(c, h.config)
		return
	}

	resp := dto.UserStateResponse{
		IsAdmin:            c.GetBool("is_admin"),
		IsBanned:           !c.GetBool("is_active"),
		BanCount:           c.GetInt("ban_count"),
		ReviewRequestCount: c.GetInt("review_request_count"),
		ReviewRequested:    c.GetBool("review_requested"),
	}
	currentUser, err := h.userService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		dto.FailErr(c, err)
		return
	}
	resp.User = &dto.UserDTO{
		ID:        currentUser.ID.String(),
		Email:     currentUser.Email,
		Name:      currentUser.Name,
		AvatarURL: currentUser.AvatarURL,
		IsAdmin:   currentUser.IsAdmin,
	}
	if resp.IsBanned {
		dto.OK(c, resp)
		return
	}

	resp.InQueue = h.queueService.IsInQueue(userID)
	profile, _ := h.userService.GetProfile(c.Request.Context(), userID)
	if profile != nil {
		resp.Profile = &dto.ProfileDTO{
			Nickname:                  profile.Nickname,
			NicknameChangeAvailableAt: nicknameChangeAvailableAt(profile),
			Age:                       profile.Age,
			IsMale:                    profile.IsMale,
			IsHidden:                  profile.IsHidden,
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

		page, err := h.messageService.GetMessagePage(
			c.Request.Context(),
			room.ID,
			nil,
			service.DefaultMessagePageSize,
		)
		if err == nil {
			resp.Messages = messageDTOs(page.Messages)
			resp.MessagesHasMore = page.HasMore
			resp.MessagesNextCursor = encodeMessageCursor(page.NextCursor)
		}
	}
	dto.OK(c, resp)
}

func (h *UserHandler) GetRoomMessages(c *gin.Context) {
	userID := getUserID(c)
	if userID == uuid.Nil {
		dto.FailErr(c, apperr.ErrUnauthenticated)
		return
	}

	roomID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		dto.FailErr(c, apperr.ErrInvalidID)
		return
	}
	room, err := h.roomService.GetRoomByID(c.Request.Context(), roomID)
	if err != nil || room == nil || !room.HasUser(userID) {
		dto.FailErr(c, apperr.ErrNotRoomMember)
		return
	}

	limit, err := parseMessagePageSize(c.Query("limit"))
	if err != nil {
		dto.FailErr(c, err)
		return
	}
	before, err := decodeMessageCursor(c.Query("before"))
	if err != nil {
		dto.FailErr(c, err)
		return
	}

	page, err := h.messageService.GetMessagePage(
		c.Request.Context(),
		roomID,
		before,
		limit,
	)
	if err != nil {
		dto.FailErr(c, err)
		return
	}

	dto.OK(c, dto.MessagePageResponse{
		Messages:   messageDTOs(page.Messages),
		NextCursor: encodeMessageCursor(page.NextCursor),
		HasMore:    page.HasMore,
	})
}

func parseMessagePageSize(raw string) (int, error) {
	if raw == "" {
		return service.DefaultMessagePageSize, nil
	}
	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 || limit > service.MaxMessagePageSize {
		return 0, apperr.ErrInvalidPageSize
	}
	return limit, nil
}

func encodeMessageCursor(cursor *chat.MessageCursor) *string {
	if cursor == nil {
		return nil
	}
	data := make([]byte, messageCursorSize)
	binary.BigEndian.PutUint64(data[:8], uint64(cursor.CreatedAt.UnixNano()))
	copy(data[8:], cursor.ID[:])
	encoded := base64.RawURLEncoding.EncodeToString(data)
	return &encoded
}

func decodeMessageCursor(raw string) (*chat.MessageCursor, error) {
	if raw == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(data) != messageCursorSize {
		return nil, apperr.ErrInvalidCursor
	}
	id, err := uuid.FromBytes(data[8:])
	if err != nil || id == uuid.Nil {
		return nil, apperr.ErrInvalidCursor
	}
	return &chat.MessageCursor{
		ID:        id,
		CreatedAt: time.Unix(0, int64(binary.BigEndian.Uint64(data[:8]))).UTC(),
	}, nil
}

func messageDTOs(messages []*chat.Message) []dto.MessageDTO {
	result := make([]dto.MessageDTO, len(messages))
	for i, msg := range messages {
		result[i] = dto.MessageDTO{
			ID:        msg.ID.String(),
			SenderID:  msg.SenderID.String(),
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Unix(),
		}
	}
	return result
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
		Nickname:                  profile.Nickname,
		NicknameChangeAvailableAt: nicknameChangeAvailableAt(profile),
		Age:                       profile.Age,
		IsMale:                    profile.IsMale,
		IsHidden:                  profile.IsHidden,
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
		dto.FailErr(c, err)
		return
	}

	dto.OKWithMessage(c, "Successfully left room", nil)
}

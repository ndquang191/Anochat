package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ndquang191/Anochat/api/internal/domain/identity"
	"github.com/ndquang191/Anochat/api/internal/domain/moderation"
	"github.com/ndquang191/Anochat/api/pkg/apperr"
	"github.com/stretchr/testify/require"
)

func TestRequireAdminUsesAuthenticatedRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ModerationHandler{}

	tests := []struct {
		name        string
		isAdmin     bool
		wantAllowed bool
		wantStatus  int
	}{
		{name: "admin", isAdmin: true, wantAllowed: true, wantStatus: http.StatusOK},
		{name: "regular user", isAdmin: false, wantAllowed: false, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response := httptest.NewRecorder()
			context, _ := gin.CreateTestContext(response)
			context.Set("is_admin", tt.isAdmin)

			if allowed := handler.requireAdmin(context); allowed != tt.wantAllowed {
				t.Errorf("requireAdmin() = %v, want %v", allowed, tt.wantAllowed)
			}
			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
		})
	}
}

func TestReportGroupCursorRoundTrip(t *testing.T) {
	want := &moderation.ReportGroupCursor{
		ReportCount:    42,
		ReportedUserID: uuid.New(),
	}

	got, err := decodeReportGroupCursor(encodeReportGroupCursor(want))

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestBannedUserCursorRoundTrip(t *testing.T) {
	want := &identity.BannedUserCursor{
		ReviewRequested: true,
		SortAt:          time.Now().UTC().Truncate(time.Microsecond),
		ID:              uuid.New(),
	}

	got, err := decodeBannedUserCursor(encodeBannedUserCursor(want))

	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestAdminCursorRejectsInvalidValue(t *testing.T) {
	_, err := decodeReportGroupCursor("not-a-cursor")
	require.ErrorIs(t, err, apperr.ErrInvalidCursor)

	_, err = decodeBannedUserCursor("not-a-cursor")
	require.ErrorIs(t, err, apperr.ErrInvalidCursor)
}

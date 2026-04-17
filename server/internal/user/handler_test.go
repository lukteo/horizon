package user_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/platform/middleware"
	"github.com/luketeo/horizon/internal/platform/testhelper"
	"github.com/luketeo/horizon/internal/user"
)

func newUserHandler(t *testing.T) *user.Handler {
	t.Helper()
	db := testhelper.DB(t)
	testhelper.Reset(t, db)
	svc := user.NewService(user.NewRepo(db), slog.New(slog.NewTextHandler(io.Discard, nil)))
	return user.NewHandler(svc)
}

func TestGetUsersMe_Unauthenticated(t *testing.T) {
	h := newUserHandler(t)

	resp, err := h.GetUsersMe(context.Background(), oapi.GetUsersMeRequestObject{})
	if err != nil {
		t.Fatalf("GetUsersMe: %v", err)
	}
	if _, ok := resp.(oapi.GetUsersMe401ApplicationProblemPlusJSONResponse); !ok {
		t.Fatalf("want 401 response, got %T", resp)
	}
}

func TestGetUsersMe_ReturnsUpsertedUser(t *testing.T) {
	h := newUserHandler(t)
	clerkU := fakeClerkUser(
		"user_handler_1",
		"gina@example.com",
		strPtr("Gina"),
		strPtr("Gray"),
		nil,
	)
	ctx := middleware.WithClerkUser(context.Background(), clerkU)

	resp, err := h.GetUsersMe(ctx, oapi.GetUsersMeRequestObject{})
	if err != nil {
		t.Fatalf("GetUsersMe: %v", err)
	}
	ok200, is := resp.(oapi.GetUsersMe200JSONResponse)
	if !is {
		t.Fatalf("want 200 response, got %T", resp)
	}
	if ok200.Email != "gina@example.com" {
		t.Errorf("email: want gina@example.com, got %q", ok200.Email)
	}
}

func TestUpdateUsersMe_Unauthenticated(t *testing.T) {
	h := newUserHandler(t)

	resp, err := h.UpdateUsersMe(context.Background(), oapi.UpdateUsersMeRequestObject{
		Body: &oapi.UpdateUsersMeJSONRequestBody{FirstName: strPtr("X")},
	})
	if err != nil {
		t.Fatalf("UpdateUsersMe: %v", err)
	}
	if _, ok := resp.(oapi.UpdateUsersMe401ApplicationProblemPlusJSONResponse); !ok {
		t.Fatalf("want 401 response, got %T", resp)
	}
}

func TestUpdateUsersMe_AppliesPatch(t *testing.T) {
	h := newUserHandler(t)
	clerkU := fakeClerkUser(
		"user_handler_2",
		"henry@example.com",
		strPtr("Henry"),
		strPtr("Hart"),
		nil,
	)
	ctx := middleware.WithClerkUser(context.Background(), clerkU)

	resp, err := h.UpdateUsersMe(ctx, oapi.UpdateUsersMeRequestObject{
		Body: &oapi.UpdateUsersMeJSONRequestBody{
			FirstName: strPtr("Hank"),
			LastName:  strPtr("Hopkins"),
		},
	})
	if err != nil {
		t.Fatalf("UpdateUsersMe: %v", err)
	}
	ok200, is := resp.(oapi.UpdateUsersMe200JSONResponse)
	if !is {
		t.Fatalf("want 200 response, got %T", resp)
	}
	if ok200.FirstName == nil || *ok200.FirstName != "Hank" {
		t.Errorf("first_name: want Hank, got %v", ok200.FirstName)
	}
	if ok200.LastName == nil || *ok200.LastName != "Hopkins" {
		t.Errorf("last_name: want Hopkins, got %v", ok200.LastName)
	}
}

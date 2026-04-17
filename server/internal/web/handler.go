package web

import (
	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/config"
	"github.com/luketeo/horizon/internal/services/orgservice"
	"github.com/luketeo/horizon/internal/user"
)

// Handler is the aggregate HTTP handler. It embeds each carved-out domain handler
// so the oapi.StrictServerInterface is satisfied by method promotion. Remaining
// org/member/api-key methods are declared directly on *Handler until those
// domains are extracted in later phases.
type Handler struct {
	config  *config.Config
	orgSvc  *orgservice.Service
	userSvc *user.Service

	*user.Handler
}

// Compile-time guarantee that every oapi route has a concrete implementation.
var _ oapi.StrictServerInterface = (*Handler)(nil)

func NewHandler(cfg *config.Config) *Handler {
	db := cfg.DB()
	logger := cfg.Logger()

	userSvc := user.NewService(user.NewRepo(db), logger)

	return &Handler{
		config:  cfg,
		orgSvc:  orgservice.New(db),
		userSvc: userSvc,
		Handler: user.NewHandler(userSvc),
	}
}

package web

import (
	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/config"
	"github.com/luketeo/horizon/internal/services/orgservice"
)

type Handler struct {
	oapi.StrictServerInterface

	config *config.Config
	orgSvc *orgservice.Service
}

func NewHandler(config *config.Config) *Handler {
	return &Handler{
		config: config,
		orgSvc: orgservice.New(config.DB()),
	}
}

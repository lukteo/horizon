package web

import (
	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/config"
)

type Handler struct {
	oapi.StrictServerInterface

	config *config.Config
}

func NewHandler(config *config.Config) *Handler {
	h := Handler{
		config: config,
	}

	return &h
}

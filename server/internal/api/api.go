package api

import (
	"github.com/luketeo/horizon/config"
	"github.com/luketeo/horizon/generated/oapi"
)

type API struct {
	oapi.StrictServerInterface

	config *config.Config
}

func NewAPI(config *config.Config) *API {
	api := API{
		config: config,
	}

	return &api
}

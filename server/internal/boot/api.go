package boot

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/luketeo/horizon/config"
	"github.com/luketeo/horizon/generated/oapi"
	"github.com/luketeo/horizon/internal/api"
	"github.com/luketeo/horizon/internal/middleware"
)

type API struct {
	router   *chi.Mux
	portAddr string
}

func NewAPI(config *config.Config) *API {
	api := api.NewAPI(config)
	portAddr := fmt.Sprintf(":%s", config.Env().ServerPort())

	r := chi.NewRouter()
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Get("/health", api.GetHealth)
	r.Group(func(r chi.Router) {
		baseURL := ""

		// clerk auth middleware
		r.Use(middleware.NewClerkAuthMiddleware(config))

		serverOptions := oapi.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w http.ResponseWriter, _ *http.Request, err error) {
				http.Error(w, err.Error(), http.StatusBadRequest)
			},
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				errMsg := "Internal server error occurred"
				slog.Default().ErrorContext(r.Context(), errMsg, slog.Any("err", err))

				http.Error(w, err.Error(), http.StatusInternalServerError)
			},
		}
		strictHandler := oapi.NewStrictHandlerWithOptions(
			api,
			[]oapi.StrictMiddlewareFunc{},
			serverOptions,
		)

		oapi.HandlerFromMuxWithBaseURL(strictHandler, r, baseURL)
	})

	return &API{
		router:   r,
		portAddr: portAddr,
	}
}

func (ab *API) Start() {
	slog.Default().Info("Server listening on " + ab.portAddr)

	headerTimeout := 1000
	httpServer := &http.Server{
		Handler:           ab.router,
		Addr:              ab.portAddr,
		ReadHeaderTimeout: time.Duration(headerTimeout) * time.Second,
	}

	err := httpServer.ListenAndServe()
	if err != nil {
		slog.Default().
			Error("Error starting server on "+ab.portAddr, slog.Any("err", err))
		os.Exit(1)
	}
}

func (ab *API) Router() *chi.Mux {
	return ab.router
}

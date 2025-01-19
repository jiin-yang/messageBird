package server

import (
	"fmt"
	"github.com/jiin-yang/messageBird/config"
	mw "github.com/jiin-yang/messageBird/internal/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/ory/graceful"
	"net/http"
	"strings"
)

type Server struct {
	e      *echo.Echo
	config *config.Config
}

func New(c *config.Config) *Server {
	server := &Server{}

	e := echo.New()
	e.HideBanner = true

	e.Validator = NewValidator()
	e.Use(middleware.Recover())
	e.Use(mw.CommonHeaderSetterMiddleware)
	e.Use(middleware.CORS())
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20)))
	e.Use(middleware.GzipWithConfig(middleware.GzipConfig{
		Level: 5,
		Skipper: func(c echo.Context) bool {
			if strings.Contains(c.Request().URL.Path, "swagger") {
				return true
			}
			return false
		},
	}))

	server.e = e
	server.config = c

	return server
}

func (s *Server) Start() error {
	s.e.Server.Addr = fmt.Sprintf(":%d", s.config.ServerConfig.Port)

	log.Info("Server Start Successfully!")
	s.e.GET("/health", s.healthCheck)
	return graceful.Graceful(s.e.Server.ListenAndServe, s.e.Server.Shutdown)
}

func (s *Server) healthCheck(ctx echo.Context) error {
	log.Info("Success health check!")
	return ctx.NoContent(http.StatusNoContent)
}

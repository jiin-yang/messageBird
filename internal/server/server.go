package server

import (
	"fmt"
	"github.com/jiin-yang/messageBird/config"
	"github.com/jiin-yang/messageBird/internal/client/webhook"
	"github.com/jiin-yang/messageBird/internal/infra/repository/mongoDB"
	"github.com/jiin-yang/messageBird/internal/message"
	mw "github.com/jiin-yang/messageBird/internal/middleware"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/ory/graceful"
	"github.com/rs/zerolog/log"
	"net/http"
	"strings"
)

type Server struct {
	echo   *echo.Echo
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

	e.HTTPErrorHandler = customErrorHandler

	server.echo = e
	server.config = c

	return server
}

func (server *Server) Start() error {
	server.echo.Server.Addr = fmt.Sprintf(":%d", server.config.ServerConfig.Port)

	mongoClient, err := mongoDB.NewClient(&server.config.MongoDBConfig)
	if err != nil {
		log.Fatal().Err(err)
	}

	webhookClient := webhook.NewWebhookClient(&webhook.NewClientOptions{
		URL: server.config.WebhookConfig.URL,
	})

	messageRepository := mongoDB.NewMessageRepository(&mongoDB.NewMessageRepositoryOpts{
		Client: mongoClient,
	})

	messageUseCase := message.NewUseCase(&message.NewUseCaseOptions{
		Repo:    messageRepository,
		Webhook: webhookClient,
	})

	cronJob := message.NewCron(messageUseCase)
	defer cronJob.StopCron()

	message.NewHandler(server.echo, messageUseCase, cronJob)

	log.Info().Msg("Server Start Successfully!")

	server.echo.GET("/health", server.healthCheck)

	return graceful.Graceful(server.echo.Server.ListenAndServe, server.echo.Server.Shutdown)
}

func (server *Server) healthCheck(ctx echo.Context) error {
	log.Info().Msg("Success health check!")
	return ctx.NoContent(http.StatusNoContent)
}

package message

import (
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
	"net/http"
)

type Handler interface {
	createMessage(ctx echo.Context) error
}

type handler struct {
	echo    *echo.Echo
	useCase UseCase
	cron    *Cron
}

func NewHandler(e *echo.Echo, u UseCase, cron *Cron) Handler {
	h := &handler{
		echo:    e,
		useCase: u,
		cron:    cron,
	}
	h.registerRoutes()
	return h
}

func (h *handler) registerRoutes() {
	h.echo.POST("/messages", h.createMessage)
	h.echo.POST("/messages/cron/start", h.startCron)
	h.echo.POST("/messages/cron/stop", h.stopCron)
	h.echo.GET("/messages", h.getSentMessages)
}

func (h *handler) createMessage(ctx echo.Context) error {
	var requestDto *CreateMessageRequest
	err := ctx.Bind(&requestDto)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request body").
			SetInternal(err)
	}

	err = ctx.Validate(requestDto)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "validation error").
			SetInternal(err)
	}

	msgResponse, err := h.useCase.CreateMessage(ctx.Request().Context(), *requestDto)
	if err != nil {
		log.Error().
			Err(err).
			Str("method", "CreateMessage").
			Str("phoneNumber", requestDto.PhoneNumber).
			Msg("failed to create message - handler")

		return echo.NewHTTPError(http.StatusInternalServerError, err.Error()).
			SetInternal(err)
	}

	return ctx.JSON(http.StatusCreated, msgResponse)
}

func (h *handler) startCron(ctx echo.Context) error {
	if h.cron.IsRunning {
		log.Warn().Msg("Cron job is already running - handler")
		return ctx.JSON(http.StatusConflict, map[string]string{
			"message": "Cron job is already running - handler",
		})
	}

	h.cron.StartCron()
	log.Info().Msg("Cron job started - handler")
	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Cron job started successfully - handler",
	})
}

func (h *handler) stopCron(ctx echo.Context) error {
	if !h.cron.IsRunning {
		log.Warn().Msg("Cron job is not running - handler")
		return ctx.JSON(http.StatusConflict, map[string]string{
			"message": "Cron job is not running - handler",
		})
	}

	h.cron.StopCron()
	log.Info().Msg("Cron job stopped - handler")
	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Cron job stopped successfully - handler",
	})
}

func (h *handler) getSentMessages(ctx echo.Context) error {
	return nil
}

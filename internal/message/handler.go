package message

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler interface {
	createMessage(ctx echo.Context) error
}

type handler struct {
	echo    *echo.Echo
	useCase UseCase
}

func NewHandler(e *echo.Echo, u UseCase) Handler {
	h := &handler{
		echo:    e,
		useCase: u,
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
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create message").
			SetInternal(err)
	}

	return ctx.JSON(http.StatusCreated, msgResponse)
}

func (h *handler) startCron(ctx echo.Context) error {
	return nil
}

func (h *handler) stopCron(ctx echo.Context) error {
	return nil
}

func (h *handler) getSentMessages(ctx echo.Context) error {
	return nil
}

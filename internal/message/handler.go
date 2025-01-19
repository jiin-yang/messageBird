package message

import "github.com/labstack/echo/v4"

type Handler interface {
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

	return h
}

func (h *handler) registerRoutes() {
	h.echo.POST("/messages", h.createMessage)
	h.echo.POST("/message/cron/start", h.startCron)
	h.echo.POST("/message/cron/stop", h.stopCron)
	h.echo.GET("/messages", h.getSentMessages)
}

func (h *handler) createMessage(ctx echo.Context) error {
	return nil
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

package server

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"net/http"
)

type ErrorField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string       `json:"message"`
	Errors  []ErrorField `json:"errors,omitempty"`
}

func customErrorHandler(err error, c echo.Context) {
	var (
		statusCode = http.StatusInternalServerError
		message    = "Internal Server Error"
		resp       = ErrorResponse{}
	)

	if he, ok := err.(*echo.HTTPError); ok {
		statusCode = he.Code

		if he.Message != nil {
			message = fmt.Sprintf("%v", he.Message)
		}

		if he.Internal != nil {
			if ve, ok := he.Internal.(validator.ValidationErrors); ok {
				statusCode = http.StatusBadRequest
				message = "Validation Failed"
				resp.Errors = buildValidationErrors(ve)
			}
		}
	} else {
		if ve, ok := err.(validator.ValidationErrors); ok {
			statusCode = http.StatusBadRequest
			message = "Validation Failed"
			resp.Errors = buildValidationErrors(ve)
		}
	}

	resp.Message = message

	if !c.Response().Committed {
		c.JSON(statusCode, resp)
	}
}

func buildValidationErrors(ve validator.ValidationErrors) []ErrorField {
	out := make([]ErrorField, len(ve))

	for i, fe := range ve {
		fieldName := fe.Field()
		tag := fe.Tag()
		param := fe.Param()

		var message string
		switch tag {
		case "required":
			message = "This field is required."
		case "max":
			message = fmt.Sprintf("Length cannot be more than %s.", param)
		case "min":
			message = fmt.Sprintf("Length cannot be less than %s.", param)
		case "e164":
			message = "Invalid phone number format. (E.164 required)"
		default:
			message = fmt.Sprintf("Validation failed on the '%s' tag.", tag)
		}

		out[i] = ErrorField{
			Field:   fieldName,
			Message: message,
		}
	}
	return out
}

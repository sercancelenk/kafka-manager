package http

import (
	"errors"
	"fmt"
	stdhttp "net/http"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func ErrorHandler() fiber.ErrorHandler {
	return func(c *fiber.Ctx, err error) error {
		zap.L().Error(fmt.Sprintf("Error occured on %s", c.Request().URI().String()), zap.Error(err))

		var e *Error
		if errors.As(err, &e) {
			c.Status(e.Code)
			return c.JSON(err)
		}

		var fe *fiber.Error
		if errors.As(err, &fe) {
			c.Status(fe.Code)
			return c.JSON(New(fe.Code, fe.Message))
		}

		c.Status(stdhttp.StatusInternalServerError)
		return c.JSON(New(stdhttp.StatusInternalServerError, err.Error()))
	}
}

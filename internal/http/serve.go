package http

import (
	"errors"

	"github.com/gofiber/fiber/v2"
)

type Empty struct{}

func Serve[I any, O any](h func(*fiber.Ctx, *I) (*O, error)) fiber.Handler {
	return func(c *fiber.Ctx) error {
		r := new(I)
		if err := c.BodyParser(r); err != nil && !errors.Is(err, fiber.ErrUnprocessableEntity) {
			return err
		}

		if err := c.ParamsParser(r); err != nil {
			return err
		}

		if err := c.QueryParser(r); err != nil {
			return err
		}

		if err := c.ReqHeaderParser(r); err != nil {
			return err
		}

		res, err := h(c, r)
		if err != nil {
			return err
		}

		if res == nil {
			return nil
		}

		return c.JSON(res)
	}
}

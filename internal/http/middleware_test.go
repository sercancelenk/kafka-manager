package http

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/valyala/fasthttp"
)

func TestErrorHandler(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler: ErrorHandler(),
	})

	testCases := []struct {
		name             string
		err              error
		expectedStatus   int
		expectedResponse string
	}{
		{
			name:             "HTTP error",
			err:              &Error{Code: 400, Message: "message", Err: "Bad Request"},
			expectedStatus:   400,
			expectedResponse: `{"code":400,"error":"Bad Request","message":"message"}`,
		},
		{
			name:             "Fiber error",
			err:              &fiber.Error{Code: 500, Message: "timeout"},
			expectedStatus:   500,
			expectedResponse: `{"code":500,"error":"Internal Server Error","message":"timeout"}`,
		},
		{
			name:             "Unknown error",
			err:              errors.New("unknown error"),
			expectedStatus:   500,
			expectedResponse: `{"code":500,"error":"Internal Server Error","message":"unknown error"}`,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			c := app.AcquireCtx(&fasthttp.RequestCtx{})
			defer app.ReleaseCtx(c)

			err := app.Config().ErrorHandler(c, tc.err)

			assert.Equal(t, tc.expectedStatus, c.Response().StatusCode())
			assert.Equal(t, tc.expectedResponse, string(c.Response().Body()))
			assert.NoError(t, err)
		})
	}
}

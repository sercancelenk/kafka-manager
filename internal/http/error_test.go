package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestError_Error(t *testing.T) {
	err := Error{
		Code:    500,
		Err:     "Internal Server Error",
		Message: "unknown error",
	}

	assert.Equal(t, "unknown error", err.Error())
}

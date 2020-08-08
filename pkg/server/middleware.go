package server

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gofiber/fiber"
	jwtware "github.com/gofiber/jwt"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"
)

func errorHandler(ctx *fiber.Ctx, err error) {
	type errResponse struct {
		Error string `json:"error"`
		Code  int    `json:"code"`
	}

	logError := func(err error) {
		if err == nil {
			return
		}
		zap.L().Error("failed to send error response", zap.Error(err))
	}

	// weirdly this error is not type, so it has to be string matched
	if has := strings.HasPrefix(err.Error(), "bodyparser: cannot parse content-type:"); has {
		logError(ctx.Status(http.StatusBadRequest).JSON(errResponse{
			Error: "Content-Type: application/json header is required",
			Code:  http.StatusBadRequest,
		}))
		return
	}

	// this is one of those, we have no choice
	if has := strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint"); has {
		logError(ctx.Status(http.StatusConflict).JSON(errResponse{
			Error: "record already exists",
			Code:  http.StatusConflict,
		}))
		return
	}

	if err == gorm.ErrRecordNotFound {
		logError(ctx.Status(http.StatusNotFound).JSON(errResponse{
			Error: "record does not exist",
			Code:  http.StatusNotFound,
		}))
		return
	}

	e, ok := err.(PublicError)
	if !ok {
		zap.L().Info("masking internal error: ", zap.Error(err))
		logError(ctx.Status(500).JSON(errResponse{
			Error: "internal error",
			Code:  500,
		}))
		return
	}

	logError(ctx.Status(e.Code()).JSON(errResponse{
		Error: e.PublicError(),
		Code:  e.Code(),
	}))
}

// Custom recover middleware to get stacktrace printed on error
// Recover will recover from panics and calls the ErrorHandler
func Recover() fiber.Handler {
	return func(ctx *fiber.Ctx) {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}
				fmt.Printf("recovered from panic: %v\n%s", err, debug.Stack())
				ctx.Next(err)
				return
			}
		}()
		ctx.Next()
	}
}

func (s *Server) protected() fiber.Handler {
	if !s.config.Auth.Enabled {
		return func(ctx *fiber.Ctx) {
			ctx.Next()
		}
	}

	// Too lazy to create this middleware myself, so here we will use the
	// built-in one. This one simply checks that the token is not expired, does
	// not check if it's valid or not on our DB.
	//
	// This means that ANY non-expired tokes will work, essentially rendering
	// the user unable to logout. May fix this if we have enough time leftover
	return jwtware.New(jwtware.Config{
		ErrorHandler: jwtError,
		SigningKey:   []byte(s.config.Auth.TokenSecret),
		ContextKey:   "token",
	})
}

func jwtError(c *fiber.Ctx, err error) {
	fmt.Printf("error jwt: %+v\n\n", err)
	if err.Error() == "Missing or malformed JWT" {
		c.Next(newValidationError("missing or malformed JWT"))
		return
	}

	c.Next(newPublicError("invalid or expired JWT", 401))
	return
}

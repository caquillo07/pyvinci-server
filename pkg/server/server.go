package server

import (
	"fmt"
	"runtime/debug"
	"strings"

	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"

	"github.com/caquillo07/pyvinci-server/pkg/conf"
)

type Server struct {
	app    *fiber.App
	config *conf.Config
	db     *gorm.DB
}

type Handler func(c *fiber.Ctx) error

func NewServer(config *conf.Config, db *gorm.DB) *Server {
	srv := &Server{
		app:    fiber.New(),
		config: config,
		db:     db,
	}

	srv.applyMiddleware()
	srv.applyRoutes()
	return srv
}

func (s *Server) Serve() error {
	port := 3000
	if s.config.REST.Port != 0 {
		port = s.config.REST.Port
	}

	return s.app.Listen(port)
}

func (s *Server) applyMiddleware() {
	s.app.Use(middleware.Logger())
	s.app.Use(Recover())
	s.app.Use(middleware.RequestID())
	// Custom error handler
	s.app.Settings.ErrorHandler = errorHandler
}

func (s *Server) applyRoutes() {
	s.app.Get("/", s.hello)
	v1Api := s.app.Group("/api/v1")
	v1Api.Post("/auth/register", handler(s.register))
	v1Api.Post("/auth/login", handler(s.login))
}

// handler is a wrapper that allows the the server route functions to return
// an error. This is useful, because otherwise you would have to do the call
// to the Next handler call on each error. Ain't no body got time for that
func handler(h Handler) fiber.Handler {
	return func(ctx *fiber.Ctx) {
		if err := h(ctx); err != nil {
			ctx.Next(err)
			return
		}
	}
}

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
		logError(ctx.Status(400).JSON(errResponse{
			Error: "Content-Type: application/json header is required",
			Code: 400,
		}))
		return
	}

	// this is one of those, we have no choice
	if has := strings.HasPrefix(err.Error(), "pq: duplicate key value violates unique constraint"); has {
		logError(ctx.Status(409).JSON(errResponse{
			Error: "record already exists",
			Code: 409,
		}))
		return
	}

	e, ok := err.(PublicError)
	if !ok {
		zap.L().Info("masking internal error: ", zap.Error(err))
		logError(ctx.Status(500).JSON(errResponse{
			Error: "internal error",
			Code: 500,
		}))
		return
	}

	logError(ctx.Status(e.Code()).JSON(errResponse{
		Error: e.PublicError(),
		Code: e.Code(),
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

func (s *Server) hello(ctx *fiber.Ctx) {
	ctx.Send("Hello World!")
}

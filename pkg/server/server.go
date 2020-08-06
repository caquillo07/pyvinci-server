package server

import (
	"fmt"
	"runtime/debug"

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

func NewServer(config *conf.Config, db *gorm.DB) *Server {
	srv := &Server{
		app:    fiber.New(),
		config: config,
		db: db,
	}

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

func (s *Server) applyRoutes() {
	s.app.Use(middleware.Logger())
	s.app.Use(Recover())
	s.app.Use(middleware.RequestID())
	// Custom error handler
	s.app.Settings.ErrorHandler = func(ctx *fiber.Ctx, err error) {
		fmt.Println(err)
		type errResponse struct {
			Error string `json:"error"`
		}
		e, ok := err.(publicError)
		if !ok {
			zap.L().Info("masking internal error: ", zap.Error(err))
			_ = ctx.Status(500).JSON(errResponse{Error: "internal error"})
			return
		}

		_ = ctx.Status(e.Code()).JSON(errResponse{Error: e.PublicError()})
	}

	s.app.Get("/", s.hello)
	v1Api := s.app.Group("/api/v1")
	v1Api.Post("/auth/register", s.register)
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

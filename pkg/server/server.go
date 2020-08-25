package server

import (
	"github.com/gofiber/cors"
	"github.com/gofiber/fiber"
	"github.com/gofiber/fiber/middleware"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"

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
	s.app.Use(cors.New())
	// Custom error handler
	s.app.Settings.ErrorHandler = errorHandler
}

func (s *Server) applyRoutes() {
	s.app.Get("/", func(c *fiber.Ctx) {
		c.Send("Hello World!")
	})
	v1Api := s.app.Group("/api/v1")
	v1Api.Post("/auth/register", handler(s.register))
	v1Api.Post("/auth/login", handler(s.login))

	// protected endpoints
	v1Api.Use(s.protected())
	v1Api.Post("/users/:user_id/projects", handler(s.createProject))
	v1Api.Get("/users/:user_id/projects", handler(s.getProjects))
	v1Api.Get("/users/:user_id/projects/:project_id", handler(s.getProject))
	v1Api.Put("/users/:user_id/projects/:project_id", handler(s.updateProject))
	v1Api.Delete("/users/:user_id/projects/:project_id", handler(s.deleteProject))
	v1Api.Post("/users/:user_id/projects/:project_id/images", handler(s.postProjectImage))
	v1Api.Get("/users/:user_id/projects/:project_id/images", handler(s.getProjectImages))
	v1Api.Get("/users/:user_id/projects/:project_id/images/:image_id", handler(s.getProjectImage))
	v1Api.Delete("/users/:user_id/projects/:project_id/images/:image_id", handler(s.deleteProjectImage))
	v1Api.Post("/users/:user_id/projects/:project_id/job", handler(s.startProjectJob))
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

func getUserID(c *fiber.Ctx) (uuid.UUID, error) {
	userID, err := uuid.FromString(c.Params("user_id"))
	if err != nil {
		return uuid.Nil, newValidationError("valid user_id is required")
	}
	return userID, nil
}

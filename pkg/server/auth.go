package server

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber"
)

func (s *Server) register(c *fiber.Ctx) {
	type RegisterRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	type RegisterResponse struct {
		ID        string    `json:"id"`
		Username  string    `json:"username"`
		CreateAt  time.Time `json:"createAt"`
		UpdatedAt time.Time `json:"updatedAt"`
	}

	var req RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		c.Next(err)
		return
	}

	if req.Username == "" || req.Password == "" {
		c.Next(newValidationError("both username and password are required"))
		return
	}


	fmt.Println(req)
}

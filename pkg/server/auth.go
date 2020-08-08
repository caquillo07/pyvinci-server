package server

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gofiber/fiber"
	"golang.org/x/crypto/bcrypt"

	"github.com/caquillo07/pyvinci-server/pkg/model"
)

func (s *Server) register(c *fiber.Ctx) error {
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
		return err
	}

	if req.Username == "" || req.Password == "" {
		return newValidationError("both username and password are required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), 10)
	if err != nil {
		return err
	}

	user := model.User{
		Username: req.Username,
	}

	if err := model.CreateUser(s.db, &user, string(hashedPassword)); err != nil {
		return err
	}

	return c.Status(201).JSON(&RegisterResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreateAt:  user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (s *Server) login(c *fiber.Ctx) error {
	type LoginRequest struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	type LoginResponse struct {
		ID        string    `json:"id"`
		Username  string    `json:"username"`
		CreateAt  time.Time `json:"createAt"`
		UpdatedAt time.Time `json:"updatedAt"`
		Token     string    `json:"token"`
		ExpireAt  int64     `json:"expireAt"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	if req.Username == "" || req.Password == "" {
		return newValidationError("both username and password are required")
	}

	user, err := model.FindUserByUsername(s.db, req.Username)
	if err != nil {
		return err
	}

	err, valid := user.VerifyPassword(s.db, req.Password)
	if err != nil {
		return err
	}
	if !valid {
		return newNotFoundError("user with give username and pass not found")
	}

	token := jwt.New(jwt.SigningMethodHS256)

	// bonkers, but whatever its a demo
	expireAt := time.Now().Add(time.Hour * 365).Unix()
	claims := token.Claims.(jwt.MapClaims)
	claims["username"] = user.Username
	claims["userId"] = user.ID.String()
	claims["exp"] = expireAt

	t, err := token.SignedString([]byte(s.config.Auth.TokenSecret))
	if err != nil {
		return err
	}

	// invalidate all old tokens before creating a new one
	if err := model.InvalidateAllTokens(s.db, user.ID); err != nil {
		return err
	}

	if err := model.CreateToken(s.db, user.ID, t); err != nil {
		return err
	}

	return c.JSON(&LoginResponse{
		ID:        user.ID.String(),
		Username:  user.Username,
		CreateAt:  user.CreatedAt,
		UpdatedAt: user.CreatedAt,
		Token:     t,
		ExpireAt:  expireAt,
	})
}

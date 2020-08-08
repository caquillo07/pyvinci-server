package server

import (
	"net/http"
	"time"

	"github.com/gofiber/fiber"
	"github.com/gofrs/uuid"

	"github.com/caquillo07/pyvinci-server/pkg/model"
)

type httpProject struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Keywords  []string  `json:"keywords"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func projectHttpStruct(p *model.Project) *httpProject {
	return &httpProject{
		ID:        p.ID.String(),
		UserID:    p.UserID.String(),
		Keywords:  p.Keywords,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func (s *Server) createProject(c *fiber.Ctx) error {
	type CreateRequest struct {
		Keywords []string `json:"keywords"`
	}
	type CreateResponse struct {
		Project *httpProject `json:"project"`
	}

	userID, err := getUserID(c)
	if err != nil {
		return newValidationError("valid user_id is required")
	}

	var req CreateRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	user, err := model.FindUserByID(s.db, userID)
	if err != nil {
		return err
	}

	newProject := &model.Project{
		Keywords: req.Keywords,
		UserID:   user.ID,
	}
	if err := model.CreateProject(s.db, newProject); err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(CreateResponse{
		Project: projectHttpStruct(newProject),
	})
}

// returns a list of all projects for the given user.
// this endpoint does not support pagination as its not required for
// this application
func (s *Server) getProjects(c *fiber.Ctx) error {
	type GetResponse struct {
		Projects []*httpProject `json:"projects"`
	}
	userID, err := getUserID(c)
	if err != nil {
		return newValidationError("valid user_id is required")
	}

	user, err := model.FindUserByID(s.db, userID)
	if err != nil {
		return err
	}

	projects, err := model.AllProjectsForUser(s.db, user.ID)
	if err != nil {
		return err
	}

	res := GetResponse{
		Projects: make([]*httpProject, len(projects)),
	}

	for i, p := range projects {
		res.Projects[i] = projectHttpStruct(p)
	}

	return c.JSON(res)
}

func (s *Server) getProject(c *fiber.Ctx) error {
	type GetResponse struct {
		Project *httpProject `json:"project"`
	}
	userID, err := getUserID(c)
	if err != nil {
		return newValidationError("valid user_id is required")
	}

	projectID, err := uuid.FromString(c.Params("project_id"))
	if err != nil {
		return newValidationError("valid project_id is required")
	}

	user, err := model.FindUserByID(s.db, userID)
	if err != nil {
		return err
	}

	project, err := model.FindProjectByID(s.db, projectID)
	if err != nil {
		return err
	}

	if project.UserID != user.ID {
		return newNotFoundError("project not found")
	}

	return c.JSON(GetResponse{
		Project: projectHttpStruct(project),
	})
}

func (s *Server) deleteProject(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return newValidationError("valid user_id is required")
	}

	projectID, err := uuid.FromString(c.Params("project_id"))
	if err != nil {
		return newValidationError("valid project_id is required")
	}

	user, err := model.FindUserByID(s.db, userID)
	if err != nil {
		return err
	}

	project, err := model.FindProjectByID(s.db, projectID)
	if err != nil {
		return err
	}

	if project.UserID != user.ID {
		return newNotFoundError("project not found")
	}

	if err := model.DeleteProjectByID(s.db, project.ID); err != nil {
		return err
	}

	c.Status(200).Send()
	return nil
}

func (s *Server) updateProject(c *fiber.Ctx) error {
	type UpdateRequest struct {
		Keywords []string `json:"keywords"`
	}

	type CreateResponse struct {
		Project *httpProject `json:"project"`
	}
	userID, err := getUserID(c)
	if err != nil {
		return newValidationError("valid user_id is required")
	}

	projectID, err := uuid.FromString(c.Params("project_id"))
	if err != nil {
		return newValidationError("valid project_id is required")
	}



	var req UpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return err
	}

	user, err := model.FindUserByID(s.db, userID)
	if err != nil {
		return err
	}

	project, err := model.FindProjectByID(s.db, projectID)
	if err != nil {
		return err
	}

	if project.UserID != user.ID {
		return newNotFoundError("project not found")
	}

	project.Keywords = req.Keywords
	if err := project.Update(s.db); err != nil {
		return err
	}

	return c.JSON(CreateResponse{
		Project: projectHttpStruct(project),
	})
}

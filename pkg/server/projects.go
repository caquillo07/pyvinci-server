package server

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gofiber/fiber"
	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
	"go.uber.org/zap"

	"github.com/caquillo07/pyvinci-server/pkg/model"
)

type httpProject struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Name      string    `json:"name"`
	Keywords  []string  `json:"keywords"`
	Labels    []string  `json:"labels"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type httpImage struct {
	ID        string    `json:"id"`
	URL       string    `json:"url"`
	ProjectID string    `json:"projectId"`
	Labels    []string  `json:"labels,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func projectHTTPStruct(p *model.Project) *httpProject {
	return &httpProject{
		ID:        p.ID.String(),
		UserID:    p.UserID.String(),
		Name:      p.Name,
		Keywords:  p.Keywords,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func imageHTTPStruct(img *model.Image) *httpImage {
	uniqueLabels := map[string]struct{}{}
	assignLabels := func(s []string) {
		for _, l := range s {
			uniqueLabels[l] = struct{}{}
		}
	}
	assignLabels(img.MasksLabels)
	assignLabels(img.LabelsStuff)
	labels := make([]string, 0)
	for l, _ := range uniqueLabels {
		labels = append(labels, l)
	}

	return &httpImage{
		ID:        img.ID.String(),
		URL:       img.URL,
		Labels:    labels,
		ProjectID: img.ProjectID.String(),
		CreatedAt: img.CreatedAt,
		UpdatedAt: img.UpdatedAt,
	}
}

func (s *Server) createProject(c *fiber.Ctx) error {
	type CreateRequest struct {
		Name     string   `json:"name"`
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
		Name:     req.Name,
	}
	if err := model.CreateProject(s.db, newProject); err != nil {
		return err
	}

	return c.Status(http.StatusCreated).JSON(CreateResponse{
		Project: projectHTTPStruct(newProject),
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
		res.Projects[i] = projectHTTPStruct(p)
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

	projectRes := projectHTTPStruct(project)

	// see if the project has a pending job to fill out the information
	job, err := model.FindJobForProject(s.db, projectID)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	if job != nil {
		projectRes.Status = job.Status
	}

	return c.JSON(GetResponse{
		Project: projectRes,
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
		Project: projectHTTPStruct(project),
	})
}

func (s *Server) postProjectImage(c *fiber.Ctx) error {
	type CreateResponse struct {
		Images []*httpImage `json:"images"`
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

	s3Client, err := s.s3Client()
	if err != nil {
		return err
	}

	// Parse the multipart form:
	form, err := c.MultipartForm()
	if err != nil {
		return err
	}

	// Get all files from "documents" key:
	// Loop through files:
	images := make([]*model.Image, 0)
	for _, fileHeader := range form.File["images"] {
		contentType := fileHeader.Header["Content-Type"][0]
		zap.L().Info(
			"upload image to S3",
			zap.String("file_name", fileHeader.Filename),
			zap.Int64("file_size", fileHeader.Size),
			zap.String("file_type", contentType),
		)

		mpFile, err := fileHeader.Open()
		if err != nil {
			return err
		}

		// Save the files to disk:
		// err := c.SaveFile(file, fmt.Sprintf("./%s", file.Filename))
		// Check for errors
		imageKey := s3ImageKey(
			user.ID,
			project.ID,
			uuid.Must(uuid.NewV4()).String()+"_"+fileHeader.Filename,
		)
		_, err = s3Client.PutObject(&s3.PutObjectInput{
			Bucket:       &s.config.S3.ImageBucket,
			Key:          &imageKey,
			Body:         mpFile,
			ACL:          aws.String(s3.BucketCannedACLPublicRead),
			ContentType:  &contentType,
			CacheControl: aws.String("no-cache"),
		})
		if err != nil {
			if err := mpFile.Close(); err != nil {
				// log and continue
				zap.L().Error(
					"error closing fileHeader",
					zap.Error(err),
					zap.String("file_name", fileHeader.Filename),
				)
			}
			return err
		}
		if err := mpFile.Close(); err != nil {
			return err
		}

		// now save the records in the DB and lets go to the next. I would
		// normally put this io intensive stuff in its own goroutines, but for
		// the sake of time since this is a demo project, serially works.
		image := &model.Image{
			URL:       s3ImageURL(s.config.S3.ImageBucket, imageKey),
			ProjectID: project.ID,
		}
		if err := model.CreateImage(s.db, image); err != nil {
			return err
		}
		images = append(images, image)
	}
	response := make([]*httpImage, len(images))
	for i, img := range images {
		response[i] = imageHTTPStruct(img)
	}
	return c.Status(http.StatusCreated).JSON(CreateResponse{
		Images: response,
	})
}

func (s *Server) s3Client() (*s3.S3, error) {
	sess, err := session.NewSession(&aws.Config{
		Region: aws.String("us-east-1"),
		Credentials: credentials.NewStaticCredentials(
			s.config.S3.AccessKey,
			s.config.S3.SecretKey,
			"",
		),
	})
	if err != nil {
		return nil, err
	}

	// Create S3 service client
	return s3.New(sess), nil
}

func (s *Server) getProjectImages(c *fiber.Ctx) error {
	type GetResponse struct {
		Images []*httpImage `json:"images"`
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

	images, err := model.AllImagesForProject(s.db, projectID)
	if err != nil {
		return nil
	}

	httpImages := make([]*httpImage, len(images))
	for i, img := range images {
		httpImages[i] = imageHTTPStruct(img)
	}
	return c.JSON(GetResponse{
		Images: httpImages,
	})
}

func (s *Server) getProjectImage(c *fiber.Ctx) error {
	type GetResponse struct {
		Image *httpImage `json:"image"`
	}
	userID, err := getUserID(c)
	if err != nil {
		return newValidationError("valid user_id is required")
	}

	projectID, err := uuid.FromString(c.Params("project_id"))
	if err != nil {
		return newValidationError("valid project_id is required")
	}

	imageID, err := uuid.FromString(c.Params("image_id"))
	if err != nil {
		return newValidationError("valid image_id is required")
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

	image, err := model.FindImageByID(s.db, imageID)
	if err != nil {
		return err
	}

	return c.JSON(GetResponse{
		Image: imageHTTPStruct(image),
	})
}

func (s *Server) deleteProjectImage(c *fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return newValidationError("valid user_id is required")
	}

	projectID, err := uuid.FromString(c.Params("project_id"))
	if err != nil {
		return newValidationError("valid project_id is required")
	}

	imageID, err := uuid.FromString(c.Params("image_id"))
	if err != nil {
		return newValidationError("valid image_id is required")
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

	image, err := model.FindImageByID(s.db, imageID)
	if err != nil {
		return nil
	}

	// first delete image from s3, then delete from DB
	s3Client, err := s.s3Client()
	if err != nil {
		return err
	}

	_, err = s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &s.config.S3.ImageBucket,
		Key:    aws.String(strings.TrimPrefix(image.URL, s3BucketURL(s.config.S3.ImageBucket))),
	})
	if err != nil {
		return err
	}

	if err := model.DeleteImageByID(s.db, image.ID); err != nil {
		return err
	}

	c.Status(200).Send()
	return nil
}

func (s *Server) startProjectJob(c *fiber.Ctx) error {
	type PostResponse struct {
		JobID  string `json:"jobId"`
		Status string `json:"status"`
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

	// make sure there is no jobs already
	job, err := model.FindJobForProject(s.db, projectID)
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return err
	}

	if job != nil {
		return newValidationError("job already exists for this project")
	}

	newJob, err := model.CreateNewJob(s.db, projectID)
	if err != nil {
		return nil
	}

	return c.Status(201).JSON(PostResponse{
		Status: newJob.Status,
		JobID:  newJob.ID.String(),
	})
}

func s3BucketURL(bucketName string) string {
	return fmt.Sprintf("https://%s.s3.amazonaws.com/", bucketName)
}

func s3ImageURL(bucketName, imgKey string) string {
	return fmt.Sprintf("%s%s", s3BucketURL(bucketName), imgKey)
}

func s3ImageKey(userID, projectID uuid.UUID, imgName string) string {
	return fmt.Sprintf("users/%s/projects/%s/%s", userID.String(), projectID.String(), imgName)
}

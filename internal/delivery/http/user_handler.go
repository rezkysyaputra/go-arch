package http

import (
	nethttp "net/http"
	"strconv"
	"time"

	"go-arch/internal/domain"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	users domain.UserUsecase
}

func NewUserHandler(users domain.UserUsecase) *UserHandler {
	return &UserHandler{users: users}
}

// Create godoc
// @Summary Create user
// @Description Create a new user and publish a user.created event.
// @Tags users
// @Accept json
// @Produce json
// @Param request body createUserRequest true "Create user request"
// @Success 201 {object} apiResponse{data=userResponse}
// @Failure 400 {object} apiResponse
// @Failure 409 {object} apiResponse
// @Failure 500 {object} apiResponse
// @Router /users [post]
func (h *UserHandler) Create(ctx *gin.Context) {
	var request createUserRequest
	if err := ctx.ShouldBindJSON(&request); err != nil {
		respondValidationError(ctx, err)
		return
	}

	user, err := h.users.Create(ctx.Request.Context(), domain.CreateUserInput{
		Name:  request.Name,
		Email: request.Email,
	})
	if err != nil {
		respondDomainError(ctx, err)
		return
	}

	respondSuccess(ctx, nethttp.StatusCreated, "user created successfully", newUserResponse(user))
}

// GetByID godoc
// @Summary Get user by ID
// @Description Retrieve a user by its numeric ID.
// @Tags users
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} apiResponse{data=userResponse}
// @Failure 400 {object} apiResponse
// @Failure 404 {object} apiResponse
// @Failure 500 {object} apiResponse
// @Router /users/{id} [get]
func (h *UserHandler) GetByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		respondFailure(ctx, nethttp.StatusBadRequest, "invalid request", []errorDetail{
			{Field: "id", Code: "invalid_id", Message: "id must be a positive integer"},
		})
		return
	}

	user, err := h.users.GetByID(ctx.Request.Context(), uint(id))
	if err != nil {
		respondDomainError(ctx, err)
		return
	}

	respondSuccess(ctx, nethttp.StatusOK, "user retrieved successfully", newUserResponse(user))
}

type createUserRequest struct {
	Name  string `json:"name" binding:"required,min=2,max=255"`
	Email string `json:"email" binding:"required,email,max=255"`
}

type userResponse struct {
	ID        uint   `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func newUserResponse(user *domain.User) userResponse {
	return userResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format(time.RFC3339),
		UpdatedAt: user.UpdatedAt.Format(time.RFC3339),
	}
}

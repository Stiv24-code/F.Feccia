package handlers

import (
	"github.com/gofiber/fiber/v2"
	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/services"
	"fratelli-feccia/pkg/utils"
)

type UserHandler struct {
	Service services.AdminService
}

func NewUserHandler(service services.AdminService) *UserHandler {
	return &UserHandler{Service: service}
}

// ListUsers godoc
// @Summary List users
// @Description Returns paginated list of users
// @Tags Admin Users
// @Security BearerAuth
// @Produce json
// @Param page  query int false "Page number (default 1)"
// @Param limit query int false "Items per page (default 20, max 100)"
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/users-list [get]
func (h *UserHandler) ListUsers(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 20)

	users, total, err := h.Service.ListUsers(ctx, page, limit)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}

	return utils.SuccessResponse(c, 200, utils.ListResult{Data: users, Total: total})
}

// ListAllUsers godoc
// @Summary List all users (frontend admin panel)
// @Description Full unpaginated roster with email/profile_id/active, mirrors admin.py's GET /admin/users
// @Tags Admin Users
// @Security BearerAuth
// @Produce json
// @Success 200 {array} dto.AuthUserResponse
// @Router /api/v1/admin/users [get]
func (h *UserHandler) ListAllUsers(c *fiber.Ctx) error {
	users, err := h.Service.ListAllUsers(utils.RequestContext(c))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, users)
}

// PatchUser godoc
// @Summary Partially update a user (frontend admin panel)
// @Description Updates name/active only, mirrors admin.py's PATCH /admin/users/{id}
// @Tags Admin Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.PatchUserRequest true "Fields to update"
// @Success 200 {object} dto.AuthUserResponse
// @Failure 400 {object} map[string]string
// @Router /api/v1/admin/users/{id} [patch]
func (h *UserHandler) PatchUser(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)

	id, err := utils.GetUintParam(c, "id")
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}

	var req dto.PatchUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	updated, err := h.Service.PatchUser(ctx, int64(id), req)
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}
	return utils.SuccessResponse(c, 200, updated)
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Retrieve a single user by ID
// @Tags Admin Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Router /api/v1/admin/users/{id} [get]
func (h *UserHandler) GetUserByID(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)

	id, err := utils.GetUintParam(c, "id")
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}

	user, err := h.Service.GetUserByID(ctx, int64(id))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}

	if user == nil {
		return utils.ErrorResponse(c, 404, "User not found")
	}

	return utils.SuccessResponse(c, 200, user)
}

// CreateUser godoc
// @Summary Create user
// @Description Create a new user
// @Tags Admin Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param user body dto.CreateUserRequest true "User data"
// @Success 201 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /api/v1/admin/users [post]
func (h *UserHandler) CreateUser(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)

	var req dto.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}

	user, err := h.Service.CreateUser(ctx, req)
	if err != nil {
		return utils.ErrorResponse(c, 422, err.Error())
	}

	return utils.SuccessResponse(c, 201, user)
}

// UpdateUser godoc
// @Summary Update user
// @Description Update existing user
// @Tags Admin Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body dto.UpdateUserRequest true "User data"
// @Success 200 {object} models.User
// @Failure 400 {object} map[string]string
// @Failure 422 {object} map[string]string
// @Router /api/v1/admin/users/{id} [put]
func (h *UserHandler) UpdateUser(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)

	id, err := utils.GetUintParam(c, "id")
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}

	var req dto.UpdateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return utils.ErrorResponse(c, 400, "Invalid request body")
	}

	if validationErrors := utils.NewValidator().Validate(&req); len(validationErrors) > 0 {
		return utils.ValidationErrorResponse(c, validationErrors)
	}

	updated, err := h.Service.UpdateUser(ctx, int64(id), req)
	if err != nil {
		return utils.ErrorResponse(c, 422, err.Error())
	}

	return utils.SuccessResponse(c, 200, updated)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete user by ID
// @Tags Admin Users
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /api/v1/admin/users/{id} [delete]
func (h *UserHandler) DeleteUser(c *fiber.Ctx) error {
	ctx := utils.RequestContext(c)

	id, err := utils.GetUintParam(c, "id")
	if err != nil {
		return utils.ErrorResponse(c, 400, "Invalid ID")
	}

	user, err := h.Service.GetUserByID(ctx, int64(id))
	if err != nil {
		return utils.HandleDatabaseError(c, err)
	}

	if user.Role == utils.RoleAdmin {
		count, err := h.Service.CountAdmins(ctx)
		if err != nil {
			return utils.HandleDatabaseError(c, err)
		}

		if count <= 1 {
			return utils.ErrorResponse(c, 400, "Cannot delete the last admin")
		}
	}

	if err := h.Service.DeleteUser(ctx, int64(id)); err != nil {
		return utils.HandleDatabaseError(c, err)
	}

	return c.SendStatus(204)
}

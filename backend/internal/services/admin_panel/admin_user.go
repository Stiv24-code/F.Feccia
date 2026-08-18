package services

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"fratelli-feccia/internal/dto"
	"fratelli-feccia/internal/models"
	"fratelli-feccia/pkg/utils"

	"golang.org/x/crypto/bcrypt"
)

type AdminService struct {
	db      *gorm.DB
	jwtConf utils.JWTConfig
}

func NewAdminService(db *gorm.DB, jwtConf utils.JWTConfig) *AdminService {
	return &AdminService{
		db:      db,
		jwtConf: jwtConf,
	}
}

func (s *AdminService) GetUserByID(ctx context.Context, id int64) (*dto.UserResponse, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toUserResponse(user), nil
}

func (s *AdminService) ListUsers(ctx context.Context, page, limit int) ([]dto.UserResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	db := s.db.WithContext(ctx).Model(&models.User{})

	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var users []models.User
	offset := (page - 1) * limit
	if err := db.Order("id ASC").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	result := make([]dto.UserResponse, len(users))
	for i, u := range users {
		result[i] = *toUserResponse(u)
	}
	return result, total, nil
}

// CreateUser is a pre-existing admin-panel path the frontend never actually
// calls (UsersPage.tsx uses POST /auth/register — see AuthService.Register
// — not this one); left as-is, not worth extending for the cliente-role
// case only the live path needs.
func (s *AdminService) CreateUser(ctx context.Context, req dto.CreateUserRequest) (*dto.UserResponse, error) {
	if !utils.IsValidRole(req.Role) {
		return nil, fmt.Errorf("invalid role")
	}

	var count int64
	s.db.WithContext(ctx).Model(&models.User{}).Where("login = ?", req.Login).Count(&count)
	if count > 0 {
		return nil, errors.New("login already exists")
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	now := time.Now()
	user := models.User{
		Login:           req.Login,
		Name:            req.Name,
		PasswordHash:    string(hash),
		Role:            req.Role,
		EmailVerifiedAt: &now,
	}

	if err := s.db.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func (s *AdminService) UpdateUser(ctx context.Context, id int64, req dto.UpdateUserRequest) (*dto.UserResponse, error) {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}

	if user.Role == utils.RoleAdmin && req.Role != utils.RoleAdmin {
		count, err := s.CountAdmins(ctx)
		if err != nil {
			return nil, err
		}
		if count <= 1 {
			req.Role = utils.RoleAdmin
		}
	}

	if !utils.IsValidRole(req.Role) {
		return nil, fmt.Errorf("invalid role")
	}

	user.Login = req.Login
	user.Name = req.Name
	user.Role = req.Role

	if req.Password != nil && *req.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(*req.Password), bcrypt.DefaultCost)
		user.PasswordHash = string(hash)
	}

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}

	return toUserResponse(user), nil
}

func (s *AdminService) CountAdmins(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.WithContext(ctx).
		Model(&models.User{}).
		Where("role = ?", utils.RoleAdmin).
		Count(&count).Error
	return count, err
}

func (s *AdminService) DeleteUser(ctx context.Context, id int64) error {
	var user models.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return err
	}

	if user.Role == utils.RoleAdmin {
		count, err := s.CountAdmins(ctx)
		if err != nil {
			return err
		}
		if count <= 1 {
			return errors.New("cannot delete the last admin")
		}
	}

	return s.db.WithContext(ctx).Delete(&user).Error
}

func toUserResponse(u models.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:        u.ID,
		Login:     u.Login,
		Name:      u.Name,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// ListAllUsers mirrors admin.py's admin_list_users: the full unpaginated
// roster (email/profile_id/active shape), as consumed by the frontend's
// UsersPage. Distinct from ListUsers (Login-shaped, paginated), which the
// frontend never calls.
func (s *AdminService) ListAllUsers(ctx context.Context) ([]dto.AuthUserResponse, error) {
	var users []models.User
	if err := s.db.WithContext(ctx).Order("id ASC").Find(&users).Error; err != nil {
		return nil, err
	}
	result := make([]dto.AuthUserResponse, len(users))
	for i, u := range users {
		result[i] = toAdminUserResponse(u)
	}
	return result, nil
}

// PatchUser mirrors admin.py's admin_update_user (PATCH /admin/users/{id}):
// partial update of name/active. ProfileID is rejected when non-empty since
// there is no profiles table to validate it against (see plan: profiles are
// deliberately out of scope).
func (s *AdminService) PatchUser(ctx context.Context, id int64, req dto.PatchUserRequest) (*dto.AuthUserResponse, error) {
	if req.ProfileID != nil && *req.ProfileID != "" {
		return nil, utils.NewAPIError(400, "Profilo non trovato")
	}
	if req.Name == nil && req.Active == nil {
		return nil, utils.NewAPIError(400, "Nessun campo da aggiornare")
	}

	var user models.User
	if err := s.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}

	if req.Active != nil && !*req.Active && user.Role == utils.RoleAdmin {
		count, err := s.CountAdmins(ctx)
		if err != nil {
			return nil, err
		}
		if count <= 1 {
			return nil, utils.NewAPIError(400, "Impossibile disattivare l'unico amministratore")
		}
	}

	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.Active != nil {
		user.Active = *req.Active
	}

	if err := s.db.WithContext(ctx).Save(&user).Error; err != nil {
		return nil, err
	}

	resp := toAdminUserResponse(user)
	return &resp, nil
}

func toAdminUserResponse(u models.User) dto.AuthUserResponse {
	return dto.AuthUserResponse{
		ID: u.ID, Email: u.Login, Name: u.Name, Role: u.Role, ProfileID: nil, Active: u.Active,
	}
}

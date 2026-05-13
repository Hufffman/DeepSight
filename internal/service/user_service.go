package service

import (
	"errors"

	"DeepSight/internal/dto"
	"DeepSight/internal/model"
	"DeepSight/internal/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(username, password, email string) (*model.User, error) {
	if _, err := s.repo.GetByUsername(username); err == nil {
		return nil, errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if _, err := s.repo.GetByEmail(email); err == nil {
		return nil, errors.New("email already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}

	if err := s.repo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	return s.repo.GetByID(id)
}

func (s *UserService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	page = normalizePage(page)
	pageSize = normalizePageSize(pageSize)
	return s.repo.List(page, pageSize)
}

func (s *UserService) UpdateUser(id uint, username, email string) (*model.User, error) {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}

	if username != "" && username != user.Username {
		if existing, err := s.repo.GetByUsername(username); err == nil && existing.ID != id {
			return nil, errors.New("username already exists")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		user.Username = username
	}

	if email != "" && email != user.Email {
		if existing, err := s.repo.GetByEmail(email); err == nil && existing.ID != id {
			return nil, errors.New("email already exists")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		user.Email = email
	}

	if err := s.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *UserService) UpdatePassword(id uint, oldPassword, newPassword string) error {
	user, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.Password = string(hashedPassword)
	return s.repo.Update(user)
}

func (s *UserService) ListUserResponses(page, pageSize int) (*dto.ListUsersResponse, error) {
	users, total, err := s.ListUsers(page, pageSize)
	if err != nil {
		return nil, err
	}

	responses := make([]dto.UserResponse, len(users))
	for i := range users {
		responses[i] = ToUserResponse(&users[i])
	}

	return &dto.ListUsersResponse{
		Users:    responses,
		Total:    total,
		Page:     normalizePage(page),
		PageSize: normalizePageSize(pageSize),
	}, nil
}

func ToUserResponse(user *model.User) dto.UserResponse {
	return dto.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt: user.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

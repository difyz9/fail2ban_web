package service

import (
	"fail2ban-web/internal/model"
	"gorm.io/gorm"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// CreateUser 创建用户
func (s *UserService) CreateUser(user *model.User) error {
	return s.db.Create(user).Error
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := s.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := s.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*model.User, error) {
	var user model.User
	err := s.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *model.User) error {
	return s.db.Save(user).Error
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	return s.db.Delete(&model.User{}, id).Error
}

// ListUsers 获取用户列表
func (s *UserService) ListUsers(offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	// 获取总数
	if err := s.db.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := s.db.Offset(offset).Limit(limit).Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ActivateUser 激活用户
func (s *UserService) ActivateUser(id uint) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).Update("is_active", true).Error
}

// DeactivateUser 禁用用户
func (s *UserService) DeactivateUser(id uint) error {
	return s.db.Model(&model.User{}).Where("id = ?", id).Update("is_active", false).Error
}
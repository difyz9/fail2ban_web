package service

import (
	"fail2ban-web/internal/model"

	"gorm.io/gorm"
)

type JailService struct {
	db *gorm.DB
}

func NewJailService(db *gorm.DB) *JailService {
	return &JailService{
		db: db,
	}
}

// CreateJail 创建jail配置
func (s *JailService) CreateJail(jail *model.Fail2banJail) error {
	return s.db.Create(jail).Error
}

// GetJailByID 根据ID获取jail配置
func (s *JailService) GetJailByID(id uint) (*model.Fail2banJail, error) {
	var jail model.Fail2banJail
	err := s.db.First(&jail, id).Error
	return &jail, err
}

// GetJailByName 根据名称获取jail配置
func (s *JailService) GetJailByName(name string) (*model.Fail2banJail, error) {
	var jail model.Fail2banJail
	err := s.db.Where("name = ?", name).First(&jail).Error
	return &jail, err
}

// GetAllJails 获取所有jail配置
func (s *JailService) GetAllJails() ([]model.Fail2banJail, error) {
	var jails []model.Fail2banJail
	err := s.db.Find(&jails).Error
	return jails, err
}

// GetEnabledJails 获取启用的jail配置
func (s *JailService) GetEnabledJails() ([]model.Fail2banJail, error) {
	var jails []model.Fail2banJail
	err := s.db.Where("enabled = ?", true).Find(&jails).Error
	return jails, err
}

// UpdateJail 更新jail配置
func (s *JailService) UpdateJail(jail *model.Fail2banJail) error {
	return s.db.Save(jail).Error
}

// DeleteJail 删除jail配置
func (s *JailService) DeleteJail(id uint) error {
	return s.db.Delete(&model.Fail2banJail{}, id).Error
}

// EnableJail 启用jail
func (s *JailService) EnableJail(id uint) error {
	return s.db.Model(&model.Fail2banJail{}).Where("id = ?", id).Update("enabled", true).Error
}

// DisableJail 禁用jail
func (s *JailService) DisableJail(id uint) error {
	return s.db.Model(&model.Fail2banJail{}).Where("id = ?", id).Update("enabled", false).Error
}

// ListJails 分页获取jail配置
func (s *JailService) ListJails(offset, limit int) ([]model.Fail2banJail, int64, error) {
	var jails []model.Fail2banJail
	var total int64

	// 获取总数
	if err := s.db.Model(&model.Fail2banJail{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 获取分页数据
	err := s.db.Offset(offset).Limit(limit).Find(&jails).Error
	return jails, total, err
}
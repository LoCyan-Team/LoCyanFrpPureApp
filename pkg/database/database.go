// database/manager.go
package database

import (
	"context"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// ClosedProxyType 定义代理关闭类型
type ClosedProxyType string

const (
	ClosedTypeUser  ClosedProxyType = "user"  // 手动关闭
	ClosedTypeAdmin ClosedProxyType = "admin" // 管理关闭
)

type ClosedProxy struct {
	RunID     string          `gorm:"primaryKey;column:run_id" json:"run_id"`
	ProxyName string          `gorm:"column:proxy_name;index" json:"proxy_name,omitempty"`
	Type      ClosedProxyType `gorm:"column:type;index;default:'manual'" json:"type"`
}

func (ClosedProxy) TableName() string {
	return "closed_proxies"
}

type ClosedProxyManager struct {
	db *gorm.DB
}

func NewClosedProxyManager(dbPath string) (*ClosedProxyManager, error) {
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&ClosedProxy{}); err != nil {
		return nil, err
	}

	return &ClosedProxyManager{db: db}, nil
}

// IsClosed 检查代理是否已关闭，支持通过 runId 或 proxyName 查找
func (m *ClosedProxyManager) IsClosed(runId, proxyName string) (bool, error) {
	if runId == "" && proxyName == "" {
		return false, nil
	}

	var count int64
	query := m.db.Model(&ClosedProxy{})

	if runId != "" && proxyName != "" {
		// 同时提供 runId 和 proxyName，使用 OR 条件
		query = query.Where("run_id = ? OR proxy_name = ?", runId, proxyName)
	} else if runId != "" {
		// 只提供 runId
		query = query.Where("run_id = ?", runId)
	} else {
		// 只提供 proxyName
		query = query.Where("proxy_name = ?", proxyName)
	}

	err := query.Count(&count).Error
	return count > 0, err
}

// IsClosedByRunID 仅通过 runId 检查
func (m *ClosedProxyManager) IsClosedByRunID(runId string) (bool, error) {
	if runId == "" {
		return false, nil
	}

	var count int64
	err := m.db.Model(&ClosedProxy{}).
		Where("run_id = ?", runId).
		Count(&count).Error
	return count > 0, err
}

// IsClosedByProxyName 仅通过 proxyName 检查
func (m *ClosedProxyManager) IsClosedByProxyName(proxyName string) (bool, error) {
	if proxyName == "" {
		return false, nil
	}

	var count int64
	err := m.db.Model(&ClosedProxy{}).
		Where("proxy_name = ?", proxyName).
		Count(&count).Error
	return count > 0, err
}

// FindClosedRecords 查找关闭的记录，返回详细信息
func (m *ClosedProxyManager) FindClosedRecords(runId, proxyName string) ([]ClosedProxy, error) {
	if runId == "" && proxyName == "" {
		return nil, nil
	}

	var records []ClosedProxy
	query := m.db.Model(&ClosedProxy{})

	if runId != "" && proxyName != "" {
		query = query.Where("run_id = ? OR proxy_name = ?", runId, proxyName)
	} else if runId != "" {
		query = query.Where("run_id = ?", runId)
	} else {
		query = query.Where("proxy_name = ?", proxyName)
	}

	err := query.Find(&records).Error
	return records, err
}

// AddClosedProxy 添加已关闭的代理记录
func (m *ClosedProxyManager) AddClosedProxy(runID, proxyName string) error {
	record := &ClosedProxy{
		RunID:     runID,
		ProxyName: proxyName,
		Type:      ClosedTypeAdmin, // 默认手动关闭
	}
	return m.db.Create(record).Error
}

// AddClosedProxyWithType 添加已关闭的代理记录（指定类型）
func (m *ClosedProxyManager) AddClosedProxyWithType(runID, proxyName string, closeType ClosedProxyType) error {
	record := &ClosedProxy{
		RunID:     runID,
		ProxyName: proxyName,
		Type:      closeType,
	}
	return m.db.Create(record).Error
}

// GetAllClosedProxies 获取所有关闭的代理记录
func (m *ClosedProxyManager) GetAllClosedProxies() ([]ClosedProxy, error) {
	var records []ClosedProxy
	err := m.db.Find(&records).Error
	return records, err
}

// GetClosedProxiesByType 根据类型获取关闭的代理记录
func (m *ClosedProxyManager) GetClosedProxiesByType(closeType ClosedProxyType) ([]ClosedProxy, error) {
	var records []ClosedProxy
	err := m.db.Where("type = ?", closeType).Order("closed_at DESC").Find(&records).Error
	return records, err
}

// GetClosedProxiesPaginated 分页获取关闭的代理记录
func (m *ClosedProxyManager) GetClosedProxiesPaginated(page, pageSize int) ([]ClosedProxy, int64, error) {
	var records []ClosedProxy
	var total int64

	// 获取总数
	if err := m.db.Model(&ClosedProxy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}

	// 获取分页数据
	err := m.db.Order("closed_at DESC").Offset(offset).Limit(pageSize).Find(&records).Error
	return records, total, err
}

// GetStats 获取统计信息
func (m *ClosedProxyManager) GetStats(ctx context.Context) (map[ClosedProxyType]int64, int64, error) {
	stats := make(map[ClosedProxyType]int64)
	var total int64

	// 获取总数
	if err := m.db.Model(&ClosedProxy{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 按类型统计
	var typeStats []struct {
		Type  ClosedProxyType
		Count int64
	}

	err := m.db.Model(&ClosedProxy{}).
		Select("type, COUNT(*) as count").
		Group("type").
		Scan(&typeStats).Error

	if err != nil {
		return nil, 0, err
	}

	for _, stat := range typeStats {
		stats[stat.Type] = stat.Count
	}

	return stats, total, nil
}

// Close 关闭数据库连接
func (m *ClosedProxyManager) Close() error {
	sqlDB, err := m.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

// DeleteByRunID 根据 runID 删除记录
func (m *ClosedProxyManager) DeleteByRunID(runID string) error {
	result := m.db.Where("run_id = ?", runID).Delete(&ClosedProxy{})
	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}

	return nil
}

// DeleteByProxyName 根据代理名称删除记录
func (m *ClosedProxyManager) DeleteByProxyName(proxyName string) (int64, error) {
	result := m.db.Where("proxy_name = ?", proxyName).Delete(&ClosedProxy{})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

// DeleteByType 根据类型删除记录
func (m *ClosedProxyManager) DeleteByType(closeType ClosedProxyType) (int64, error) {
	result := m.db.Where("type = ?", closeType).Delete(&ClosedProxy{})
	if result.Error != nil {
		return 0, result.Error
	}

	return result.RowsAffected, nil
}

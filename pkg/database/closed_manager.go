package database

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	"go.etcd.io/bbolt"
)

// ClosedProxyType 定义代理关闭类型
type ClosedProxyType string

const (
	ClosedTypeUser  ClosedProxyType = "user"  // 手动关闭
	ClosedTypeAdmin ClosedProxyType = "admin" // 管理关闭

	bucketName = "closed_proxies"
)

type ClosedProxy struct {
	RunID    string          `json:"run_id"`
	Type     ClosedProxyType `json:"type"`
	ClosedAt int64           `json:"closed_at"` // 记录时间戳用于排序
}

type ClosedProxyManager struct {
	db *bbolt.DB
}

func NewClosedProxyManager(dbPath string) (*ClosedProxyManager, error) {
	// 设置超时时间，防止文件锁阻塞
	db, err := bbolt.Open(dbPath, 0600, &bbolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	// 初始化 Bucket
	err = db.Update(func(tx *bbolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		return err
	})

	if err != nil {
		return nil, err
	}

	return &ClosedProxyManager{db: db}, nil
}

// IsClosed 检查代理是否已关闭 (O(1) 性能极高)
func (m *ClosedProxyManager) IsClosed(runId string) (bool, error) {
	if runId == "" {
		return false, nil
	}

	found := false
	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		if b.Get([]byte(runId)) != nil {
			found = true
		}
		return nil
	})
	return found, err
}

// FindClosedRecords 查找关闭的记录
func (m *ClosedProxyManager) FindClosedRecords(runId string) ([]ClosedProxy, error) {
	if runId == "" {
		return nil, nil
	}

	var records []ClosedProxy
	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get([]byte(runId))
		if v != nil {
			var p ClosedProxy
			if err := json.Unmarshal(v, &p); err == nil {
				records = append(records, p)
			}
		}
		return nil
	})
	return records, err
}

// AddClosedProxy 添加已关闭的代理记录
func (m *ClosedProxyManager) AddClosedProxy(runID string) error {
	return m.AddClosedProxyWithType(runID, ClosedTypeAdmin)
}

// AddClosedProxyWithType 添加已关闭的代理记录（指定类型）
func (m *ClosedProxyManager) AddClosedProxyWithType(runID string, closeType ClosedProxyType) error {
	if runID == "" {
		return errors.New("runID is required")
	}

	record := &ClosedProxy{
		RunID:    runID,
		Type:     closeType,
		ClosedAt: time.Now().Unix(),
	}

	data, err := json.Marshal(record)
	if err != nil {
		return err
	}

	return m.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Put([]byte(runID), data)
	})
}

// GetAllClosedProxies 获取所有关闭的代理记录
func (m *ClosedProxyManager) GetAllClosedProxies() ([]ClosedProxy, error) {
	var records []ClosedProxy
	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var p ClosedProxy
			if err := json.Unmarshal(v, &p); err == nil {
				records = append(records, p)
			}
		}
		return nil
	})
	return records, err
}

// GetClosedProxiesByType 根据类型获取关闭的代理记录
func (m *ClosedProxyManager) GetClosedProxiesByType(closeType ClosedProxyType) ([]ClosedProxy, error) {
	var records []ClosedProxy
	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var p ClosedProxy
			if err := json.Unmarshal(v, &p); err == nil {
				if p.Type == closeType {
					records = append(records, p)
				}
			}
		}
		return nil
	})

	// 按时间倒序
	sort.Slice(records, func(i, j int) bool {
		return records[i].ClosedAt > records[j].ClosedAt
	})
	return records, err
}

// GetClosedProxiesPaginated 分页获取关闭的代理记录
func (m *ClosedProxyManager) GetClosedProxiesPaginated(page, pageSize int) ([]ClosedProxy, int64, error) {
	var allRecords []ClosedProxy

	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		// 预分配内存优化
		allRecords = make([]ClosedProxy, 0, b.Stats().KeyN)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var p ClosedProxy
			if err := json.Unmarshal(v, &p); err == nil {
				allRecords = append(allRecords, p)
			}
		}
		return nil
	})
	if err != nil {
		return nil, 0, err
	}

	total := int64(len(allRecords))

	// 按时间倒序排序
	sort.Slice(allRecords, func(i, j int) bool {
		return allRecords[i].ClosedAt > allRecords[j].ClosedAt
	})

	// 分页处理
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = 0
	}
	if int64(offset) >= total {
		return []ClosedProxy{}, total, nil
	}

	end := offset + pageSize
	if int64(end) > total {
		end = int(total)
	}

	return allRecords[offset:end], total, nil
}

// GetStats 获取统计信息
func (m *ClosedProxyManager) GetStats() (map[ClosedProxyType]int64, int64, error) {
	stats := make(map[ClosedProxyType]int64)
	var total int64

	err := m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		total = int64(b.Stats().KeyN)

		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			var p ClosedProxy
			if err := json.Unmarshal(v, &p); err == nil {
				stats[p.Type]++
			}
		}
		return nil
	})

	return stats, total, err
}

// Close 关闭数据库连接
func (m *ClosedProxyManager) Close() error {
	return m.db.Close()
}

// DeleteClosed 根据 runID 删除记录
func (m *ClosedProxyManager) DeleteClosed(runID string) error {
	return m.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		// 如果 key 不存在，DeleteClosed 也是返回 nil (成功)，无需额外判断
		return b.Delete([]byte(runID))
	})
}

// DeleteClosedByType 根据类型删除记录
func (m *ClosedProxyManager) DeleteClosedByType(closeType ClosedProxyType) (int64, error) {
	var count int64
	err := m.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		c := b.Cursor()

		var keysToDelete [][]byte

		for k, v := c.First(); k != nil; k, v = c.Next() {
			var p ClosedProxy
			if err := json.Unmarshal(v, &p); err == nil {
				if p.Type == closeType {
					keysToDelete = append(keysToDelete, k)
				}
			}
		}

		for _, key := range keysToDelete {
			if err := b.Delete(key); err == nil {
				count++
			}
		}
		return nil
	})
	return count, err
}

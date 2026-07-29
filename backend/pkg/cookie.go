package pkg

import (
	"sync"

	"gorm.io/gorm"
)

// CookieManager manages per-source cookie strings, backed by SQLite.
type CookieManager struct {
	mu      sync.RWMutex
	cookies map[string]string
}

// CM is the global cookie manager singleton.
var CM = &CookieManager{cookies: make(map[string]string)}

// Load reads cookies from the SQLite config DB.
func (m *CookieManager) Load() {
	if err := ensureConfigDB(); err != nil {
		return
	}
	var rows []cookieEntry
	if err := configDB.Order("source ASC").Find(&rows).Error; err != nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cookies = make(map[string]string, len(rows))
	for _, row := range rows {
		m.cookies[row.Source] = row.Value
	}
}

// Save persists all cookies to SQLite.
func (m *CookieManager) Save() {
	if err := ensureConfigDB(); err != nil {
		return
	}
	m.mu.RLock()
	rows := make([]cookieEntry, 0, len(m.cookies))
	for source, value := range m.cookies {
		source = trimSpace(source)
		value = trimSpace(value)
		if source == "" || value == "" {
			continue
		}
		rows = append(rows, cookieEntry{Source: source, Value: value})
	}
	m.mu.RUnlock()

	_ = configDB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&cookieEntry{}).Error; err != nil {
			return err
		}
		if len(rows) == 0 {
			return nil
		}
		return tx.Create(&rows).Error
	})
}

// Get returns the cookie for a given source.
func (m *CookieManager) Get(source string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cookies[source]
}

// SetAll replaces all cookies.
func (m *CookieManager) SetAll(c map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k, v := range c {
		if v == "" {
			delete(m.cookies, k)
		} else {
			m.cookies[k] = v
		}
	}
}

// GetAll returns a copy of all cookies.
func (m *CookieManager) GetAll() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	res := make(map[string]string, len(m.cookies))
	for k, v := range m.cookies {
		res[k] = v
	}
	return res
}

type cookieEntry struct {
	Source    string `gorm:"primaryKey;size:64"`
	Value     string `gorm:"type:text;not null"`
	UpdatedAt int64  `gorm:"autoUpdateTime"`
}

package wingo

import (
	"github.com/jinzhu/gorm"
)

type SqlModel interface {
	init(db *gorm.DB, self interface{})
	DB() *gorm.DB
	Sync(cols ...interface{}) error
	Del() error
	SyncInTx(tx *gorm.DB, cols ...interface{}) error
	DelInTx(tx *gorm.DB) error
}

type BaseSqlModel struct {
	db   *gorm.DB
	self interface{}
	Id   uint32 `gorm:"primary_key"`
}

func (m *BaseSqlModel) init(db *gorm.DB, self interface{}) {
	m.db = db
	m.self = self
}

func (m *BaseSqlModel) DB() *gorm.DB {
	return m.db
}

// 更新到数据库
func (m *BaseSqlModel) Sync(cols ...interface{}) error {
	if cols != nil && len(cols) > 0 {
		return m.DB().Model(m.self).UpdateColumn(cols).Error
	} else {
		return m.DB().Model(m.self).Updates(m.self).Error
	}
}

// 从数据库移除，ID必须存在
func (m *BaseSqlModel) Del() error {
	return m.DB().Delete(m.self).Error
}

// 在事务里更新
func (m *BaseSqlModel) SyncInTx(tx *gorm.DB, cols ...interface{}) error {
	if cols != nil && len(cols) > 0 {
		return tx.Model(m.self).UpdateColumn(cols).Error
	} else {
		return tx.Model(m.self).Updates(m.self).Error
	}
}

// 在事务里移除，ID必须存在
func (m *BaseSqlModel) DelInTx(tx *gorm.DB) error {
	return tx.Delete(m.self).Error
}

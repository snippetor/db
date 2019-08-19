package wingo

import (
	"github.com/jinzhu/gorm"
)

type SqlDB interface {
	Dial(dbType SqlDBType, host, port, user, pwd, defaultDb, tbPrefix string, debugMode bool) error
	DialSqlite3(path, tbPrefix string, debugMode bool) error
	DB() *gorm.DB
	TableName(tbName string) string
	AutoMigrate(model SqlModel)

	Create(model SqlModel) error
	Find(model SqlModel) error
	FindAll(models interface{}) error
	FindMany(models interface{}, limit int, orderBy string, whereAndArgs ...interface{}) error
	Begin() *gorm.DB
	Rollback()
	Commit()
	Close() error
}

type SqlDBType byte

const (
	MySql SqlDBType = iota
	MsSql
	Postgres
)

func NewSqlDB() SqlDB {
	return &sqlDB{}
}

type sqlDB struct {
	tbPrefix string
	db       *gorm.DB
}

func (m *sqlDB) Dial(dbType SqlDBType, host, port, user, pwd, defaultDb, tbPrefix string, debugMode bool) error {
	m.tbPrefix = tbPrefix
	// db
	var db *gorm.DB
	var err error
	switch dbType {
	case MySql:
		db, err = gorm.Open("mysql", user+":"+pwd+"@tcp("+host+":"+port+")/"+defaultDb+"?charset=utf8&parseTime=True&loc=Local")
	case MsSql:
		db, err = gorm.Open("mssql", "sqlserver://"+user+":"+pwd+"@"+host+":"+port+"?database="+defaultDb)
	case Postgres:
		db, err = gorm.Open("postgres", "host="+host+" port="+port+" user="+user+" dbname="+defaultDb+" password="+pwd)
	}
	if err != nil {
		return err
	}
	m.db = db
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return m.tbPrefix + "_" + defaultTableName
	}
	db.LogMode(debugMode)
	return nil
}

// @path: file path or :memory:
func (m *sqlDB) DialSqlite3(path, tbPrefix string, debugMode bool) error {
	m.tbPrefix = tbPrefix
	// db
	db, err := gorm.Open("sqlite3", path)
	if err != nil {
		return err
	}
	m.db = db
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return m.tbPrefix + "_" + defaultTableName
	}
	db.LogMode(debugMode)
	return nil
}

func (m *sqlDB) DB() *gorm.DB {
	if m.db == nil {
		panic("- DB not be initialized, invoke Dial at first!")
	}
	return m.db
}

func (m *sqlDB) TableName(tbName string) string {
	if m.tbPrefix != "" {
		return m.tbPrefix + "_" + tbName
	}
	return tbName
}

func (m *sqlDB) AutoMigrate(model SqlModel) {
	model.init(m.db, model)
	m.DB().AutoMigrate(model)
}

func (m *sqlDB) Create(model SqlModel) error {
	model.init(m.db, model)
	res := m.DB().Create(model)
	return res.Error
}

func (m *sqlDB) Find(model SqlModel) error {
	model.init(m.db, model)
	res := m.DB().Where(model).First(model)
	return res.Error
}

func (m *sqlDB) FindAll(models interface{}) error {
	res := m.DB().Find(models)
	return res.Error
}

func (m *sqlDB) FindMany(models interface{}, limit int, orderBy string, whereAndArgs ...interface{}) error {
	db := m.DB()
	if limit > 0 {
		db = db.Limit(limit)
	}
	if orderBy != "" {
		db = db.Order(orderBy)
	}
	if len(whereAndArgs) > 0 && len(whereAndArgs)%2 == 0 {
		var args = make(map[string]interface{})
		for i := 0; i < len(whereAndArgs); i += 2 {
			args[whereAndArgs[i].(string)] = whereAndArgs[i+1]
		}
		db = db.Where(args)
	}
	db = db.Find(models)
	return db.Error
}

func (m *sqlDB) Begin() *gorm.DB {
	return m.DB().Begin()
}

func (m *sqlDB) Rollback() {
	m.DB().Rollback()
}

func (m *sqlDB) Commit() {
	m.DB().Rollback()
}

func (m *sqlDB) Close() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

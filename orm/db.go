package orm

import (
	"github.com/jinzhu/gorm"
)

type Database interface {
	DefaultDB() *gorm.DB
	TableName(tbName string) string
	AutoMigrate(model Model)

	Create(model Model) error
	Find(model Model) error
	FindAll(models interface{}) error
	FindMany(models interface{}, limit int, orderBy string, whereAndArgs ...interface{}) error
	ExecSql(string, ...interface{}) error
	ExecSqlWithResult(interface{}, string, ...interface{}) error
	Sync(Model, ...interface{}) error
	Delete(Model, ...interface{}) error

	Begin() Database
	Rollback()
	Commit()
	Close() error
}

func connect(dialect, tbPrefix string, debugMode bool, args ...interface{}) (Database, error) {
	// database
	db, err := gorm.Open(dialect, args...)
	if err != nil {
		return nil, err
	}
	gorm.DefaultTableNameHandler = func(db *gorm.DB, defaultTableName string) string {
		return tbPrefix + "_" + defaultTableName
	}
	db.LogMode(debugMode)
	return &database{tbPrefix, db}, nil
}

func ConnectMysql(host, port, user, pwd, defaultDb, tbPrefix string, debugMode bool) (Database, error) {
	return connect("mysql", tbPrefix, debugMode, user+":"+pwd+"@tcp("+host+":"+port+")/"+defaultDb+"?charset=utf8&parseTime=True&loc=Local")
}

func ConnectMssql(host, port, user, pwd, defaultDb, tbPrefix string, debugMode bool) (Database, error) {
	return connect("mssql", tbPrefix, debugMode, "sqlserver://"+user+":"+pwd+"@"+host+":"+port+"?database="+defaultDb)
}

func ConnectPostgres(host, port, user, pwd, defaultDb, tbPrefix string, debugMode bool) (Database, error) {
	return connect("postgres", tbPrefix, debugMode, "host="+host+" port="+port+" user="+user+" dbname="+defaultDb+" password="+pwd)
}

// @path: file path or :memory:
func ConnectSqlite(path, tbPrefix string, debugMode bool) (Database, error) {
	return connect("sqlite3", tbPrefix, debugMode, path)
}

type database struct {
	tbPrefix string
	db       *gorm.DB
}

func (d *database) DefaultDB() *gorm.DB {
	return d.db
}

func (d *database) TableName(tbName string) string {
	if d.tbPrefix != "" {
		return d.tbPrefix + "_" + tbName
	}
	return tbName
}

func (d *database) AutoMigrate(model Model) {
	d.db.AutoMigrate(model)
}

func (d *database) Create(model Model) error {
	res := d.db.Create(model)
	return res.Error
}

func (d *database) Find(model Model) error {
	res := d.db.Where(model).First(model)
	return res.Error
}

func (d *database) FindAll(models interface{}) error {
	res := d.db.Find(models)
	return res.Error
}

func (d *database) FindMany(models interface{}, limit int, orderBy string, whereAndArgs ...interface{}) error {
	db := d.db
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

// execute sql and no result return
func (d *database) ExecSql(sql string, values ...interface{}) error {
	res := d.db.Exec(sql, values...)
	return res.Error
}

// execute sql and no result return
func (d *database) ExecSqlWithResult(result interface{}, sql string, values ...interface{}) error {
	res := d.db.Raw(sql, values...).Scan(result)
	return res.Error
}

func (d *database) Sync(model Model, cols ...interface{}) error {
	if cols != nil && len(cols) > 0 {
		return d.db.Model(model).UpdateColumn(cols...).Error
	} else {
		return d.db.Model(model).Updates(model).Error
	}
}

// id must set if just delete single model
func (d *database) Delete(model Model, where ...interface{}) error {
	return d.db.Delete(model, where...).Error
}

func (d *database) Begin() Database {
	return &database{d.tbPrefix, d.db.Begin()}
}

func (d *database) Rollback() {
	d.db.Rollback()
}

func (d *database) Commit() {
	d.db.Rollback()
}

func (d *database) Close() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

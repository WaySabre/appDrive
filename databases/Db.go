package Db

import (
	"app/config"
	"github.com/WaySabre/appDrive/logging"
	"sync"
	"time"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
)

var Base *gorm.DB

type SqlData struct {
	Bool  bool
	Empty bool
}

type MysqlConnectiPool struct {
	Db *gorm.DB
}

var instance *MysqlConnectiPool
var once sync.Once

var err_db error

func Pool() *MysqlConnectiPool {
	once.Do(func() {
		instance = &MysqlConnectiPool{}
	})
	return instance
}

func (m *MysqlConnectiPool) Load() error {
	conf := config.GetConfAll()
	var err error
	Base, err = gorm.Open("mysql", conf.DbUserName+":"+conf.DbPassWord+"@("+conf.DbHost+")/"+conf.DbDataBase+"?charset=utf8mb4&parseTime=True&loc=Local")
	Base.LogMode(false)
	if err != nil {
		return err
	}
	Base.DB().SetMaxIdleConns(10)
	Base.DB().SetMaxOpenConns(100)
	Base.DB().SetConnMaxLifetime(time.Hour)
	m.Db = Base
	return nil
}

func (m *MysqlConnectiPool) Do() *gorm.DB {
	return Base
}

//use first return record not found,use find return nil
func Handel(err *gorm.DB) SqlData {
	var parent SqlData
	parent.Bool = true
	parent.Empty = false
	if err.RecordNotFound() {
		parent.Empty = true
		return parent
	}
	if err.Error != nil {
		logging.Error("sql", err.Error)
		parent.Bool = false
	}
	return parent
}

package model

import (
	"carrota-plugin-homework/logs"
	"strconv"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	Hostname string `config:"hostname"`
	Port     int    `config:"port"`
	User     string `config:"user"`
	Password string `config:"password"`
	DbName   string `config:"dbName"`
	SslMode  bool   `config:"sslMode"`
	TimeZone string `config:"timeZone"`
}

func Connect(d Database) error {
	isSSL := "disable"
	if d.SslMode {
		isSSL = "enable"
	}
	dsn := "host=" + d.Hostname +
		" user=" + d.User +
		" password=" + d.Password +
		" dbname=" + d.DbName +
		" port=" + strconv.Itoa(d.Port) +
		" sslmode=" + isSSL +
		" TimeZone=" + d.TimeZone
	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logs.Logs.Error("Failed to connect database "+d.DbName+".", zap.Error(err))
	}
	return err
}

// Almost used for creating tables
func AutoMigrateTable(dst ...interface{}) error {
	for _, d := range dst {
		err := db.Migrator().AutoMigrate(d)
		if err != nil {
			logs.Logs.Error("Auto migrate table failed.", zap.Any("model", d), zap.Error(err))
			return err
		}
		logs.Logs.Debug("Auto migrate table successfully. ", zap.Any("model", d))
	}
	return nil
}

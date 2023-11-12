package model

import (
	"carrota-plugin-center/utils/logs"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lib/pq"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PluginParam struct {
	Key         string `json:"key"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

type PluginParamArray []PluginParam

func (p *PluginParamArray) Scan(value interface{}) error {
	switch v := value.(type) {
	case string:
		return json.Unmarshal([]byte(v), p)
	case []byte:
		return json.Unmarshal(v, p)
	default:
		return fmt.Errorf("unsupported type: %T", value)
	}
}

func (p PluginParamArray) Value() (driver.Value, error) {
	if len(p) == 0 {
		return nil, nil
	}
	return json.Marshal(p)
}

type Plugin struct {
	ID          string           `json:"id"          form:"id"          query:"id"          gorm:"primaryKey;unique;not null"`
	CreatedAt   time.Time        `json:"created_at"  form:"created_at"  query:"created_at" `
	UpdatedAt   time.Time        `json:"updated_at"  form:"updated_at"  query:"updated_at" `
	DeletedAt   gorm.DeletedAt   `json:"deleted_at"  form:"deleted_at"  query:"deleted_at" `
	Name        string           `json:"name"        form:"name"        query:"name"        gorm:"not null"`
	Author      string           `json:"author"      form:"author"      query:"author"      gorm:"not null"`
	Description string           `json:"description" form:"description" query:"description" gorm:"not null"`
	Prompt      string           `json:"prompt"      form:"prompt"      query:"prompt"      gorm:"not null"`
	Params      PluginParamArray `json:"param"       form:"param"       query:"param"       gorm:"type:jsonb"`
	Format      pq.StringArray   `json:"format"      form:"format"      query:"format"      gorm:"type:text[]"`
	Example     pq.StringArray   `json:"example"     form:"example"     query:"example"     gorm:"type:text[]"`
	Url         string           `json:"url"         form:"url"         query:"url"         gorm:"not null"`
}

type PluginInfo struct {
	ID          string           `json:"id"          `
	Name        string           `json:"name"        `
	Author      string           `json:"author"      `
	Description string           `json:"description" `
	Prompt      string           `json:"prompt"      `
	Params      PluginParamArray `json:"param"       `
	Format      []string         `json:"format"      `
	Example     []string         `json:"example"     `
	Url         string           `json:"url"         `
}

func CreatePluginRegisterRecord(plugin PluginInfo) error {
	m := GetModel()
	defer m.Close()

	record := Plugin{
		ID:          plugin.ID,
		Name:        plugin.Name,
		Author:      plugin.Author,
		Description: plugin.Description,
		Prompt:      plugin.Prompt,
		Params:      plugin.Params,
		Format:      pq.StringArray(plugin.Format),
		Example:     pq.StringArray(plugin.Example),
		Url:         plugin.Url,
	}
	result := m.tx.Clauses(clause.OnConflict{UpdateAll: true}).Create(&record)
	if result.Error != nil {
		logs.Warn("Create PluginRegisterRecord failed.", zap.Error(result.Error))
		m.Abort()
		return result.Error
	}

	m.tx.Commit()
	return nil
}

func FindPluginList() ([]PluginInfo, error) {
	m := GetModel()
	defer m.Close()

	var plugins []Plugin
	result := m.tx.Model(&Plugin{}).Find(&plugins)
	if result.Error != nil {
		logs.Info("Find plugin list failed.", zap.Error(result.Error))
		m.Abort()
		return nil, result.Error
	}

	m.tx.Commit()
	var pluginInfos []PluginInfo
	for _, plugin := range plugins {
		pluginInfos = append(pluginInfos, PluginInfo{
			ID:          plugin.ID,
			Name:        plugin.Name,
			Author:      plugin.Author,
			Description: plugin.Description,
			Prompt:      plugin.Prompt,
			Params:      plugin.Params,
			Format:      plugin.Format,
			Example:     plugin.Example,
			Url:         plugin.Url,
		})
	}
	return pluginInfos, nil
}

func FindPluginById(id string) (PluginInfo, error) {
	m := GetModel()
	defer m.Close()

	var plugin PluginInfo
	result := m.tx.Model(&Plugin{}).Where("id = ?", id).First(&plugin)
	if result.Error != nil {
		logs.Info("Find plugin by id failed.", zap.Error(result.Error))
		m.Abort()
		return plugin, result.Error
	}

	m.tx.Commit()
	return plugin, nil
}

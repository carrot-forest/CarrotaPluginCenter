package model

import (
	"carrota-plugin-homework/logs"

	"go.uber.org/zap"
)

type Homework struct {
	Subject string `json:"subject" gorm:"column:subject"`
	Content string `json:"content" gorm:"column:content"`
}

func CreateHomeworkRecord(homework Homework) error {
	m := GetModel()
	defer m.Close()

	result := m.tx.Create(&homework)
	if result.Error != nil {
		logs.Logs.Warn("Create HomeworkRecord failed.", zap.Error(result.Error))
		m.Abort()
		return result.Error
	}

	m.tx.Commit()
	return nil
}

func FindHomeworkBySubject(subject string) ([]Homework, error) {
	m := GetModel()
	defer m.Close()

	var homework []Homework
	result := m.tx.Model(&Homework{}).Where("subject = ?", subject).Find(&homework)
	if result.Error != nil {
		logs.Logs.Info("Find homework by subject failed.", zap.Error(result.Error))
		m.Abort()
		return []Homework{}, result.Error
	}

	m.tx.Commit()
	return homework, nil
}

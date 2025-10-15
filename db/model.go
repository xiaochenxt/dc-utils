package db

import (
	"github.com/dc-utils/ddatetime"
	"github.com/dc-utils/snowflake"
	"gorm.io/gorm"
	"gorm.io/plugin/optimisticlock"
	"strconv"
)

type BaseModel struct {
	Id      *string                 `json:"id" gorm:"primaryKey"`
	Version *optimisticlock.Version `json:"version"`

	CreatedTime ddatetime.LocalDateTime `json:"createdTime"`

	ModifiedTime ddatetime.LocalDateTime `json:"modifiedTime"`

	CreatedBy string `json:"createdBy"`

	ModifiedBy string `json:"modifiedBy"`
}

func (*BaseModel) TableId() string {
	return "id"
}

func (m *BaseModel) NewId() *string {
	id := strconv.FormatInt(snowflake.NextId(), 10)
	return &id
}

func (m *BaseModel) BeforeSave(tx *gorm.DB) (err error) {
	if m.Id == nil {
		m.Id = m.NewId()
	}
	return
}

func (m *BaseModel) BeforeCreate(tx *gorm.DB) (err error) {
	currentUser, exists := GetCurrentUserFromContext(tx.Statement.Context)
	if exists {
		m.CreatedBy = *currentUser.UserId
	}
	m.CreatedTime = ddatetime.Now()
	return
}

func (m *BaseModel) BeforeUpdate(tx *gorm.DB) (err error) {
	currentUser, exists := GetCurrentUserFromContext(tx.Statement.Context)
	if exists {
		m.ModifiedBy = *currentUser.UserId
	}
	m.ModifiedTime = ddatetime.Now()
	return
}

package repository

import (
	"errors"
	"github.com/dc-utils/args"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/optimisticlock"
	"reflect"
)

// Repository 接口定义
type Repository[T any] interface {
	FindById(id any) (*T, error)
	FindAll() ([]T, error)
	FindByCond(cond any, args ...any) (*T, error)
	FindFirstByCond(cond any, args ...any) (*T, error)
	FindInIds(id ...any) ([]T, error)
	FindByIdForUpdate(id any) (*T, error)
	FindByIdForShare(id any) (*T, error)
	Create(entity *T) error
	Update(entity *T) error
	Save(entity *T) error
	Delete(id ...any) error
	Count(cond any, args ...any) (int64, error)
	Exists(cond any, args ...any) (bool, error)
	ExistsById(id any) (bool, error)
}

var debug bool

func init() {
	debug = args.GetBool("db.debug", false)
}

// GormRepository 实现接口
type GormRepository[T any] struct {
	db    *gorm.DB
	getDB func() *gorm.DB
}

var (
	versionValueType   = reflect.TypeOf(optimisticlock.Version{})
	versionPointerType = reflect.TypeOf(&optimisticlock.Version{})
)

// NewRepository 工厂函数
func NewRepository[T any](db *gorm.DB) Repository[T] {
	var getDB func() *gorm.DB
	if debug {
		getDB = func() *gorm.DB {
			return db.Debug()
		}
	} else {
		getDB = func() *gorm.DB {
			return db
		}
	}
	return &GormRepository[T]{db: db, getDB: getDB}
}

func (r *GormRepository[T]) FindById(id any) (*T, error) {
	var entity T

	err := r.getDB().First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T]) FindAll() ([]T, error) {
	var entities []T
	err := r.getDB().Find(&entities).Error
	return entities, err
}

func (r *GormRepository[T]) FindByCond(cond any, args ...any) (*T, error) {
	var entity T
	err := r.getDB().Where(cond, args).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T]) FindFirstByCond(cond any, args ...any) (*T, error) {
	var entity T
	err := r.getDB().Where(cond, args).Order("id desc").Limit(1).Find(&entity).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T]) FindInIds(id ...any) ([]T, error) {
	var entities []T
	err := r.getDB().Find(&entities, id).Error
	return entities, err
}

func (r *GormRepository[T]) FindByIdForUpdate(id any) (*T, error) {
	var entity T
	err := r.getDB().Clauses(clause.Locking{Strength: "UPDATE"}).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T]) FindByIdForShare(id any) (*T, error) {
	var entity T
	err := r.getDB().Clauses(clause.Locking{Strength: "SHARE"}).First(&entity, id).Error
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (r *GormRepository[T]) Create(entity *T) error {
	return r.getDB().Create(entity).Error
}

func (r *GormRepository[T]) CreateBatch(entity []*T, batchSize int) error {
	return r.getDB().CreateInBatches(entity, batchSize).Error
}

func (r *GormRepository[T]) Update(entity *T) error {
	return r.getDB().Updates(entity).Error
}

func (r *GormRepository[T]) Save(entity *T) error {
	tx := r.getDB()
	reflectValue := reflect.Indirect(reflect.ValueOf(entity))
	if err := tx.Statement.Parse(entity); err == nil && tx.Statement.Schema != nil {
		isCreate := false
		version := tx.Statement.Schema.FieldsByName["Version"]
		if version != nil && (version.FieldType == versionPointerType || version.FieldType == versionValueType) {
			if _, isZero := version.ValueOf(tx.Statement.Context, reflectValue); isZero {
				isCreate = true
			}
		}
		for _, pf := range tx.Statement.Schema.PrimaryFields {
			if _, isZero := pf.ValueOf(tx.Statement.Context, reflectValue); isZero {
				isCreate = true
			}
		}
		if isCreate {
			return tx.Create(entity).Error
		}
	}
	return tx.Updates(entity).Error
}

func (r *GormRepository[T]) Delete(id ...any) error {
	var entity T
	return r.getDB().Delete(&entity, id).Error
}

func (r *GormRepository[T]) Count(cond any, args ...any) (int64, error) {
	var count int64
	err := r.getDB().Model(new(T)).Where(cond, args...).Count(&count).Error
	return count, err
}

// Exists 检查记录是否存在（优化版本）
func (r *GormRepository[T]) Exists(cond any, args ...any) (bool, error) {
	var exists bool
	query := r.getDB().Model(new(T)).Select("1").Where(cond, args...).Limit(1)

	// 使用 Row().Scan() 直接查询是否存在记录
	err := query.Row().Scan(&exists)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *GormRepository[T]) ExistsById(id any) (bool, error) {
	var exists bool
	query := r.getDB().Model(new(T)).Select("1").Where("id = ?", id).Limit(1)
	err := query.Row().Scan(&exists)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

package repository

import (
	"github.com/dc-utils/db"
	"gorm.io/gorm"
	"reflect"
	"sync"
)

func init() {
	InitManager(db.Get())
}

// Manager 管理所有类型的单例仓储
type Manager struct {
	db           *gorm.DB
	repositories sync.Map // 存储单例仓储实例
}

// 全局单例管理器
var manager = &Manager{}

// InitManager 初始化管理器（应用启动时调用一次）
func InitManager(db *gorm.DB) {
	manager.db = db
}

// GetRepository 获取指定类型的单例仓储
func GetRepository[T any]() Repository[T] {
	// 使用类型作为键
	key := typeKey[T]()
	// 尝试获取已存在的实例
	if repo, loaded := manager.repositories.Load(key); loaded {
		return repo.(Repository[T])
	}
	// 创建新实例并存储（利用sync.Map的原子操作）
	repo := NewRepository[T](manager.db)
	if existing, loaded := manager.repositories.LoadOrStore(key, repo); loaded {
		return existing.(Repository[T]) // 其他goroutine已创建，使用现有实例
	}
	return repo // 使用当前创建的实例
}

// typeKey 生成类型的唯一键
func typeKey[T any]() string {
	var t T
	return reflect.TypeOf(&t).Elem().String()
}

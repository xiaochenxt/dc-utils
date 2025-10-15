package db

import (
	"context"
	"github.com/dc-utils/args"
	"github.com/dc-utils/datasize"
	"github.com/dc-utils/ddatetime"
	"github.com/dc-utils/mem"
	"github.com/dc-utils/token"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
	"time"
)

var (
	db   *gorm.DB
	once sync.Once
)

func init() {
	enabled := args.GetBool("db.enabled", true)
	if !enabled {
		return
	}
	Enable()
}

func Enable() {
	memory, _ := mem.GetSystemAvailableMemory()
	var prepareStmt = false
	if memory > datasize.GB*8 {
		prepareStmt = true
	}
	EnableWithConfig(&gorm.Config{
		SkipDefaultTransaction:                   args.GetBool("db.skipDefaultTransaction", true),
		DefaultTransactionTimeout:                args.GetDuration("db.defaultTransactionTimeout", 0),
		PrepareStmt:                              args.GetBool("db.prepareStmt", prepareStmt),
		PrepareStmtMaxSize:                       args.GetInt("db.prepareStmtMaxSize", 0),
		PrepareStmtTTL:                           args.GetDuration("db.prepareStmtTTL", 0),
		CreateBatchSize:                          args.GetInt("db.createBatchSize", 0),
		DryRun:                                   args.GetBool("db.dryRun", false),
		DisableAutomaticPing:                     args.GetBool("db.disableAutomaticPing", false),
		DisableForeignKeyConstraintWhenMigrating: args.GetBool("db.disableForeignKeyConstraintWhenMigrating", false),
		IgnoreRelationshipsWhenMigrating:         args.GetBool("db.ignoreRelationshipsWhenMigrating", false),
		DisableNestedTransaction:                 args.GetBool("db.disableNestedTransaction", false),
		AllowGlobalUpdate:                        args.GetBool("db.allowGlobalUpdate", false),
		QueryFields:                              args.GetBool("db.queryFields", false),
		TranslateError:                           args.GetBool("db.translateError", false),
		PropagateUnscoped:                        args.GetBool("db.propagateUnscoped", false),
	})
}

func EnableWithConfig(config *gorm.Config) {
	dsn := args.Get("db.dsn")
	if dsn == "" {
		return
	}
	once.Do(func() {
		_db, err := gorm.Open(mysql.Open(dsn), config)
		if err != nil {
			log.Fatalf("failed to connect database: %v", err)
			return
		}
		// 配置连接池（只配置一次）
		sqlDB, err := _db.DB()
		if err != nil {
			log.Fatalf("failed to get SQL DB: %v", err)
			return
		}
		sqlDB.SetMaxIdleConns(args.GetInt("db.maxIdleConns", 20))                        // 最大空闲连接数
		sqlDB.SetMaxOpenConns(args.GetInt("db.maxOpenConns", 200))                       // 最大打开连接数
		sqlDB.SetConnMaxLifetime(args.GetDuration("dc.connMaxLifetime", 30*time.Minute)) // 连接最大存活时间
		sqlDB.SetConnMaxIdleTime(args.GetDuration("dc.connMaxIdleTime", 10*time.Minute)) // 连接最大空闲时间
		if args.GetBool("db.audit.enabled", false) {
			audit(_db)
		}
		db = _db
	})
}

// Get 返回单例数据库连接
func Get() *gorm.DB {
	return db
}

func GetNoTransaction() *gorm.DB {
	return db.Session(&gorm.Session{SkipDefaultTransaction: true})
}

func audit(db *gorm.DB) {
	// 创建前回调
	_ = db.Callback().Create().Before("gorm:create").Register("set_created_by", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil {
			currentUser, exists := GetCurrentUserFromContext(tx.Statement.Context)
			if exists {
				if field := tx.Statement.Schema.LookUpField("CreatedBy"); field != nil {
					_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, currentUser.UserId)
				}
				if field := tx.Statement.Schema.LookUpField("ModifiedBy"); field != nil {
					_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, currentUser.UserId)
				}
			}
			now := ddatetime.Now()
			if field := tx.Statement.Schema.LookUpField("CreatedTime"); field != nil {
				_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, now)
			}
			if field := tx.Statement.Schema.LookUpField("ModifiedTime"); field != nil {
				_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, now)
			}
		}
	})
	_ = db.Callback().Update().Before("gorm:update").Register("set_updated_by", func(tx *gorm.DB) {
		if tx.Statement.Schema != nil {
			currentUser, exists := GetCurrentUserFromContext(tx.Statement.Context)
			if exists {
				if field := tx.Statement.Schema.LookUpField("ModifiedBy"); field != nil {
					_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, currentUser.UserId)
				}
			}
			if field := tx.Statement.Schema.LookUpField("ModifiedTime"); field != nil {
				now := ddatetime.Now()
				_ = field.Set(tx.Statement.Context, tx.Statement.ReflectValue, now)
			}
		}
	})
}

func GetCurrentUserFromContext(ctx context.Context) (token.Authentication, bool) {
	user, ok := ctx.Value("userInfo").(token.Authentication)
	return user, ok
}

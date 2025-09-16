package database

import (
	"context"
	"fmt"

	"github.com/openzipkin/zipkin-go"
	"gorm.io/gorm"
)

const zipkinGormSpanKey = "zipkin:span"

// ZipkinPlugin Zipkin追踪插件
type ZipkinPlugin struct {
	tracer *zipkin.Tracer
}

// Name 插件名称
func (z *ZipkinPlugin) Name() string {
	return "zipkin"
}

// Initialize 初始化插件
func (z *ZipkinPlugin) Initialize(db *gorm.DB) error {
	// 注册回调
	db.Callback().Create().Before("gorm:create").Register("zipkin:before_create", z.before)
	db.Callback().Create().After("gorm:create").Register("zipkin:after_create", z.after)

	db.Callback().Query().Before("gorm:query").Register("zipkin:before_query", z.before)
	db.Callback().Query().After("gorm:query").Register("zipkin:after_query", z.after)

	db.Callback().Update().Before("gorm:update").Register("zipkin:before_update", z.before)
	db.Callback().Update().After("gorm:update").Register("zipkin:after_update", z.after)

	db.Callback().Delete().Before("gorm:delete").Register("zipkin:before_delete", z.before)
	db.Callback().Delete().After("gorm:delete").Register("zipkin:after_delete", z.after)

	db.Callback().Row().Before("gorm:row").Register("zipkin:before_row", z.before)
	db.Callback().Row().After("gorm:row").Register("zipkin:after_row", z.after)

	db.Callback().Raw().Before("gorm:raw").Register("zipkin:before_raw", z.before)
	db.Callback().Raw().After("gorm:raw").Register("zipkin:after_raw", z.after)

	return nil
}

// before 在操作前的回调
func (z *ZipkinPlugin) before(db *gorm.DB) {
	if z.tracer == nil {
		return
	}

	// 从上下文获取parent span
	ctx := db.Statement.Context
	if ctx == nil {
		ctx = context.Background()
	}

	// 创建操作名称
	operationName := fmt.Sprintf("db:%s", getOperationType(db))
	if db.Statement.Table != "" {
		operationName += ":" + db.Statement.Table
	}

	// 创建span
	span := z.tracer.StartSpan(operationName, zipkin.Parent(zipkin.SpanFromContext(ctx).Context()))

	// 设置标签
	span.Tag("db.type", "postgresql")
	span.Tag("db.operation", getOperationType(db))
	if db.Statement.Table != "" {
		span.Tag("db.table", db.Statement.Table)
	}

	// 将span存储到context中
	newCtx := zipkin.NewContext(ctx, span)
	db.Statement.Context = newCtx

	// 将span存储到实例变量中，以便在after回调中使用
	db.InstanceSet(zipkinGormSpanKey, span)
}

// after 在操作后的回调
func (z *ZipkinPlugin) after(db *gorm.DB) {
	// 获取span
	val, exists := db.InstanceGet(zipkinGormSpanKey)
	if !exists {
		return
	}

	span, ok := val.(zipkin.Span)
	if !ok {
		return
	}

	// 设置SQL语句（如果可用）
	if db.Statement.SQL.String() != "" {
		span.Tag("db.statement", db.Statement.SQL.String())
	}

	// 记录受影响的行数
	if db.Statement.RowsAffected >= 0 {
		span.Tag("db.rows_affected", fmt.Sprintf("%d", db.Statement.RowsAffected))
	}

	// 如果有错误，记录错误信息
	if db.Error != nil {
		span.Tag("error", "true")
		span.Tag("error.message", db.Error.Error())
	}

	// 完成span
	span.Finish()
}

// getOperationType 获取操作类型
func getOperationType(db *gorm.DB) string {
	switch {
	case db.Statement.SQL.String() != "":
		return "raw"
	default:
		// 根据不同的SQL操作返回不同的类型
		sql := db.Statement.SQL.String()
		if sql == "" && db.Statement.Dest != nil {
			return "query"
		}
		return "unknown"
	}
}

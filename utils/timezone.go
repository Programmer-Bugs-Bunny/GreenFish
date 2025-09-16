package utils

import (
	"fmt"
	"time"
)

var localTimezone *time.Location

// InitTimezone 初始化时区设置
func InitTimezone(timezone string) error {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return fmt.Errorf("加载时区失败 %s: %w", timezone, err)
	}

	localTimezone = loc
	return nil
}

// GetCurrentTime 获取当前时间（使用设置的时区）
func GetCurrentTime() time.Time {
	if localTimezone != nil {
		return time.Now().In(localTimezone)
	}
	return time.Now()
}

// GetCurrentTimeString 获取当前时间字符串
func GetCurrentTimeString() string {
	return GetCurrentTime().Format("2006-01-02 15:04:05")
}

// GetCurrentDate 获取当前日期字符串
func GetCurrentDate() string {
	return GetCurrentTime().Format("2006-01-02")
}

// ParseTime 解析时间字符串到本地时区
func ParseTime(layout, value string) (time.Time, error) {
	t, err := time.Parse(layout, value)
	if err != nil {
		return time.Time{}, err
	}

	if localTimezone != nil {
		return t.In(localTimezone), nil
	}
	return t, nil
}

// FormatTime 将时间格式化为字符串（使用本地时区）
func FormatTime(t time.Time, layout string) string {
	if localTimezone != nil {
		return t.In(localTimezone).Format(layout)
	}
	return t.Format(layout)
}

// GetTimezone 获取当前时区名称
func GetTimezone() string {
	if localTimezone != nil {
		return localTimezone.String()
	}
	return time.Local.String()
}

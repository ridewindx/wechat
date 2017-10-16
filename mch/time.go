package mch

import "time"

var BeijingLocation = time.FixedZone("Asia/Shanghai", 8*60*60)

// FormatTime 将参数 t 格式化成北京时间 yyyyMMddHHmmss 字符串.
func FormatTime(t time.Time) string {
	return t.In(BeijingLocation).Format("20060102150405")
}

// ParseTime 将北京时间 yyyyMMddHHmmss 字符串解析到 time.Time.
func ParseTime(value string) (time.Time, error) {
	return time.ParseInLocation("20060102150405", value, BeijingLocation)
}

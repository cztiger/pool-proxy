package util

import(
	"time"
)

//解析时间字符串
func MustParseDuration(s string) time.Duration {
	value, err := time.ParseDuration(s)
	if err != nil {
		panic("解析时间单位字符串 '" + s + "' 失败,详情: " + err.Error())
	}
	return value
}
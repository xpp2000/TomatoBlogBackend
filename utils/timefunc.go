package utils

import (
	"errors"
	"time"
)

func PharseTimeString(timeStr string) (*time.Time, error) {
	var FinalPublishedAt *time.Time
	if timeStr != "" {
		// 尝试按 RFC3339 解析
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			FinalPublishedAt = &t
		} else if t2, err2 := time.Parse(time.DateTime, timeStr); err2 == nil {
			// 2. 尝试按 "YYYY-MM-DD HH:MM:SS" 解析
			FinalPublishedAt = &t2
		} else if t3, err3 := time.Parse(time.DateOnly, timeStr); err3 == nil {
			// 3. 尝试按 "YYYY-MM-DD" 解析 (Go 1.20+)
			// 解析结果的默认时分秒会是 00:00:00
			FinalPublishedAt = &t3
		} else {
			return nil, errors.New("fail to phrase time string")
		}
	}
	return FinalPublishedAt, nil
}

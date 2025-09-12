package model

import (
	"time"
)

// 设备监控数据结构
type DeviceMonitor struct {
	ID                    int       `json:"id"`
	Name                  string    `json:"name"`
	Code                  string    `json:"code"`
	TotalPoints           int       `json:"totalPoints"`
	CurrentAlarmCount     int       `json:"currentAlarmCount"`
	Status                string    `json:"status"`
	LastCommunicationTime time.Time `json:"lastCommunicationTime"`
}

// 告警详情数据结构
type AlarmDetail struct {
	Point        string `json:"point"`
	Description  string `json:"description"`
	CurrentValue string `json:"currentValue"`
	Level        string `json:"level"`
	Condition    string `json:"condition"`
}

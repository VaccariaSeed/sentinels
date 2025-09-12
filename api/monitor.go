package api

import (
	"fmt"
	"net/http"
	"sentinels/global"
	"sentinels/model"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func flushMonitorHandler(router *gin.Engine) {
	router.POST("/api/system/monitor", monitorHandler)
	router.GET("/api/system/monitor", alarmsHandler)
}

func monitorHandler(context *gin.Context) {
	var ms []*model.DeviceMonitor
	for index := 0; index < 10; index++ {
		dm := model.DeviceMonitor{
			ID:                    index,
			Name:                  fmt.Sprintf("test%d", index),
			Code:                  fmt.Sprintf("test%d", index),
			TotalPoints:           239,
			CurrentAlarmCount:     index,
			Status:                "在线",
			LastCommunicationTime: time.Now(),
		}
		ms = append(ms, &dm)
	}
	context.JSON(http.StatusOK, ms)
}

func alarmsHandler(context *gin.Context) {
	id, _ := strconv.Atoi(context.DefaultQuery("id", ""))
	global.SystemLog.Debug("alarm id: " + strconv.Itoa(id))
	var result []*model.AlarmDetail
	for index := 0; index < 10; index++ {
		alarm := model.AlarmDetail{
			Point:        "T33",
			Description:  "描述",
			CurrentValue: "24",
			Level:        "高",
			Condition:    "判定条件",
		}
		result = append(result, &alarm)
	}
	context.JSON(http.StatusOK, result)
}

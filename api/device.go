package api

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"net/http"
	"sentinels/global"
	"sentinels/model"
)

func flushDeviceHandler(router *gin.Engine) {
	router.GET("/api/devices/", selectDeviceHandler)
	router.GET("/api/devices/:id", oneDeviceHandler)
	router.DELETE("/api/devices/:id", deleteDeviceHandler)
	router.POST("/api/devices", saveDeviceHandler)
	router.PUT("/api/devices/:id/status", statusDeviceHandler)
	router.PUT("/api/devices/:id", updateDeviceHandler)
}

// 查询所有设备
func selectDeviceHandler(context *gin.Context) {
	context.JSON(http.StatusOK, model.NewDevice())
}

// 查询指定的设备
func oneDeviceHandler(context *gin.Context) {
	deviceID := context.Param("id")
	global.SystemLog.Debugf("select device id: %s", deviceID)
	context.JSON(http.StatusOK, model.Devices[1])
}

// 删除设备 逻辑删除
func deleteDeviceHandler(context *gin.Context) {
	deviceID := context.Param("id")
	global.SystemLog.Debugf("delete device id: %s", deviceID)
	context.JSON(http.StatusOK, nil)
}

func saveDeviceHandler(context *gin.Context) {
	var device model.Device
	if err := context.ShouldBindJSON(&device); err != nil {
		// 绑定失败，返回错误信息（通常是 400 Bad Request）
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	data, _ := json.Marshal(device)
	global.SystemLog.Debugf("save device data: %s", string(data))
	context.JSON(http.StatusOK, nil)
}

// 更改设备切入切出，切出就是屏蔽设备不采集不控制了
func statusDeviceHandler(context *gin.Context) {
	deviceID := context.Param("id")
	// 使用map获取单个值
	var payload map[string]interface{}
	if err := context.ShouldBindJSON(&payload); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "无效的JSON数据"})
		return
	}
	status, exists := payload["status"]
	if !exists {
		context.JSON(http.StatusBadRequest, gin.H{"error": "缺少status字段"})
		return
	}
	global.SystemLog.Debugf("change device:%s status to %v", deviceID, status)
	context.JSON(http.StatusOK, nil)
	context.JSON(http.StatusOK, nil)
}

func updateDeviceHandler(context *gin.Context) {
	var device model.Device
	if err := context.ShouldBindJSON(&device); err != nil {
		// 绑定失败，返回错误信息（通常是 400 Bad Request）
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	data, _ := json.Marshal(device)
	global.SystemLog.Debugf("update device data: %s", string(data))
	context.JSON(http.StatusOK, model.NewDevice())
}

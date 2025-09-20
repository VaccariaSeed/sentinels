package api

import (
	"net/http"
	"sentinels/global"
	"sentinels/model"
	"sentinels/store"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

func flushPointHandler(router *gin.Engine) {
	router.GET("/api/points", selectPointsHandler)
	router.POST("/api/points", savePointsHandler)
	router.PUT("/api/points/:id", savePointsHandler)
	router.DELETE("/api/points/:id", deletePointHandler)
	router.GET("/api/points/:id", onePointHandler)
	router.GET("/api/collection-rules", collectRulesHandler)
	router.GET("/api/collection-rules/:id", oneRulesHandler)
	router.POST("/api/collection-rules", saveRuleHandler)
	router.PUT("/api/collection-rules/:id", saveRuleHandler)
	router.DELETE("/api/collection-rules/:id", deleteRuleHandler)
}

// 查询指定设备的点位
func selectPointsHandler(context *gin.Context) {
	page, _ := strconv.Atoi(context.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(context.DefaultQuery("pageSize", "20"))
	deviceId, _ := strconv.Atoi(context.DefaultQuery("deviceId", ""))
	deviceMark := context.DefaultQuery("deviceMark", "")
	totalCount, points, err := store.DbClient.SelectPoints(page, pageSize, deviceId, deviceMark)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	var totalPages int
	if len(points)%pageSize == 0 {
		totalPages = totalCount / pageSize
	} else {
		totalPages = totalCount/pageSize + 1
	}
	context.JSON(http.StatusOK, &model.PointResponse{
		TotalCount: totalCount,
		Page:       page,
		PageSize:   len(points),
		TotalPages: totalPages,
		Points:     points,
	})
}

func deletePointHandler(context *gin.Context) {
	deviceID := context.Param("id")
	err := store.DbClient.DeletePoint(deviceID)
	if err != nil {
		// 绑定失败，返回错误信息（通常是 400 Bad Request）
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, nil)
}

func onePointHandler(context *gin.Context) {
	deviceID := context.Param("id")
	point, err := store.DbClient.SelectPointById(deviceID)
	if err != nil {
		// 绑定失败，返回错误信息（通常是 400 Bad Request）
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, point)
}

func savePointsHandler(context *gin.Context) {
	var point model.Point
	if err := context.ShouldBindJSON(&point); err != nil {
		// 绑定失败，返回错误信息（通常是 400 Bad Request）
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(point.DeviceID) == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "deviceId is empty"})
		return
	}
	device, err := store.DbClient.SelectDeviceById(point.DeviceID)
	if err != nil || device == nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "device not exist"})
		return
	}
	if device.ProtocolType == global.ModbusRTU {
		_, err = strconv.ParseUint(point.Address, 0, 16)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err = strconv.ParseUint(point.FunctionCode, 0, 8)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	err = store.DbClient.SavePoint(&point)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, nil)
}

func collectRulesHandler(context *gin.Context) {
	deviceId := context.DefaultQuery("deviceId", "")
	collects, err := store.DbClient.SelectCollectByDeviceId(deviceId)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, collects)
}

func saveRuleHandler(context *gin.Context) {
	var collect model.Collect
	if err := context.ShouldBindJSON(&collect); err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if strings.TrimSpace(collect.DeviceId) == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "not found deviceId"})
		return
	}
	device, err := store.DbClient.SelectDeviceById(collect.DeviceId)
	if err != nil || device == nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": "device not exist"})
		return
	}
	if device.ProtocolType == global.ModbusRTU {
		_, err = strconv.ParseUint(collect.StartPoint, 0, 16)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		_, err = strconv.ParseUint(collect.EndPoint, 0, 16)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}
	err = store.DbClient.SaveCollect(&collect)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, nil)
}

func oneRulesHandler(context *gin.Context) {
	collectID := context.Param("id")
	collect, err := store.DbClient.SelectOneCollect(collectID)
	if err != nil {
		// 绑定失败，返回错误信息（通常是 400 Bad Request）
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, collect)
}

func deleteRuleHandler(context *gin.Context) {
	deviceID := context.Param("id")
	err := store.DbClient.DeleteCollect(deviceID)
	if err != nil {
		// 绑定失败，返回错误信息（通常是 400 Bad Request）
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, nil)
}

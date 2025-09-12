package api

import (
	"encoding/json"
	"net/http"
	"sentinels/global"
	"sentinels/model"
	"strconv"

	"github.com/gin-gonic/gin"
)

func flushPointHandler(router *gin.Engine) {
	router.GET("/api/points", selectPointsHandler)
	router.POST("/api/points", savePointsHandler)
	router.PUT("/api/points/:id", savePointsHandler)
	router.DELETE("/api/points/:id", deletePointHandler)
	router.GET("/api/points/:id", onePointHandler)
}

// 查询指定设备的点位
func selectPointsHandler(context *gin.Context) {
	page, _ := strconv.Atoi(context.DefaultQuery("page", ""))
	pageSize, _ := strconv.Atoi(context.DefaultQuery("pageSize", ""))
	deviceId, _ := strconv.Atoi(context.DefaultQuery("deviceId", ""))
	deviceMark := context.DefaultQuery("deviceMark", "")
	global.SystemLog.Debugf("point select page:%d, pageSize:%d, deviceId:%d, deviceMark:%s", page, pageSize, deviceId, deviceMark)
	context.JSON(http.StatusOK, &model.PointResponse{
		TotalCount: 1000,
		Page:       1,
		PageSize:   100,
		TotalPages: 100,
		Points:     model.Points,
	})
}

func deletePointHandler(context *gin.Context) {
	deviceID := context.Param("id")
	global.SystemLog.Debugf("delete point id: %s", deviceID)
	context.JSON(http.StatusOK, nil)
}

func onePointHandler(context *gin.Context) {
	deviceID := context.Param("id")
	global.SystemLog.Debugf("select one point id: %s", deviceID)
	context.JSON(http.StatusOK, model.Points[4])
}

// 需要判断id是否存在
func savePointsHandler(context *gin.Context) {
	var point model.Point
	if err := context.ShouldBindJSON(&point); err != nil {
		// 绑定失败，返回错误信息（通常是 400 Bad Request）
		context.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}
	data, _ := json.Marshal(point)
	global.SystemLog.Debugf("save point data: %s", string(data))
	context.JSON(http.StatusOK, nil)
}

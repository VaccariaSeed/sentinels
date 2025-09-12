package api

import (
	"net/http"
	"sentinels/model"
	"sentinels/store"
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

// 需要判断id是否存在
func savePointsHandler(context *gin.Context) {
	var point model.Point
	if err := context.ShouldBindJSON(&point); err != nil {
		// 绑定失败，返回错误信息（通常是 400 Bad Request）
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	err := store.DbClient.SavePoint(&point)
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	context.JSON(http.StatusOK, nil)
}

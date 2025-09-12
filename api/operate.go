package api

import (
	"net/http"
	"os"
	"path/filepath"
	"sentinels/global"

	"github.com/gin-gonic/gin"
)

func flushOperateHandler(router *gin.Engine) {
	router.POST("/api/system/pause", pauseHandler)
	router.POST("/api/system/flush", flushHandler)
	router.POST("/api/data/clear", clearHandler)
	router.POST("/api/config/import", importHandler)
	router.GET("/api/config/template", templateHandler)
}

func pauseHandler(context *gin.Context) {
	global.SystemLog.Debug("暂停数据采集和监控")
	context.JSON(http.StatusOK, nil)
}

func flushHandler(context *gin.Context) {
	global.SystemLog.Debug("刷新采集控制配置")
	context.JSON(http.StatusOK, nil)
}

func clearHandler(context *gin.Context) {
	global.SystemLog.Debug("清空所有历史数据")
	context.JSON(http.StatusOK, nil)
}

func importHandler(context *gin.Context) {
	global.SystemLog.Debug("导入已经配置好的配置文件")
	file, err := context.FormFile("file")
	if err != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "无法获取上传文件: " + err.Error(),
		})
		return
	}
	ext := filepath.Ext(file.Filename)
	if ext != ".xlsx" {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "不支持的文件类型: " + ext,
		})
		return
	}
	dst := "./uploads/" + file.Filename
	if err := context.SaveUploadedFile(file, dst); err != nil {
		context.JSON(http.StatusInternalServerError, gin.H{
			"error": "文件保存失败: " + err.Error(),
		})
		return
	}
	context.JSON(http.StatusOK, nil)
}

func templateHandler(context *gin.Context) {
	global.SystemLog.Debug("下载导入模板文件")
	if _, err := os.Stat(global.TemplatePath); os.IsNotExist(err) {
		context.JSON(http.StatusNotFound, gin.H{
			"error": "模板文件不存在",
		})
		return
	}
	// 设置响应头，告诉浏览器这是要下载的文件
	context.Header("Content-Description", "File Transfer")
	context.Header("Content-Disposition", "attachment; filename="+filepath.Base(global.TemplatePath))
	context.Header("Content-Type", "application/octet-stream")
	context.Header("Content-Transfer-Encoding", "binary")
	context.Header("Expires", "0")
	context.Header("Cache-Control", "must-revalidate")
	context.Header("Pragma", "public")

	// 发送文件
	context.File(global.TemplatePath)
}

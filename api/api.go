package api

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"sentinels/global"

	"github.com/gin-gonic/gin"
)

func init() {
	go func() {
		gin.SetMode(gin.ReleaseMode)
		gin.DefaultWriter = io.Discard
		router := gin.Default()
		// 设置文件上传大小限制 (默认是32MiB)
		router.MaxMultipartMemory = 8 << 20 // 8MiB
		router.Use(gin.Recovery())
		router.Static("/static", global.Config.Static)
		router.Static("/css", path.Join(global.Config.Static, "css"))
		router.Static("/webfonts", path.Join(global.Config.Static, "webfonts"))
		router.Static("/js", path.Join(global.Config.Static, "js"))
		router.LoadHTMLGlob(path.Join(global.Config.Static, "*.html"))
		router.GET("/", homeHandler)
		flushDeviceHandler(router)
		flushPointHandler(router)
		flushOperateHandler(router)
		flushMonitorHandler(router)
		global.SystemLog.Debugf("starting http server -> 127.0.0.1:%d", global.Config.Port)
		err := router.Run(fmt.Sprintf(":%d", global.Config.Port))
		if err != nil {
			global.SystemLog.Errorf("start http server err:%s", err.Error())
			os.Exit(1)
		}
	}()
}

func homeHandler(context *gin.Context) {
	context.HTML(http.StatusOK, "home.html", gin.H{
		"title": "哨兵",
	})
}

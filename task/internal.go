package task

import (
	"encoding/json"
	"fmt"
	"sentinels/model"
	"sentinels/snap"
)

// 设备成功建立连接
func devConnected(dev *model.Device) {
	fmt.Println("dev connected:", dev.Identifier())
}

// 设备断开连接
func devDisConnected(dev *model.Device, err error) {
	fmt.Println("dev disconnected:", dev.Identifier())
}

// 读取到的数据
func devSwap(dev *model.Device, data map[string]interface{}, ts int64) {
	resp, _ := json.Marshal(data)
	fmt.Println("读取到数据：" + string(resp))
}

// 采集数据报错
func collectPointsFail(dev *model.Device, point snap.PointSnap, err error) {
	fmt.Println("collectPointsFail:", dev, point, err)
}

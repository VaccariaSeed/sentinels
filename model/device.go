package model

import "fmt"

type Device struct {
	Id            string `json:"id" gorm:"primaryKey"`
	Status        bool   `json:"status"` //切入切出状态
	Name          string `json:"name"`
	Code          string `json:"code"`
	Table         string `json:"table"`         //标志
	InterfaceType string `json:"interfaceType"` //接口类型
	Address       string `json:"address"`       //连接地址
	BaudRate      int    `json:"baudRate"`      //波特率
	StopBits      int    `json:"stopBits"`      //停止位
	DataBits      int    `json:"dataBits"`      //停止位
	Parity        string `json:"parity"`        //校验位
	ProtocolType  string `json:"protocolType"`  //协议类型
	DeviceAddress string `json:"deviceAddress"` //设备地址
	WriteTimeout  int    `json:"writeTimeout"`
	ReadTimeout   int    `json:"readTimeout"`
}

func (d *Device) Identifier() string {
	return fmt.Sprintf("%s_%s_%s", d.Id, d.Name, d.Code)
}

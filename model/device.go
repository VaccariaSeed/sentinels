package model

var Devices []*Device

func init() {
	dev1 := &Device{
		Id:            "1",
		Status:        false,
		Name:          "test1",
		Code:          "td1",
		Table:         "td1_1",
		InterfaceType: "TCP",
		Address:       "127.0.0.1:998",
		BaudRate:      "0",
		StopBits:      "0",
		DataBits:      "0",
		Parity:        "",
		ProtocolType:  "698",
		DeviceAddress: "1",
	}
	Devices = append(Devices, dev1)
	dev2 := Device{
		Id:            "2",
		Status:        true,
		Name:          "testq",
		Code:          "tdq",
		Table:         "tdw_w",
		InterfaceType: "RS485",
		Address:       "com3",
		BaudRate:      "0",
		StopBits:      "0",
		DataBits:      "0",
		Parity:        "E",
		ProtocolType:  "HTTP",
		DeviceAddress: "1",
	}
	Devices = append(Devices, &dev2)
}

type Device struct {
	Id            string `json:"id"`
	Status        bool   `json:"status"` //切入切出状态
	Name          string `json:"name"`
	Code          string `json:"code"`
	Table         string `json:"table"`         //表名
	InterfaceType string `json:"interfaceType"` //接口类型
	Address       string `json:"address"`       //连接地址
	BaudRate      string `json:"baudRate"`      //波特率
	StopBits      string `json:"stopBits"`      //停止位
	DataBits      string `json:"dataBits"`      //停止位
	Parity        string `json:"parity"`        //校验位
	ProtocolType  string `json:"protocolType"`  //协议类型
	DeviceAddress string `json:"deviceAddress"` //设备地址

}

func NewDevice() []*Device {
	return Devices
}

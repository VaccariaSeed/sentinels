package model

import "fmt"

var Points []*Point

func init() {
	for index := 0; index < 20; index++ {
		point := &Point{
			ID:             fmt.Sprintf("%d", index),
			FunctionCode:   fmt.Sprintf("%d", index),
			Address:        "0x33",
			DataType:       "uint16",
			Tag:            "T33",
			LuaExpression:  "ewqeqw",
			Description:    "测试",
			AlarmFlag:      "======",
			AlarmLevel:     "low",
			Multiplier:     1,
			Unit:           "",
			Priority:       "",
			Endianness:     "",
			BitCalculation: "",
			StartBit:       0,
			EndBit:         0,
			DeviceID:       "2",
		}
		Points = append(Points, point)
	}
}

type Point struct {
	ID             string  `json:"id" gorm:"primaryKey"`
	FunctionCode   string  `json:"functionCode"` //功能码
	Address        string  `json:"address"`      //点位地址
	DataType       string  `json:"dataType"`     //数据类型
	Tag            string  `json:"tag"`
	LuaExpression  string  `json:"luaExpression"`  //lua表达式
	Description    string  `json:"description"`    //描述
	AlarmFlag      string  `json:"alarmFlag"`      //告警标志
	AlarmLevel     string  `json:"alarmLevel"`     //告警等级
	Multiplier     float64 `json:"multiplier"`     //倍率
	Unit           string  `json:"unit"`           //单位
	Priority       string  `json:"priority"`       //优先级
	Endianness     string  `json:"endianness"`     //大小端
	BitCalculation string  `json:"bitCalculation"` //bit位计算
	StartBit       int     `json:"startBit"`       //起始bit
	EndBit         int     `json:"endBit"`         //结束bit
	StorageMethod  string  `json:"storageMethod"`  //存储方式，变存，直存
	Offset         int     `json:"offset"`         //偏移量
	Store          int     `json:"store"`          //入库间隔秒
	DeviceID       string  `json:"deviceId"`       //设备id
}

type PointResponse struct {
	TotalCount int      `json:"totalCount"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
	TotalPages int      `json:"totalPages"`
	Points     []*Point `json:"points"`
}

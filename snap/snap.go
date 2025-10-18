package snap

import "sentinels/model"

var _ PointSnap = (*ModbusPointSnap)(nil)

type PointSnap interface {
	Address() []byte //地址
	Length() byte    //数量
	FunctionCode() []byte
	String() string
	Point(key interface{}) ([]*model.Point, error)
	Parse(resp []byte) (map[string]interface{}, error)
}

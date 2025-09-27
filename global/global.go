package global

import "time"

const (
	defaultPort       = 9970
	defaultStaticPath = "./static"
	defaultDbPath     = "./bin/sentinels.db"
	DefaultTimeout    = 5000 * time.Millisecond
)

func init() {
	flushConf()
	flushSystemLog()
}

// 连接模式
const (
	RS485          = "RS485"
	RS232          = "RS232"
	TcpClient      = "TCP_CLIENT"
	TcpClientReuse = "TCP_CLIENT_REUSE" //支持端口复用
	TcpServer      = "TCP_SERVER"
	UdpClient      = "UDP_CLIENT"
	UdpServer      = "UDP_SERVER"
	Can            = "CAN"
	SPS            = "SPS" //串口服务器
)

// 规约类型
const (
	ModbusRTU = "modbusRTU"
	ModbusTCP = "modbusTCP"
	GB28181   = "GB28181"
	GBT698    = "698.45"
	DLT645    = "DLT645"
	GBT13761  = "1376.1"
	GBT1867   = "1867"
)

// 优先级
const (
	PriorityHigh   = 3 //高
	PriorityLow    = 2 //低
	PriorityMiddle = 1 //中
)

const (
	BigEndian    = "BIG"
	LittleEndian = "LITTLE"
)

const (
	AllBit      = "all"
	SingleBit   = "single"
	MultipleBit = "multiple"
)

const (
	DTInt8       = "int8"
	DTInt16      = "int16"
	DTInt32      = "int32"
	DTInt64      = "int64"
	DTUint8      = "uint8"
	DTUint16     = "uint16"
	DTUint32     = "uint32"
	DTUint64     = "uint64"
	DTFloat32    = "float32"
	DTFloat64    = "float64"
	DTInt8Arr    = "int8Arr"
	DTInt16Arr   = "int16Arr"
	DTInt32Arr   = "int32Arr"
	DTInt64Arr   = "int64Arr"
	DTUint8Arr   = "uint8Arr"
	DTUint16Arr  = "uint16Arr"
	DTUint32Arr  = "uint32Arr"
	DTUint64Arr  = "uint64Arr"
	DTFloat32Arr = "float32Arr"
	DTFloat64Arr = "float64Arr"
	DTString     = "string"
)

const (
	LogoTypeId      = "id"
	LogoTypeAddress = "address"
)

// 命令类型
const (
	CopyRead    = "copyRead"    //抄读
	SetCmd      = "setCmd"      //设置
	Passthrough = "passthrough" //透传
)

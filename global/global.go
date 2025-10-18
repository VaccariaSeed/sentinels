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
	DLT645FE  = "DLT645FE" //报文前存在4个0xFE
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
	DTBit     = "bit"
	DTInt8    = "int8"
	DTInt16   = "int16"
	DTInt32   = "int32"
	DTInt64   = "int64"
	DTByte    = "byte"
	DTUint16  = "uint16"
	DTUint32  = "uint32"
	DTUint64  = "uint64"
	DTFloat32 = "float32"
	DTFloat64 = "float64"
)

const (
	LogoTypeId        = "id"
	LogoTypeTableFlag = "tableFlag"
)

// 命令类型
const (
	CopyRead    = "copyRead"    //抄读
	SetCmd      = "setCmd"      //设置
	Passthrough = "passthrough" //透传
)

package global

const (
	defaultPort       = 9970
	defaultStaticPath = "./static"
)

func init() {
	flushConf()
	flushSystemLog()
}

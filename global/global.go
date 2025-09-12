package global

const (
	defaultPort       = 9970
	defaultStaticPath = "./static"
	defaultDbPath     = "./bin/sentinels.db"
)

func init() {
	flushConf()
	flushSystemLog()
}

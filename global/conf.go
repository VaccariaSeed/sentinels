package global

import (
	"gopkg.in/ini.v1"
)

var Config = &Conf{
	Port:   defaultPort,
	Static: defaultStaticPath,
	DbPath: defaultDbPath,
}

type Conf struct {
	Port   int    `ini:"port"`
	Static string `ini:"static"`
	DbPath string `ini:"dbPath"`
}

func flushConf() {
	//解析
	cfg, err := ini.Load(configPath)
	if err != nil {
		SystemLog.Errorf("Fail to load config file: %v", err)
		return
	}
	err = cfg.MapTo(Config)
	if err != nil {
		SystemLog.Errorf("Fail to parse config file: %v", err)
	}
}

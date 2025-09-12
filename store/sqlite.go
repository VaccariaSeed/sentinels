package store

import (
	"fmt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"os"
	"sentinels/global"
	"sentinels/model"
	"sync"
	"time"
)

var DbClient *SqliteClient

func init() {
	// 配置 GORM 日志
	// 连接数据库
	global.SystemLog.Infof("db path:%s", global.Config.DbPath)
	db, err := gorm.Open(sqlite.Open(global.Config.DbPath), nil)
	if err != nil {
		global.SystemLog.Errorf("sqlite open err:%s", err.Error())
		os.Exit(1)
	}
	DbClient = &SqliteClient{db: db}
	//初始化表
	err = db.AutoMigrate(&model.Device{})
	if err != nil {
		global.SystemLog.Errorf("sqlite Device migrate err:%s", err.Error())
		os.Exit(1)
	}
	err = db.AutoMigrate(&model.Point{})
	if err != nil {
		global.SystemLog.Errorf("sqlite Point migrate err:%s", err.Error())
		os.Exit(1)
	}
}

type SqliteClient struct {
	lock sync.Mutex //单一锁
	db   *gorm.DB
}

func (s *SqliteClient) SelectAllDevice() []*model.Device {
	s.lock.Lock()
	defer s.lock.Unlock()
	devices := make([]*model.Device, 0)
	_ = s.db.Find(&devices)
	return devices
}

func (s *SqliteClient) SaveDevice(device *model.Device) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	device.Id = fmt.Sprintf("%d", time.Now().Unix())
	result := s.db.Create(device)
	return result.Error
}

func (s *SqliteClient) UpdateDeviceStatus(id string, status bool) error {
	//切入切出
	s.lock.Lock()
	defer s.lock.Unlock()
	tx := s.db.Model(&model.Device{}).Where("id = ?", id).Update("status", status)
	return tx.Error
}

func (s *SqliteClient) SelectDeviceById(id string) (*model.Device, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var device *model.Device
	err := s.db.First(&device, "id = ?", id).Error
	return device, err
}

func (s *SqliteClient) UpdateDevice(m *model.Device) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.db.Save(m).Error
}

func (s *SqliteClient) DeleteDevice(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.db.Where("id = ?", id).Delete(&model.Device{}).Error
}

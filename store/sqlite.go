package store

import (
	"fmt"
	"os"
	"sentinels/global"
	"sentinels/model"
	"strings"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

var DbClient *SqliteClient

type SqliteClient struct {
	lock sync.Mutex //单一锁
	db   *gorm.DB
}

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
	err = db.AutoMigrate(&model.Collect{})
	if err != nil {
		global.SystemLog.Errorf("sqlite Collect migrate err:%s", err.Error())
		os.Exit(1)
	}
}

func (s *SqliteClient) SelectAllDevice() []*model.Device {
	s.lock.Lock()
	defer s.lock.Unlock()
	devices := make([]*model.Device, 0)
	_ = s.db.Find(&devices)
	return devices
}

func (s *SqliteClient) SelectCutInDevice() []*model.Device {
	s.lock.Lock()
	defer s.lock.Unlock()
	devices := make([]*model.Device, 0)
	_ = s.db.Find(&devices, "status = ?", 1)
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
	err := s.db.Where("id = ?", id).Delete(&model.Device{}).Error
	if err == nil {
		err = s.db.Delete(&model.Point{}, "device_id = ?", id).Error
	}
	return err
}

func (s *SqliteClient) SavePoint(m *model.Point) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if strings.TrimSpace(m.ID) == "" {
		m.ID = fmt.Sprintf("%d", time.Now().Unix())
	}
	return s.db.Save(m).Error
}

func (s *SqliteClient) SelectPoints(page int, size int, id int, mark string) (int, []*model.Point, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	//查询总记录数
	var totalCount int64
	query := s.db.Model(&model.Point{}).Where("device_id = ?", id)
	if mark != "" {
		query.Where("description LIKE ?", "%"+mark+"%")
	}
	err := query.Count(&totalCount).Error
	if err != nil {
		return 0, nil, err
	}
	if totalCount == 0 {
		return 0, make([]*model.Point, 0), nil
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * size
	var points []*model.Point
	err = query.Limit(size).Offset(offset).Find(&points).Error
	if err != nil {
		return 0, nil, err
	}
	return int(totalCount), points, nil
}

func (s *SqliteClient) SelectPointById(id string) (*model.Point, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var point *model.Point
	err := s.db.First(&point, "id = ?", id).Error
	return point, err
}

func (s *SqliteClient) DeletePoint(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.db.Where("id = ?", id).Delete(&model.Point{}).Error
}

func (s *SqliteClient) SelectCollectByDeviceId(deviceId string) ([]*model.Collect, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	collects := make([]*model.Collect, 0)
	err := s.db.Find(&collects, "device_id = ?", deviceId).Error
	return collects, err
}

func (s *SqliteClient) SaveCollect(m *model.Collect) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	if strings.TrimSpace(m.ID) == "" {
		m.ID = fmt.Sprintf("%d", time.Now().Unix())
	}
	return s.db.Save(m).Error
}

func (s *SqliteClient) SelectOneCollect(id string) (*model.Collect, error) {
	s.lock.Lock()
	defer s.lock.Unlock()
	var collect *model.Collect
	err := s.db.First(&collect, "id = ?", id).Error
	return collect, err
}

func (s *SqliteClient) DeleteCollect(id string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	return s.db.Where("id = ?", id).Delete(&model.Collect{}).Error
}

func (s *SqliteClient) SelectPointsByDeviceId(id string) []*model.Point {
	s.lock.Lock()
	defer s.lock.Unlock()
	points := make([]*model.Point, 0)
	_ = s.db.Find(&points, "device_id = ?", id)
	return points
}

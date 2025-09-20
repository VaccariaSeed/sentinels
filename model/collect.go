package model

type Collect struct {
	ID           string `json:"id" gorm:"primaryKey"`
	Description  string `json:"description"`
	RuleFuncCode byte   `json:"ruleFuncCode"`
	StartPoint   string `json:"startPoint"`
	EndPoint     string `json:"endPoint"`
	DeviceId     string `json:"deviceId"`
}

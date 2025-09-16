package models

import "time"

// Authen represents session records
type Authen struct {
	SessionID  string     `gorm:"primaryKey;size:128" json:"session_id"`
	StartDate  time.Time  `json:"start_date"`
	EndDate    *time.Time `json:"end_date"`
	LastAccess time.Time  `json:"last_access"`
	UserID     string     `gorm:"size:10" json:"user_id"`
}

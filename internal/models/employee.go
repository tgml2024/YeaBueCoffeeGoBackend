package models

import "time"

type Employee struct {
	EmpID     string     `gorm:"primaryKey;size:10" json:"emp_id"`
	Username  string     `gorm:"uniqueIndex;size:20" json:"username"`
	Password  string     `json:"-"`
	FirstName string     `gorm:"size:100" json:"first_name"`
	LastName  string     `gorm:"size:100" json:"last_name"`
	TitleName string     `gorm:"size:10" json:"title_name"`
	Nickname  string     `gorm:"size:20" json:"nickname"`
	Email     string     `gorm:"size:50" json:"email"`
	Position  string     `gorm:"size:20;default:employee" json:"position"`
	AddDate   time.Time  `json:"add_date"`
	AddUser   string     `gorm:"size:10" json:"add_user"`
	EditDate  *time.Time `json:"edit_date"`
	EditUser  *string    `gorm:"size:10" json:"edit_user"`
	DelDate   *time.Time `json:"del_date"`
	DelUser   *string    `gorm:"size:10" json:"del_user"`
}

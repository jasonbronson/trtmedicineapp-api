package models

import "time"

type User struct {
	ID                       string     `gorm:"primaryKey;type:text" json:"id"`
	Email                    string     `gorm:"uniqueIndex;not null;type:text" json:"email"`
	PasswordHash             *string    `gorm:"type:text" json:"-"`
	GoogleSub                *string    `gorm:"uniqueIndex;type:text" json:"-"`
	AuthProvider             string     `gorm:"not null;type:text" json:"auth_provider"`
	EmailVerified            bool       `gorm:"not null;default:false" json:"email_verified"`
	ApplePurchaseReceiptData *string    `gorm:"type:text" json:"apple_purchase_receipt_data,omitempty"`
	AppleOrderID             *string    `gorm:"type:text" json:"apple_order_id,omitempty"`
	PlanID                   *string    `gorm:"type:text" json:"plan_id,omitempty"`
	SubscriptionStatus       string     `gorm:"not null;default:'inactive';type:text" json:"subscription_status"`
	CancelDate               *time.Time `json:"cancel_date,omitempty"`
	SubscriptionExpiresAt    *time.Time `json:"subscription_expires_at,omitempty"`
	CreatedAt                time.Time  `json:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at"`
	DeletedAt                *time.Time `gorm:"index" json:"-"`
}

type Medicine struct {
	ID         string     `gorm:"primaryKey;type:text" json:"id"`
	UserID     string     `gorm:"index;not null;type:text" json:"user_id"`
	Name       string     `gorm:"not null;type:text" json:"name"`
	DoseAmount string     `gorm:"type:text" json:"dose_amount"`
	DoseUnit   string     `gorm:"type:text" json:"dose_unit"`
	Notes      string     `gorm:"type:text" json:"notes"`
	SoundID    *string    `gorm:"type:text" json:"sound_id"`
	Active     bool       `gorm:"not null;default:true" json:"active"`
	Color      string     `gorm:"type:text;default:blue" json:"color"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	DeletedAt  *time.Time `gorm:"index" json:"-"`
	Schedules  []Schedule `json:"schedules,omitempty"`
}

type Schedule struct {
	ID                string       `gorm:"primaryKey;type:text" json:"id"`
	UserID            string       `gorm:"index;not null;type:text" json:"user_id"`
	MedicineID        string       `gorm:"index;not null;type:text" json:"medicine_id"`
	ScheduleType      string       `gorm:"not null;type:text" json:"schedule_type"`
	StartDate         time.Time    `gorm:"not null" json:"start_date"`
	EndDate           *time.Time   `json:"end_date"`
	TimeOfDay         *string      `gorm:"type:text" json:"time_of_day"`
	IntervalHours     *int         `json:"interval_hours"`
	IntervalDays      *int         `json:"interval_days"`
	WeeklyInterval    *int         `json:"weekly_interval"`
	MonthlyDay        *int         `json:"monthly_day"`
	CycleActiveDays   *int         `json:"cycle_active_days"`
	CycleInactiveDays *int         `json:"cycle_inactive_days"`
	CycleStartDate    *time.Time   `json:"cycle_start_date"`
	DaysOfWeek        []DayOfWeek  `json:"days_of_week,omitempty"`
	DaysOfMonth       []DayOfMonth `json:"days_of_month,omitempty"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	DeletedAt         *time.Time   `gorm:"index" json:"-"`
}

type DayOfWeek struct {
	ID         string `gorm:"primaryKey;type:text" json:"id"`
	ScheduleID string `gorm:"index;not null;type:text" json:"schedule_id"`
	Day        int    `gorm:"not null" json:"day"`
}

type DayOfMonth struct {
	ID         string `gorm:"primaryKey;type:text" json:"id"`
	ScheduleID string `gorm:"index;not null;type:text" json:"schedule_id"`
	Day        int    `gorm:"not null" json:"day"`
}

type Sound struct {
	ID        string     `gorm:"primaryKey;type:text" json:"id"`
	UserID    string     `gorm:"index;not null;type:text" json:"user_id"`
	Name      string     `gorm:"not null;type:text" json:"name"`
	FileName  string     `gorm:"not null;type:text" json:"file_name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
}

type ReminderLog struct {
	ID         string    `gorm:"primaryKey;type:text" json:"id"`
	UserID     string    `gorm:"index;not null;type:text" json:"user_id"`
	MedicineID string    `gorm:"index;not null;type:text" json:"medicine_id"`
	ScheduleID string    `gorm:"index;not null;type:text" json:"schedule_id"`
	DoseAmount string    `gorm:"type:text" json:"dose_amount"`
	DoseUnit   string    `gorm:"type:text" json:"dose_unit"`
	DueAt      time.Time `gorm:"index;not null" json:"due_at"`
	Action     string    `gorm:"not null;type:text" json:"action"`
	Notes      string    `gorm:"type:text" json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
}

type ManualNote struct {
	ID         string    `gorm:"primaryKey;type:text" json:"id"`
	UserID     string    `gorm:"index;not null;type:text" json:"user_id"`
	MedicineID *string   `gorm:"index;type:text" json:"medicine_id"`
	NoteDate   time.Time `gorm:"index;not null" json:"note_date"`
	Notes      string    `gorm:"type:text" json:"notes"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

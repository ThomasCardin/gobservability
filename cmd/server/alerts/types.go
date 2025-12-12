package alerts

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MetricType string

const (
	MetricCPU     MetricType = "cpu"
	MetricMemory  MetricType = "memory"
	MetricNetwork MetricType = "network"
	MetricDisk    MetricType = "disk"
)

type OperatorType string

const (
	OpGreater      OperatorType = ">"
	OpGreaterEqual OperatorType = ">="
	OpLess         OperatorType = "<"
	OpLessEqual    OperatorType = "<="
)

type AlertStatus string

const (
	StatusFiring   AlertStatus = "firing"
	StatusResolved AlertStatus = "resolved"
)

// AlertRule - GORM model for alerting rules
type AlertRule struct {
	ID                     uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	NodeName               string       `gorm:"not null;index" json:"node_name"`
	Target                 string       `gorm:"not null" json:"target"`                      // "node" or "pod:name"
	Metric                 MetricType   `gorm:"type:varchar(20);not null" json:"metric"`
	Operator               OperatorType `gorm:"type:varchar(5);not null" json:"operator"`
	Threshold              float64      `gorm:"not null" json:"threshold"`
	ResolveThreshold       *float64     `gorm:"default:null" json:"resolve_threshold,omitempty"`
	DurationSeconds        int          `gorm:"default:60" json:"duration_seconds"`
	DiscordFrequencyMinutes int          `gorm:"default:5" json:"discord_frequency_minutes"` // Discord notification frequency
	DiscordMentions        string       `gorm:"default:''" json:"discord_mentions"`         // Discord mentions (@admin, @here, etc.)
	Enabled                bool         `gorm:"default:true" json:"enabled"`
	CreatedAt              time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt              time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt              gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations GORM
	Alerts []Alert `gorm:"foreignKey:RuleID" json:"alerts,omitempty"`
}

// Alert - GORM model for alerts
type Alert struct {
	ID                uuid.UUID    `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	RuleID            uuid.UUID    `gorm:"not null;index" json:"rule_id"`
	NodeName          string       `gorm:"not null;index" json:"node_name"`
	Target            string       `gorm:"not null" json:"target"`
	Metric            MetricType   `gorm:"type:varchar(20);not null" json:"metric"`
	Status            AlertStatus  `gorm:"type:varchar(10);not null;index" json:"status"`
	CurrentValue      float64      `gorm:"not null" json:"current_value"`
	ThresholdValue    float64      `gorm:"not null" json:"threshold_value"`
	StartedAt         time.Time    `gorm:"not null;index" json:"started_at"`
	ResolvedAt        *time.Time   `gorm:"default:null" json:"resolved_at,omitempty"`
	LastNotifiedAt    *time.Time   `gorm:"default:null" json:"last_notified_at,omitempty"`
	NotificationCount int          `gorm:"default:0" json:"notification_count"`
	CreatedAt         time.Time    `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt         time.Time    `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`

	// Relations GORM
	Rule AlertRule `gorm:"foreignKey:RuleID" json:"rule,omitempty"`
}

// AlertEvaluation - Structure for cached evaluation (not GORM)
type AlertEvaluation struct {
	RuleID       uuid.UUID
	NodeName     string
	Target       string
	Metric       MetricType
	Value        float64
	Threshold    float64
	IsTriggered  bool
	FirstSeen    time.Time
	LastChecked  time.Time
}
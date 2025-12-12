package alerts

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/patrickmn/go-cache"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type AlertsStorage struct {
	db              *gorm.DB
	rulesCache      *cache.Cache // Rules par node (1h TTL)
	evaluationCache *cache.Cache // Ongoing evaluations (10min TTL)
	rateLimitCache  *cache.Cache // Discord rate limiting (5min TTL)
}

func NewAlertsStorage(postgresURL string) (*AlertsStorage, error) {
	// Connection GORM avec optimisations
	db, err := gorm.Open(postgres.Open(postgresURL), &gorm.Config{
		PrepareStmt:                              true,  // Performance boost
		DisableForeignKeyConstraintWhenMigrating: false, // Keep FK constraints
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Configuration du pool de connexions
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	// Auto-migration of schemas
	err = db.AutoMigrate(&AlertRule{}, &Alert{})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return &AlertsStorage{
		db:              db,
		rulesCache:      cache.New(5*time.Minute, 1*time.Minute),  // Shorter cache for reactive UI
		evaluationCache: cache.New(10*time.Minute, 5*time.Minute),
		rateLimitCache:  cache.New(5*time.Minute, 1*time.Minute),
	}, nil
}

// CRUD for AlertRules with GORM
func (s *AlertsStorage) CreateRule(rule *AlertRule) error {
	result := s.db.Create(rule)
	if result.Error == nil {
		s.invalidateRulesCache(rule.NodeName)
	}
	return result.Error
}

// GetRulesByNode retrieves all rules for UI display
func (s *AlertsStorage) GetRulesByNode(nodeName string) ([]AlertRule, error) {
	// Cache first
	cacheKey := "rules_all_" + nodeName
	if cached, found := s.rulesCache.Get(cacheKey); found {
		return cached.([]AlertRule), nil
	}

	var rules []AlertRule
	err := s.db.Where("node_name = ?", nodeName).
		Order("created_at DESC").
		Find(&rules).Error

	if err == nil {
		s.rulesCache.Set(cacheKey, rules, cache.DefaultExpiration)
	}
	return rules, err
}

// GetEnabledRulesByNode retrieves only active rules for evaluation
func (s *AlertsStorage) GetEnabledRulesByNode(nodeName string) ([]AlertRule, error) {
	// Cache first
	cacheKey := "rules_enabled_" + nodeName
	if cached, found := s.rulesCache.Get(cacheKey); found {
		return cached.([]AlertRule), nil
	}

	var rules []AlertRule
	err := s.db.Where("node_name = ? AND enabled = ?", nodeName, true).
		Order("created_at DESC").
		Find(&rules).Error

	if err == nil {
		s.rulesCache.Set(cacheKey, rules, cache.DefaultExpiration)
	}
	return rules, err
}

func (s *AlertsStorage) UpdateRule(rule *AlertRule) error {
	result := s.db.Save(rule)
	if result.Error == nil {
		s.invalidateRulesCache(rule.NodeName)
	}
	return result.Error
}

func (s *AlertsStorage) DeleteRule(ruleID uuid.UUID) error {
	var rule AlertRule
	if err := s.db.First(&rule, ruleID).Error; err != nil {
		return err
	}

	// Soft delete avec GORM
	result := s.db.Delete(&rule)
	if result.Error == nil {
		s.invalidateRulesCache(rule.NodeName)
	}
	return result.Error
}

func (s *AlertsStorage) GetRuleByID(ruleID uuid.UUID) (*AlertRule, error) {
	var rule AlertRule
	err := s.db.First(&rule, ruleID).Error
	if err != nil {
		return nil, err
	}
	return &rule, nil
}

// Gestion des alertes avec GORM
func (s *AlertsStorage) CreateAlert(alert *Alert) error {
	return s.db.Create(alert).Error
}

func (s *AlertsStorage) ResolveAlert(alertID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"status":      StatusResolved,
			"resolved_at": now,
		}).Error
}

func (s *AlertsStorage) GetActiveAlerts(nodeName string) ([]Alert, error) {
	var alerts []Alert
	err := s.db.Preload("Rule").
		Where("node_name = ? AND status = ?", nodeName, StatusFiring).
		Order("started_at DESC").
		Find(&alerts).Error
	return alerts, err
}

func (s *AlertsStorage) GetActiveAlertByRule(ruleID uuid.UUID, target string, metric MetricType) (*Alert, error) {
	var alert Alert
	// Use silent session to avoid "record not found" logs
	err := s.db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).
		Where("rule_id = ? AND target = ? AND metric = ? AND status = ?",
			ruleID, target, metric, StatusFiring).
		First(&alert).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &alert, err
}

func (s *AlertsStorage) GetAlertHistory(nodeName string, days int) ([]Alert, error) {
	var alerts []Alert
	since := time.Now().AddDate(0, 0, -days)

	err := s.db.Preload("Rule").
		Where("node_name = ? AND started_at > ?", nodeName, since).
		Order("started_at DESC").
		Limit(100).
		Find(&alerts).Error
	return alerts, err
}

func (s *AlertsStorage) UpdateAlertNotification(alertID uuid.UUID) error {
	now := time.Now()
	return s.db.Model(&Alert{}).
		Where("id = ?", alertID).
		Updates(map[string]interface{}{
			"last_notified_at":    now,
			"notification_count": gorm.Expr("notification_count + 1"),
		}).Error
}

// Cache management
func (s *AlertsStorage) invalidateRulesCache(nodeName string) {
	s.rulesCache.Delete("rules_all_" + nodeName)
	s.rulesCache.Delete("rules_enabled_" + nodeName)
}

func (s *AlertsStorage) SetEvaluation(eval *AlertEvaluation) {
	key := fmt.Sprintf("eval_%s_%s_%s", eval.NodeName, eval.Target, eval.Metric)
	s.evaluationCache.Set(key, eval, cache.DefaultExpiration)
}

func (s *AlertsStorage) GetEvaluation(nodeName, target string, metric MetricType) *AlertEvaluation {
	key := fmt.Sprintf("eval_%s_%s_%s", nodeName, target, metric)
	if cached, found := s.evaluationCache.Get(key); found {
		eval := cached.(*AlertEvaluation)
		return eval
	}
	return nil
}

// CanNotifyDiscord checks if we can notify Discord (to avoid spam)
func (s *AlertsStorage) CanNotifyDiscord(ruleID uuid.UUID) bool {
	key := "discord_spam_" + ruleID.String()
	if _, found := s.rateLimitCache.Get(key); found {
		return false // Protection anti-spam (30s)
	}
	s.rateLimitCache.Set(key, true, 30*time.Second)
	return true
}

// GetAlertsNeedingNotification returns active alerts that need to be re-notified
func (s *AlertsStorage) GetAlertsNeedingNotification() ([]Alert, error) {
	var alerts []Alert
	err := s.db.Preload("Rule").
		Where("status = ?", StatusFiring).
		Find(&alerts).Error
	
	if err != nil {
		return nil, err
	}
	
	// Filter those that need to be re-notified
	var needingNotification []Alert
	now := time.Now()
	
	for _, alert := range alerts {
		// If never notified OR if delay has elapsed
		if alert.LastNotifiedAt == nil {
			needingNotification = append(needingNotification, alert)
		} else {
			nextNotification := alert.LastNotifiedAt.Add(time.Duration(alert.Rule.DiscordFrequencyMinutes) * time.Minute)
			if now.After(nextNotification) {
				needingNotification = append(needingNotification, alert)
			}
		}
	}
	
	return needingNotification, nil
}

// GetDB exposes DB for special cases
func (s *AlertsStorage) GetDB() *gorm.DB {
	return s.db
}

func (s *AlertsStorage) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
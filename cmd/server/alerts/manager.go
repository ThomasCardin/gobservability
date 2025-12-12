package alerts

import (
	"fmt"
	"os"
	"time"

	"github.com/ThomasCardin/gobservability/shared/types"
)

type AlertsManager struct {
	storage   *AlertsStorage
	evaluator *AlertEvaluator
	discord   *DiscordNotifier
	stopChan  chan bool
}

func NewAlertsManager() (*AlertsManager, error) {
	// Configuration
	postgresURL := os.Getenv("POSTGRES_URL")
	if postgresURL == "" {
		return nil, fmt.Errorf("POSTGRES_URL environment variable required")
	}

	discordURL := os.Getenv("DISCORD_WEBHOOK_URL")
	if discordURL == "" {
		return nil, fmt.Errorf("DISCORD_WEBHOOK_URL environment variable required")
	}

	// Initialisation avec GORM
	storage, err := NewAlertsStorage(postgresURL)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	discord := NewDiscordNotifier(discordURL)
	evaluator := NewAlertEvaluator(storage, discord)

	manager := &AlertsManager{
		storage:   storage,
		evaluator: evaluator,
		discord:   discord,
		stopChan:  make(chan bool),
	}
	
	// Start periodic notifications
	go manager.startPeriodicNotifications()
	
	return manager, nil
}

// Main entry point - called from storage.StoreNodeStats
func (m *AlertsManager) EvaluateMetrics(nodeStats types.NodeStatsPayload) {
	if err := m.evaluator.EvaluateMetrics(nodeStats); err != nil {
		fmt.Printf("Error evaluating alerts for node %s: %v\n", nodeStats.NodeName, err)
	}
}

// API pour l'interface web
func (m *AlertsManager) GetStorage() *AlertsStorage {
	return m.storage
}

func (m *AlertsManager) GetDiscord() *DiscordNotifier {
	return m.discord
}

// startPeriodicNotifications lance une goroutine pour re-notifier les alertes actives
func (m *AlertsManager) startPeriodicNotifications() {
	ticker := time.NewTicker(1 * time.Minute) // Check every minute
	defer ticker.Stop()
	
	for {
		select {
		case <-m.stopChan:
			fmt.Println("[ALERT] Stopping periodic notifications")
			return
		case <-ticker.C:
			if err := m.processPeriodicNotifications(); err != nil {
				fmt.Printf("[ALERT ERROR] Failed to process periodic notifications: %v\n", err)
			}
		}
	}
}

func (m *AlertsManager) processPeriodicNotifications() error {
	alerts, err := m.storage.GetAlertsNeedingNotification()
	if err != nil {
		return err
	}
	
	for _, alert := range alerts {
		// Anti-spam protection
		if !m.storage.CanNotifyDiscord(alert.RuleID) {
			continue
		}
		
		// Re-notify Discord
		if err := m.discord.SendAlertOngoing(alert.Rule, &alert); err != nil {
			fmt.Printf("[ALERT ERROR] Failed to send ongoing alert notification: %v\n", err)
			continue
		}
		
		// Update last_notified_at
		if err := m.storage.UpdateAlertNotification(alert.ID); err != nil {
			fmt.Printf("[ALERT ERROR] Failed to update notification timestamp: %v\n", err)
		}
	}
	
	if len(alerts) > 0 {
		fmt.Printf("[ALERT DEBUG] Sent %d periodic Discord notifications\n", len(alerts))
	}
	
	return nil
}

func (m *AlertsManager) Close() error {
	// Stop periodic notifications
	select {
	case m.stopChan <- true:
	default:
	}
	return m.storage.Close()
}
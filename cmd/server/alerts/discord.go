package alerts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

func NewDiscordNotifier(webhookURL string) *DiscordNotifier {
	return &DiscordNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}
}

type DiscordMessage struct {
	Content string         `json:"content,omitempty"`
	Embeds  []DiscordEmbed `json:"embeds"`
}

type DiscordEmbed struct {
	Title       string         `json:"title"`
	Description string         `json:"description"`
	Color       int            `json:"color"`
	Fields      []DiscordField `json:"fields"`
	Timestamp   string         `json:"timestamp"`
}

type DiscordField struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline"`
}

func (d *DiscordNotifier) SendAlertTriggered(rule AlertRule, alert *Alert) error {
	// Determine unit based on metric
	unit := d.getMetricUnit(rule.Metric)
	
	embed := DiscordEmbed{
		Title:       "🚨 ALERT TRIGGERED",
		Description: fmt.Sprintf("Alert triggered for %s", rule.Target),
		Color:       15158332, // Rouge
		Timestamp:   alert.StartedAt.Format(time.RFC3339),
		Fields: []DiscordField{
			{Name: "Node", Value: rule.NodeName, Inline: true},
			{Name: "Target", Value: rule.Target, Inline: true},
			{Name: "Metric", Value: string(rule.Metric), Inline: true},
			{Name: "Current Value", Value: fmt.Sprintf("%.2f%s", alert.CurrentValue, unit), Inline: true},
			{Name: "Threshold", Value: fmt.Sprintf("%s %.2f%s", rule.Operator, rule.Threshold, unit), Inline: true},
			{Name: "Duration", Value: fmt.Sprintf("%ds", rule.DurationSeconds), Inline: true},
		},
	}

	message := DiscordMessage{
		Content: rule.DiscordMentions, // Configured mentions
		Embeds:  []DiscordEmbed{embed},
	}

	return d.sendMessage(message)
}

func (d *DiscordNotifier) SendAlertResolved(rule AlertRule, alert *Alert, currentValue float64) error {
	duration := time.Since(alert.StartedAt)
	unit := d.getMetricUnit(rule.Metric)

	embed := DiscordEmbed{
		Title:       "✅ ALERT RESOLVED",
		Description: fmt.Sprintf("Alert resolved for %s", rule.Target),
		Color:       3066993, // Vert
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []DiscordField{
			{Name: "Node", Value: rule.NodeName, Inline: true},
			{Name: "Target", Value: rule.Target, Inline: true},
			{Name: "Metric", Value: string(rule.Metric), Inline: true},
			{Name: "Current Value", Value: fmt.Sprintf("%.2f%s", currentValue, unit), Inline: true},
			{Name: "Was Above", Value: fmt.Sprintf("%.2f%s", rule.Threshold, unit), Inline: true},
			{Name: "Active Duration", Value: duration.Round(time.Second).String(), Inline: true},
		},
	}

	message := DiscordMessage{
		Content: rule.DiscordMentions, // Configured mentions 
		Embeds:  []DiscordEmbed{embed},
	}

	return d.sendMessage(message)
}

// SendAlertOngoing sends a periodic notification for an ongoing alert
func (d *DiscordNotifier) SendAlertOngoing(rule AlertRule, alert *Alert) error {
	duration := time.Since(alert.StartedAt)
	unit := d.getMetricUnit(rule.Metric)
	
	embed := DiscordEmbed{
		Title:       "⚠️ ALERT ONGOING",
		Description: fmt.Sprintf("Alert still active for %s", rule.Target),
		Color:       16776960, // Orange/Jaune
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []DiscordField{
			{Name: "Node", Value: rule.NodeName, Inline: true},
			{Name: "Target", Value: rule.Target, Inline: true},
			{Name: "Metric", Value: string(rule.Metric), Inline: true},
			{Name: "Current Value", Value: fmt.Sprintf("%.2f%s", alert.CurrentValue, unit), Inline: true},
			{Name: "Threshold", Value: fmt.Sprintf("%s %.2f%s", rule.Operator, rule.Threshold, unit), Inline: true},
			{Name: "Active For", Value: duration.Round(time.Second).String(), Inline: true},
			{Name: "Notifications", Value: fmt.Sprintf("%d sent", alert.NotificationCount+1), Inline: true},
		},
	}

	message := DiscordMessage{
		Content: rule.DiscordMentions, // Configured mentions
		Embeds:  []DiscordEmbed{embed},
	}

	return d.sendMessage(message)
}

// SendAlertDismissed sends a notification for manually dismissed alert
func (d *DiscordNotifier) SendAlertDismissed(rule AlertRule, alert *Alert, currentValue float64) error {
	duration := time.Since(alert.StartedAt)
	unit := d.getMetricUnit(rule.Metric)

	embed := DiscordEmbed{
		Title:       "🛑 ALERT DISMISSED",
		Description: fmt.Sprintf("Alert manually dismissed for %s", rule.Target),
		Color:       9936031, // Purple
		Timestamp:   time.Now().Format(time.RFC3339),
		Fields: []DiscordField{
			{Name: "Node", Value: rule.NodeName, Inline: true},
			{Name: "Target", Value: rule.Target, Inline: true},
			{Name: "Metric", Value: string(rule.Metric), Inline: true},
			{Name: "Last Value", Value: fmt.Sprintf("%.2f%s", currentValue, unit), Inline: true},
			{Name: "Threshold", Value: fmt.Sprintf("%s %.2f%s", rule.Operator, rule.Threshold, unit), Inline: true},
			{Name: "Was Active For", Value: duration.Round(time.Second).String(), Inline: true},
			{Name: "Action", Value: "Manually dismissed", Inline: false},
		},
	}

	message := DiscordMessage{
		Content: rule.DiscordMentions, // Configured mentions
		Embeds:  []DiscordEmbed{embed},
	}

	return d.sendMessage(message)
}

func (d *DiscordNotifier) getMetricUnit(metric MetricType) string {
	switch metric {
	case MetricCPU, MetricMemory:
		return "%"
	case MetricNetwork, MetricDisk:
		return "MB/s"
	default:
		return ""
	}
}

func (d *DiscordNotifier) sendMessage(message DiscordMessage) error {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal Discord message: %w", err)
	}

	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send Discord webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("discord webhook returned status %d", resp.StatusCode)
	}

	return nil
}
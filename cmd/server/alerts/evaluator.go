package alerts

import (
	"fmt"
	"time"

	"github.com/ThomasCardin/gobservability/shared/types"
)

type AlertEvaluator struct {
	storage *AlertsStorage
	discord *DiscordNotifier
}

func NewAlertEvaluator(storage *AlertsStorage, discord *DiscordNotifier) *AlertEvaluator {
	return &AlertEvaluator{
		storage: storage,
		discord: discord,
	}
}

// Main entry point - called on each received metric
func (e *AlertEvaluator) EvaluateMetrics(nodeStats types.NodeStatsPayload) error {
	rules, err := e.storage.GetEnabledRulesByNode(nodeStats.NodeName)
	if err != nil {
		return fmt.Errorf("failed to get enabled rules for node %s: %w", nodeStats.NodeName, err)
	}

	fmt.Printf("[ALERT DEBUG] Evaluating %d enabled rules for node %s\n", len(rules), nodeStats.NodeName)
	
	// Evaluate each rule
	for _, rule := range rules {
		if err := e.evaluateRule(rule, nodeStats); err != nil {
			// Log error but continue with other rules
			fmt.Printf("Error evaluating rule %s: %v\n", rule.ID, err)
		}
	}

	return nil
}

func (e *AlertEvaluator) evaluateRule(rule AlertRule, nodeStats types.NodeStatsPayload) error {
	// Extract value according to rule
	value, err := e.extractMetricValue(rule, nodeStats)
	if err != nil {
		return err
	}

	// Check if threshold is exceeded
	isTriggered := e.checkThreshold(rule.Operator, value, rule.Threshold)
	
	// Debug log
	fmt.Printf("[ALERT DEBUG] Rule %s: target=%s, metric=%s, value=%.2f, threshold=%.2f, operator=%s, triggered=%v\n",
		rule.ID, rule.Target, rule.Metric, value, rule.Threshold, rule.Operator, isTriggered)

	// Get previous evaluation
	eval := e.storage.GetEvaluation(rule.NodeName, rule.Target, rule.Metric)

	if eval == nil {
		// First evaluation
		eval = &AlertEvaluation{
			RuleID:      rule.ID,
			NodeName:    rule.NodeName,
			Target:      rule.Target,
			Metric:      rule.Metric,
			Value:       value,
			Threshold:   rule.Threshold,
			IsTriggered: isTriggered,
			FirstSeen:   time.Now(),
			LastChecked: time.Now(),
		}
		e.storage.SetEvaluation(eval)
		fmt.Printf("[ALERT DEBUG] First evaluation for rule %s, will wait for next check\n", rule.ID)
		return nil // Wait for next evaluation
	}

	// Update evaluation
	eval.Value = value
	eval.LastChecked = time.Now()

	if isTriggered && !eval.IsTriggered {
		// New trigger
		eval.IsTriggered = true
		eval.FirstSeen = time.Now()

	} else if isTriggered && eval.IsTriggered {
		// Still triggered - check duration
		duration := time.Since(eval.FirstSeen)
		fmt.Printf("[ALERT DEBUG] Rule %s still triggered for %.0f seconds (wait time: %d seconds)\n", 
			rule.ID, duration.Seconds(), rule.DurationSeconds)
		if duration >= time.Duration(rule.DurationSeconds)*time.Second {
			// Trigger alert
			fmt.Printf("[ALERT DEBUG] Triggering alert for rule %s after %v\n", rule.ID, duration)
			return e.triggerAlert(rule, eval)
		}

	} else if !isTriggered && eval.IsTriggered {
		// Potential resolution - check hysteresis
		resolveThreshold := rule.Threshold
		if rule.ResolveThreshold != nil {
			resolveThreshold = *rule.ResolveThreshold
		}

		isResolved := e.checkResolveThreshold(rule.Operator, value, resolveThreshold)
		if isResolved {
			// Resolved according to hysteresis
			eval.IsTriggered = false
			return e.resolveAlert(rule, eval)
		}
	}

	e.storage.SetEvaluation(eval)
	return nil
}

func (e *AlertEvaluator) extractMetricValue(rule AlertRule, nodeStats types.NodeStatsPayload) (float64, error) {
	if rule.Target == "node" {
		// Node metrics
		switch rule.Metric {
		case MetricCPU:
			return nodeStats.Metrics.CPU.CPUPercent, nil
		case MetricMemory:
			total := float64(nodeStats.Metrics.Memory.MemTotal)
			available := float64(nodeStats.Metrics.Memory.MemAvailable)
			return ((total - available) / total) * 100, nil
		case MetricNetwork:
			return nodeStats.Metrics.Network.TotalRate, nil
		case MetricDisk:
			return nodeStats.Metrics.Disk.TotalRate, nil
		}
	} else if len(rule.Target) > 4 && rule.Target[:4] == "pod:" {
		// Pod metrics
		podName := rule.Target[4:] // Remove "pod:"
		for _, pod := range nodeStats.Metrics.Pods {
			if pod.Name == podName {
				switch rule.Metric {
				case MetricCPU:
					return pod.PodMetrics.CPU.CPUPercent, nil
				case MetricMemory:
					return pod.PodMetrics.Memory.MemPercent, nil
				case MetricNetwork:
					return float64(pod.PodMetrics.Network.BytesReceived+
						pod.PodMetrics.Network.BytesTransmitted) / 1024 / 1024, nil
				case MetricDisk:
					return float64(pod.PodMetrics.Disk.ReadBytes+
						pod.PodMetrics.Disk.WriteBytes) / 1024 / 1024, nil
				}
			}
		}
		return 0, fmt.Errorf("pod %s not found", podName)
	}

	return 0, fmt.Errorf("unsupported metric %s for target %s", rule.Metric, rule.Target)
}

func (e *AlertEvaluator) checkThreshold(operator OperatorType, value, threshold float64) bool {
	switch operator {
	case OpGreater:
		return value > threshold
	case OpGreaterEqual:
		return value >= threshold
	case OpLess:
		return value < threshold
	case OpLessEqual:
		return value <= threshold
	default:
		return false
	}
}

func (e *AlertEvaluator) checkResolveThreshold(operator OperatorType, value, threshold float64) bool {
	// For resolution, we invert the logic
	switch operator {
	case OpGreater, OpGreaterEqual:
		// If alert was "value > threshold", it resolves when "value < resolve_threshold"
		return value < threshold
	case OpLess, OpLessEqual:
		// If alert was "value < threshold", it resolves when "value > resolve_threshold"
		return value > threshold
	default:
		return false
	}
}

func (e *AlertEvaluator) triggerAlert(rule AlertRule, eval *AlertEvaluation) error {
	// Check if there's not already an active alert for this rule
	existingAlert, err := e.storage.GetActiveAlertByRule(rule.ID, rule.Target, rule.Metric)
	if err != nil {
		return fmt.Errorf("failed to check existing alert: %w", err)
	}
	if existingAlert != nil {
		// Alert already active, don't create a new one
		return nil
	}

	// Create alert in database
	alert := &Alert{
		RuleID:         rule.ID,
		NodeName:       rule.NodeName,
		Target:         rule.Target,
		Metric:         rule.Metric,
		Status:         StatusFiring,
		CurrentValue:   eval.Value,
		ThresholdValue: rule.Threshold,
		StartedAt:      time.Now(),
	}

	if err := e.storage.CreateAlert(alert); err != nil {
		return fmt.Errorf("failed to create alert: %w", err)
	}

	// Notify Discord (anti-spam protection only)
	if e.storage.CanNotifyDiscord(rule.ID) {
		e.discord.SendAlertTriggered(rule, alert)
		e.storage.UpdateAlertNotification(alert.ID)
	}

	return nil
}

func (e *AlertEvaluator) resolveAlert(rule AlertRule, eval *AlertEvaluation) error {
	// Find active alert
	activeAlert, err := e.storage.GetActiveAlertByRule(rule.ID, rule.Target, rule.Metric)
	if err != nil {
		return fmt.Errorf("failed to find active alert: %w", err)
	}
	if activeAlert == nil {
		// No active alert, nothing to resolve
		return nil
	}

	// Resolve alert
	if err := e.storage.ResolveAlert(activeAlert.ID); err != nil {
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	// Notify Discord
	if e.storage.CanNotifyDiscord(rule.ID) {
		e.discord.SendAlertResolved(rule, activeAlert, eval.Value)
		e.storage.UpdateAlertNotification(activeAlert.ID)
	}

	return nil
}
package api

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/ThomasCardin/gobservability/cmd/server/alerts"
	"github.com/ThomasCardin/gobservability/cmd/server/storage"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

var alertsStorage *alerts.AlertsStorage
var discordNotifier *alerts.DiscordNotifier

func SetAlertsStorage(storage *alerts.AlertsStorage) {
	alertsStorage = storage
}

func SetDiscordNotifier(discord *alerts.DiscordNotifier) {
	discordNotifier = discord
}

func getDiscordNotifier() *alerts.DiscordNotifier {
	return discordNotifier
}

// GET /api/alerts/:nodename - JSON API
func GetAlertsHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	if nodeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Node name required"})
		return
	}

	rules, err := alertsStorage.GetRulesByNode(nodeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	activeAlerts, err := alertsStorage.GetActiveAlerts(nodeName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"rules":         rules,
		"active_alerts": activeAlerts,
	})
}

// GET /api/alerts/:nodename/fragment - HTML Fragment pour HTMX
func GetAlertsFragmentHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	if nodeName == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Node name required"})
		return
	}

	rules, err := alertsStorage.GetRulesByNode(nodeName)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": err.Error()})
		return
	}

	c.HTML(http.StatusOK, "alerts-fragment.html", gin.H{
		"rules": rules,
	})
}

// POST /api/alerts/:nodename
func CreateAlertRuleHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	
	// Debug log
	fmt.Printf("[ALERT DEBUG] Creating rule for node: %s\n", nodeName)

	var rule alerts.AlertRule
	if err := c.ShouldBindJSON(&rule); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	rule.NodeName = nodeName

	if err := alertsStorage.CreateRule(&rule); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	
	fmt.Printf("[ALERT DEBUG] Rule created: ID=%s, NodeName=%s, Target=%s, Metric=%s\n", 
		rule.ID, rule.NodeName, rule.Target, rule.Metric)

	c.JSON(http.StatusCreated, rule)
}

// PUT /api/alerts/:nodename/:ruleid
func UpdateAlertRuleHandler(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("ruleid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	// Get existing rule
	existingRule, err := alertsStorage.GetRuleByID(ruleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}

	// Check if there are active alerts for this rule
	activeAlert, err := alertsStorage.GetActiveAlertByRule(ruleID, existingRule.Target, existingRule.Metric)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check active alerts"})
		return
	}
	
	if activeAlert != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot update rule with active alert. Dismiss the alert first."})
		return
	}

	// Bind updates
	var updates alerts.AlertRule
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preserve ID and NodeName
	updates.ID = existingRule.ID
	updates.NodeName = existingRule.NodeName
	updates.CreatedAt = existingRule.CreatedAt

	if err := alertsStorage.UpdateRule(&updates); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updates)
}

// DELETE /api/alerts/:nodename/:ruleid
func DeleteAlertRuleHandler(c *gin.Context) {
	ruleID, err := uuid.Parse(c.Param("ruleid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid rule ID"})
		return
	}

	// Check if there are active alerts for this rule
	rule, err := alertsStorage.GetRuleByID(ruleID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Rule not found"})
		return
	}
	
	activeAlert, err := alertsStorage.GetActiveAlertByRule(ruleID, rule.Target, rule.Metric)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check active alerts"})
		return
	}
	
	if activeAlert != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Cannot delete rule with active alert. Disable the rule or wait for alert to resolve."})
		return
	}

	if err := alertsStorage.DeleteRule(ruleID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Rule deleted successfully"})
}

// PUT /api/alerts/dismiss/:alertid - Dismiss une alerte active manuellement
func DismissAlertHandler(c *gin.Context) {
	alertID, err := uuid.Parse(c.Param("alertid"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid alert ID"})
		return
	}

	// Check that alert exists and is active
	var alert alerts.Alert
	if err := alertsStorage.GetDB().Preload("Rule").Where("id = ? AND status = ?", alertID, alerts.StatusFiring).First(&alert).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Active alert not found"})
		return
	}

	// Send Discord notification for manual dismissal (always send, bypass rate limiting)
	if discord := getDiscordNotifier(); discord != nil {
		// Get current metric value for the notification
		currentValue := alert.CurrentValue // Use stored value as fallback
		if err := discord.SendAlertDismissed(alert.Rule, &alert, currentValue); err != nil {
			fmt.Printf("[ALERT ERROR] Failed to send dismissal notification: %v\n", err)
		}
	}

	// Mark as manually resolved
	if err := alertsStorage.ResolveAlert(alertID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Alert dismissed successfully"})
}

// GET /api/alerts/:nodename/history?days=7
func GetAlertHistoryHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	days, _ := strconv.Atoi(c.DefaultQuery("days", "7"))

	history, err := alertsStorage.GetAlertHistory(nodeName, days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"alerts": history})
}

// GET /alerts/:nodename - Page principale
func AlertsPageHandler(c *gin.Context) {
	nodeName := c.Param("nodename")
	if nodeName == "" {
		c.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Node name required"})
		return
	}

	// Get pods for dropdown
	var podNames []string
	if nodeStats, found := storage.GlobalStore.GetNodeStats(nodeName); found {
		for _, pod := range nodeStats.Metrics.Pods {
			podNames = append(podNames, pod.Name)
		}
	}

	c.HTML(http.StatusOK, "alerts.html", gin.H{
		"NodeName": nodeName,
		"Pods":     podNames,
	})
}
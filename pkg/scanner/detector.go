package scanner

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/devports/devpt/pkg/models"
)

// AgentDetector identifies servers likely started by AI agents
type AgentDetector struct {
	knownAgents map[string]string
}

// NewAgentDetector creates a new agent detector
func NewAgentDetector() *AgentDetector {
	return &AgentDetector{
		knownAgents: map[string]string{
			"opencode": "opencode",
			"cursor":   "cursor",
			"claude":   "claude",
			"gemini":   "gemini",
			"copilot":  "copilot",
		},
	}
}

// DetectAgent analyzes a process and returns an AgentTag if detected
func (ad *AgentDetector) DetectAgent(record *models.ProcessRecord) *models.AgentTag {
	// Check parent process name
	if agentName := ad.checkParentProcess(record.PPID); agentName != "" {
		return &models.AgentTag{
			Source:     models.SourceAgent,
			AgentName:  agentName,
			Confidence: models.ConfidenceHigh,
		}
	}

	// Check for TTY
	if ad.hasTTY(record.PID) {
		return nil // Has TTY, likely manual
	}

	// Check environment variables set by agents
	if ad.hasAgentEnvVars(record.PID) {
		return &models.AgentTag{
			Source:     models.SourceAgent,
			AgentName:  "unknown",
			Confidence: models.ConfidenceMedium,
		}
	}

	// Check for no TTY + typical agent characteristics
	if !ad.hasTTY(record.PID) && ad.isLikelyAgentProcess(record) {
		return &models.AgentTag{
			Source:     models.SourceAgent,
			AgentName:  "unknown",
			Confidence: models.ConfidenceLow,
		}
	}

	return nil
}

// checkParentProcess checks if parent process is a known agent
func (ad *AgentDetector) checkParentProcess(ppid int) string {
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", ppid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	parentName := strings.TrimSpace(string(output))
	for key, agentName := range ad.knownAgents {
		if strings.Contains(parentName, key) {
			return agentName
		}
	}

	return ""
}

// hasTTY checks if process has attached TTY
func (ad *AgentDetector) hasTTY(pid int) bool {
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-o", "tty=")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	tty := strings.TrimSpace(string(output))
	return tty != "" && tty != "?"
}

// hasAgentEnvVars checks for environment variables commonly set by agents
func (ad *AgentDetector) hasAgentEnvVars(pid int) bool {
	// Try to read environment from /proc or ps
	cmd := exec.Command("ps", "-p", fmt.Sprintf("%d", pid), "-e", "-o", "environ=")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	env := string(output)
	agentEnvVars := []string{
		"OPENCODE_",
		"CURSOR_",
		"CLAUDE_",
		"GEMINI_",
		"COPILOT_",
		"AI_AGENT_",
	}

	for _, envVar := range agentEnvVars {
		if strings.Contains(env, envVar) {
			return true
		}
	}

	return false
}

// isLikelyAgentProcess checks if process has typical agent characteristics
func (ad *AgentDetector) isLikelyAgentProcess(record *models.ProcessRecord) bool {
	// Typical agent characteristics:
	// - process doesn't have TTY
	// - common dev server commands (npm, python, go run, etc.)
	// - short-lived parent

	agentKeywords := []string{
		"node",
		"python",
		"ruby",
		"php",
		"go run",
		"npm",
		"yarn",
		"pnpm",
	}

	for _, keyword := range agentKeywords {
		if strings.Contains(record.Command, keyword) {
			return true
		}
	}

	return false
}

// EnrichProcessRecord adds agent detection to a process record
func (ad *AgentDetector) EnrichProcessRecord(record *models.ProcessRecord) {
	if tag := ad.DetectAgent(record); tag != nil {
		record.AgentTag = tag
	}
}

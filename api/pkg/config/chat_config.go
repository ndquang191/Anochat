package config

import (
	"strings"
)

// ChatCategory represents different chat categories
type ChatCategory string

const (
	CategoryPolite ChatCategory = "polite"
)

// ChatCategories returns all available chat categories
func ChatCategories() []ChatCategory {
	return []ChatCategory{
		CategoryPolite,
	}
}

// IsValidCategory checks if a category is valid
func IsValidCategory(category string) bool {
	return category == string(CategoryPolite)
}

// GetDefaultCategory returns the default chat category
func GetDefaultCategory() ChatCategory {
	return CategoryPolite
}

// SensitiveKeywords contains words that trigger sensitive content detection
var SensitiveKeywords = []string{
	"hà nội",
	"bikini",
}

// MessageAnalyzer analyzes messages for sensitive content
type MessageAnalyzer struct{}

// NewMessageAnalyzer creates a new message analyzer
func NewMessageAnalyzer() *MessageAnalyzer {
	return &MessageAnalyzer{}
}

// AnalyzeMessage analyzes a message for sensitive content
func (ma *MessageAnalyzer) AnalyzeMessage(content string) *MessageAnalysis {
	analysis := &MessageAnalysis{
		IsSensitive: false,
		Content:     content,
	}

	// Convert to lowercase for keyword matching
	lowerContent := strings.ToLower(content)

	// Check for sensitive keywords
	for _, keyword := range SensitiveKeywords {
		if strings.Contains(lowerContent, keyword) {
			analysis.IsSensitive = true
			break
		}
	}

	return analysis
}

// MessageAnalysis represents the result of message analysis
type MessageAnalysis struct {
	IsSensitive bool   `json:"is_sensitive"`
	Content     string `json:"content"`
}

// ShouldRetainRoom determines if a room should be retained based on sensitivity
func (ma *MessageAnalyzer) ShouldRetainRoom(analyses []*MessageAnalysis) bool {
	for _, analysis := range analyses {
		if analysis.IsSensitive {
			return true
		}
	}
	return false
}

// ChatSettings contains the main configuration
var ChatSettings = struct {
	Categories        []string
	SensitiveKeywords []string
	MatchTimeoutSec   int
}{
	Categories:        []string{"polite"},
	SensitiveKeywords: []string{"hà nội", "bikini"},
	MatchTimeoutSec:   30,
}

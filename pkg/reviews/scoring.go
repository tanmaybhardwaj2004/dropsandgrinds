package reviews

import (
	"fmt"
	"math"
)

// Weight represents the weight for each review source
type Weight struct {
	Source ReviewSource
	Weight float64
}

// DefaultWeights defines the default weights for each source
var DefaultWeights = []Weight{
	{Source: SourceMetacritic, Weight: 0.25},
	{Source: SourceOpenCritic, Weight: 0.25},
	{Source: SourceSteam, Weight: 0.30},
	{Source: SourceIGN, Weight: 0.10},
	{Source: SourceGameSpot, Weight: 0.10},
}

// NormalizeScore normalizes a score to 0-100 range
func NormalizeScore(score int, min, max int) int {
	if max == min {
		return 50 // Return middle value if range is zero
	}

	normalized := float64(score-min) / float64(max-min) * 100
	return int(math.Round(normalized))
}

// CalculateWeightedAverage calculates the weighted average of review scores
// with proportional redistribution if a source is unavailable
func CalculateWeightedAverage(scores []ReviewScore, weights []Weight) (int, string, error) {
	if len(scores) < 2 {
		return 0, "not enough reviews", fmt.Errorf("require at least 2 review sources")
	}

	// Create a map of available scores
	scoreMap := make(map[ReviewSource]int)
	for _, score := range scores {
		scoreMap[score.Source] = score.Score
	}

	// Calculate total weight of available sources
	var totalWeight float64
	for _, w := range weights {
		if _, exists := scoreMap[w.Source]; exists {
			totalWeight += w.Weight
		}
	}

	if totalWeight == 0 {
		return 0, "no valid sources", fmt.Errorf("no valid review sources available")
	}

	// Calculate weighted sum with proportional redistribution
	var weightedSum float64
	for _, w := range weights {
		if score, exists := scoreMap[w.Source]; exists {
			// Redistribute weight proportionally
			redistributedWeight := w.Weight / totalWeight
			weightedSum += float64(score) * redistributedWeight
		}
	}

	average := int(math.Round(weightedSum))

	// Clamp to 0-100
	if average < 0 {
		average = 0
	} else if average > 100 {
		average = 100
	}

	return average, "", nil
}

// GetScoreColor returns a color code based on the score
func GetScoreColor(score int) string {
	if score >= 85 {
		return "green" // Excellent
	} else if score >= 70 {
		return "amber" // Good
	} else if score >= 50 {
		return "orange" // Average
	}
	return "red" // Poor
}

// GetScoreLabel returns a label based on the score
func GetScoreLabel(score int) string {
	if score >= 90 {
		return "Masterpiece"
	} else if score >= 85 {
		return "Excellent"
	} else if score >= 75 {
		return "Great"
	} else if score >= 65 {
		return "Good"
	} else if score >= 50 {
		return "Average"
	} else if score >= 30 {
		return "Poor"
	}
	return "Terrible"
}

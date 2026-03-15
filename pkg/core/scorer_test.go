package core

import (
	"testing"

	"llm-detector/pkg/fingerprints"
)

func TestNewScorer(t *testing.T) {
	weights := DefaultWeights()
	scorer, err := NewScorer(weights, ModeStrict)
	if err != nil {
		t.Fatalf("Failed to create scorer: %v", err)
	}

	if scorer.GetThreshold() != 0.7 {
		t.Errorf("Expected threshold 0.7, got %f", scorer.GetThreshold())
	}

	// Test loose mode
	scorerLoose, err := NewScorer(weights, ModeLoose)
	if err != nil {
		t.Fatalf("Failed to create loose scorer: %v", err)
	}

	if scorerLoose.GetThreshold() != 0.5 {
		t.Errorf("Expected threshold 0.5 for loose mode, got %f", scorerLoose.GetThreshold())
	}
}

func TestScorerInvalidWeights(t *testing.T) {
	weights := ScoringWeights{
		HeaderMatch:   0.5,
		BodyKeywords:  0.5,
		JSONStructure: 0.5, // Sum = 1.5, invalid
	}

	_, err := NewScorer(weights, ModeStrict)
	if err == nil {
		t.Error("Expected error for invalid weights, got nil")
	}
}

func TestScoreFramework(t *testing.T) {
	weights := DefaultWeights()
	scorer, _ := NewScorer(weights, ModeStrict)

	// Create a test framework fingerprint
	fp := &fingerprints.FrameworkFingerprint{
		Name:        "TestFramework",
		Type:        "test",
		Description: "Test framework",
		Endpoints: []fingerprints.EnhancedEndpoint{
			{Path: "/api/test", Method: "GET", Description: "Test endpoint"},
		},
		Headers: []fingerprints.EnhancedHeaderPattern{
			{Name: "X-Test-Header", Pattern: "test-.*", Required: false, Weight: 0.25},
		},
		BodyPatterns: []fingerprints.EnhancedBodyPattern{
			{Field: "test_field", Required: false, Weight: 0.25},
		},
		Scoring: fingerprints.ScoringConfig{
			HeaderMatch:   0.30,
			BodyKeywords:  0.40,
			JSONStructure: 0.30,
			Threshold:     0.70,
		},
	}

	// Test with matching API result
	apiResults := []APIResult{
		{
			Type:       "test",
			Endpoint:   "/api/test",
			Available:  true,
			StatusCode: 200,
			Headers:    map[string]string{"X-Test-Header": "test-value"},
			Body:       `{"test_field": "value"}`,
		},
	}

	result := scorer.ScoreFramework(fp, apiResults)

	if result.Score < 0 {
		t.Errorf("Expected positive score, got %f", result.Score)
	}

	if len(result.MatchedRules) == 0 {
		t.Error("Expected matched rules, got none")
	}
}

func TestScoreModel(t *testing.T) {
	weights := DefaultWeights()
	scorer, _ := NewScorer(weights, ModeStrict)

	// Create a test model fingerprint
	fp := &fingerprints.ModelFingerprint{
		Name:        "TestModel",
		Provider:    "TestProvider",
		Type:        "test",
		Description: "Test model",
		Response: fingerprints.ResponseFeatures{
			Headers: []fingerprints.EnhancedHeaderPattern{
				{Name: "X-Model", Pattern: "test-model", Required: false, Weight: 0.25},
			},
			BodyPatterns: []fingerprints.EnhancedBodyPattern{
				{Field: "model", Pattern: "test-.*", Required: false, Weight: 0.25},
			},
		},
		Scoring: fingerprints.ScoringConfig{
			HeaderMatch:   0.30,
			BodyKeywords:  0.40,
			JSONStructure: 0.30,
			Threshold:     0.70,
		},
	}

	// Test with matching API result
	apiResults := []APIResult{
		{
			Type:       "test",
			Endpoint:   "/api/test",
			Available:  true,
			StatusCode: 200,
			Headers:    map[string]string{"X-Model": "test-model"},
			Body:       `{"model": "test-model-v1"}`,
		},
	}

	result := scorer.ScoreModel(nil, fp, apiResults)

	if result.Score < 0 {
		t.Errorf("Expected positive score, got %f", result.Score)
	}
}

func TestScoringDetails(t *testing.T) {
	weights := DefaultWeights()
	scorer, _ := NewScorer(weights, ModeStrict)

	details := ScoringDetails{
		HeaderScore:  0.8,
		BodyScore:    0.6,
		JSONScore:    0.7,
		HeaderWeight: 0.3,
		BodyWeight:   0.4,
		JSONWeight:   0.3,
	}

	// Calculate expected weighted score
	expectedScore := 0.8*0.3 + 0.6*0.4 + 0.7*0.3
	calculatedScore := details.HeaderScore*details.HeaderWeight +
		details.BodyScore*details.BodyWeight +
		details.JSONScore*details.JSONWeight

	if calculatedScore != expectedScore {
		t.Errorf("Expected score %f, got %f", expectedScore, calculatedScore)
	}

	_ = scorer // Use scorer to avoid unused variable warning
}

func TestRegexCache(t *testing.T) {
	cache := NewRegexCache()

	// Test Match
	if !cache.Match(`test.*`, "testing") {
		t.Error("Expected match for 'testing' with pattern 'test.*'")
	}

	if cache.Match(`test.*`, "hello") {
		t.Error("Expected no match for 'hello' with pattern 'test.*'")
	}

	// Test caching
	if cache.Size() != 1 {
		t.Errorf("Expected cache size 1, got %d", cache.Size())
	}

	// Test second match uses cache
	if !cache.Match(`test.*`, "tester") {
		t.Error("Expected match for 'tester' with pattern 'test.*'")
	}

	if cache.Size() != 1 {
		t.Errorf("Expected cache size still 1, got %d", cache.Size())
	}
}

func TestMatcher(t *testing.T) {
	matcher := NewMatcher()

	// Test MatchHeader
	headers := map[string]string{
		"Content-Type": "application/json",
		"X-Custom":     "custom-value",
	}

	result := matcher.MatchHeader(headers, "Content-Type", `application/.*`, true)
	if !result.Matched {
		t.Error("Expected header to match")
	}

	// Test MatchBody
	bodyResult := matcher.MatchBody(`{"test": "value"}`, `"test":\s*"value"`, true)
	if !bodyResult.Matched {
		t.Error("Expected body to match")
	}

	// Test MatchJSONPath
	jsonResult := matcher.MatchJSONPath(`{"data": {"name": "test"}}`, "data.name")
	if !jsonResult.Matched {
		t.Error("Expected JSON path to match")
	}

	if jsonResult.Data["value"] != "test" {
		t.Errorf("Expected value 'test', got %v", jsonResult.Data["value"])
	}
}

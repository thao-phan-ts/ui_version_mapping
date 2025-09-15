package analyzer

import (
	"context"
	"fmt"
	"strings"

	"github.com/tsocial/ui-version-mapping/pkg/config"
)

// AnalyzerService là service chính cho việc phân tích configs
type AnalyzerService struct {
	configProvider config.ConfigProvider
}

// NewAnalyzerService tạo analyzer service mới
func NewAnalyzerService(provider config.ConfigProvider) *AnalyzerService {
	return &AnalyzerService{
		configProvider: provider,
	}
}

// SearchRelatedConfigs tìm các configs liên quan đến một config ID
func (s *AnalyzerService) SearchRelatedConfigs(ctx context.Context, configID int, leadSource string, folderPath string) ([]config.RelatedConfigResult, error) {
	// Load source config
	sourceConfig, err := s.configProvider.LoadConfig(ctx, configID, leadSource)
	if err != nil {
		return nil, fmt.Errorf("failed to load source config %d: %w", configID, err)
	}

	// Load all configs from path
	allConfigs, err := s.configProvider.LoadConfigs(ctx, folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configs from %s: %w", folderPath, err)
	}

	fmt.Printf("Found %d configs in %s\n", len(allConfigs), folderPath)

	var results []config.RelatedConfigResult
	resultMap := make(map[int]bool)

	// Detect A/B testing variants first
	abVariants := DetectABTestingVariants(sourceConfig, allConfigs)
	var abVariantIDs []int
	for _, variant := range abVariants {
		abVariantIDs = append(abVariantIDs, variant.ConfigID)
	}

	// Get important tags from source config
	sourceTags := make(map[string]string)
	for _, tag := range sourceConfig.Tags {
		sourceTags[tag.Name] = tag.Value
	}

	// Override lead_source if specified
	if leadSource != "" {
		sourceTags["lead_source"] = leadSource
	}

	for _, cfg := range allConfigs {
		if cfg.ID == configID || resultMap[cfg.ID] {
			continue
		}

		var matchedTags []config.Tag
		var matchReason string
		isABTesting := false
		abTestingGroup := ""

		// Check if this is an A/B testing variant
		for _, variant := range abVariants {
			if variant.ConfigID == cfg.ID {
				isABTesting = true
				abTestingGroup = fmt.Sprintf("A/B Test: %s", cfg.Name)
				matchReason = fmt.Sprintf("A/B Testing variant (Weight: %d, Differences: %s)",
					cfg.Weight, strings.Join(variant.Differences, "; "))
				break
			}
		}

		// Check tag compatibility if not A/B variant
		if !isABTesting && s.isCompatibleByTags(cfg, sourceTags, sourceConfig.Name, &matchedTags, &matchReason) {
			results = append(results, config.RelatedConfigResult{
				ConfigID:       cfg.ID,
				Name:           cfg.Name,
				FlowType:       s.getFlowTypeFromTags(cfg.Tags),
				UIVersion:      cfg.UIVersion,
				Weight:         cfg.Weight,
				MatchReason:    matchReason,
				MatchedTags:    matchedTags,
				IsABTesting:    false,
				ABTestingGroup: "",
				ABVariants:     []int{},
			})
			resultMap[cfg.ID] = true
		} else if isABTesting {
			results = append(results, config.RelatedConfigResult{
				ConfigID:       cfg.ID,
				Name:           cfg.Name,
				FlowType:       s.getFlowTypeFromTags(cfg.Tags),
				UIVersion:      cfg.UIVersion,
				Weight:         cfg.Weight,
				MatchReason:    matchReason,
				MatchedTags:    matchedTags,
				IsABTesting:    true,
				ABTestingGroup: abTestingGroup,
				ABVariants:     abVariantIDs,
			})
			resultMap[cfg.ID] = true
		}
	}

	return results, nil
}

// FindABTestingGroups tìm tất cả A/B testing groups
func (s *AnalyzerService) FindABTestingGroups(ctx context.Context, folderPath string) ([]ABTestingGroup, error) {
	allConfigs, err := s.configProvider.LoadConfigs(ctx, folderPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load configs: %w", err)
	}

	return FindAllABTestingGroups(allConfigs), nil
}

// isCompatibleByTags kiểm tra tính tương thích của tags
func (s *AnalyzerService) isCompatibleByTags(cfg *config.LenderConfig, sourceTags map[string]string, sourceName string, matchedTags *[]config.Tag, matchReason *string) bool {
	// Exclude configs with same name
	if cfg.Name == sourceName {
		return false
	}

	var configTags = make(map[string]string)
	var configLeadSources, configTelcoCodes []string

	for _, tag := range cfg.Tags {
		configTags[tag.Name] = tag.Value
		switch tag.Name {
		case "lead_source":
			configLeadSources = append(configLeadSources, tag.Value)
		case "telco_code":
			configTelcoCodes = append(configTelcoCodes, tag.Value)
		}
	}

	var matches []config.Tag
	var reasons []string

	// Check product_code match (required)
	if sourceTags["product_code"] != "" && configTags["product_code"] == sourceTags["product_code"] {
		matches = append(matches, config.Tag{Name: "product_code", Value: sourceTags["product_code"]})
		reasons = append(reasons, "same product_code")
	} else if sourceTags["product_code"] != "" {
		return false // No product_code match, exclude
	}

	// Check lead_source match (if specified)
	if sourceTags["lead_source"] != "" {
		hasLeadSourceMatch := false
		for _, configLS := range configLeadSources {
			if configLS == sourceTags["lead_source"] {
				matches = append(matches, config.Tag{Name: "lead_source", Value: sourceTags["lead_source"]})
				reasons = append(reasons, "same lead_source")
				hasLeadSourceMatch = true
				break
			}
		}
		if !hasLeadSourceMatch {
			return false
		}
	}

	// Check telco_code compatibility
	if sourceTags["telco_code"] != "" {
		for _, configTC := range configTelcoCodes {
			if configTC == sourceTags["telco_code"] {
				matches = append(matches, config.Tag{Name: "telco_code", Value: sourceTags["telco_code"]})
				reasons = append(reasons, "shared telco_code: "+sourceTags["telco_code"])
				break
			}
		}
	}

	// Check flow_type
	sourceFlowType := s.getFlowTypeFromSourceTags(sourceTags)
	configFlowType := s.getFlowTypeFromTags(cfg.Tags)

	if sourceFlowType != "" && configFlowType != "" {
		if configFlowType == sourceFlowType {
			matches = append(matches, config.Tag{Name: "flow_type", Value: configFlowType})
			reasons = append(reasons, "same flow_type")
		} else {
			matches = append(matches, config.Tag{Name: "flow_type", Value: configFlowType})
			reasons = append(reasons, "different flow_type: "+configFlowType)
		}
	}

	*matchedTags = matches
	*matchReason = strings.Join(reasons, ", ")

	return len(matches) >= 1
}

// getFlowTypeFromTags gets flow_type from tags (prioritizes esign_flow_type first)
func (s *AnalyzerService) getFlowTypeFromTags(tags []config.Tag) string {
	// Prioritize esign_flow_type first
	for _, tag := range tags {
		if tag.Name == "esign_flow_type" {
			return tag.Value
		}
	}

	// If no esign_flow_type, find flow_type
	for _, tag := range tags {
		if tag.Name == "flow_type" {
			return tag.Value
		}
	}

	return "unknown"
}

// getFlowTypeFromSourceTags gets flow_type from source tags map
func (s *AnalyzerService) getFlowTypeFromSourceTags(sourceTags map[string]string) string {
	// Prioritize esign_flow_type first
	if esignFlowType, exists := sourceTags["esign_flow_type"]; exists && esignFlowType != "" {
		return esignFlowType
	}

	// If no esign_flow_type, find flow_type
	if flowType, exists := sourceTags["flow_type"]; exists && flowType != "" {
		return flowType
	}

	return ""
}

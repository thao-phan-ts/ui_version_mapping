package analyzer

import (
	"fmt"

	"github.com/tsocial/ui-version-mapping/pkg/config"
)

// ABTestingVariant represents an A/B testing variant
type ABTestingVariant struct {
	ConfigID    int      `json:"config_id"`
	Name        string   `json:"name"`
	Weight      int      `json:"weight"`
	UIFlow      []string `json:"ui_flow"`
	Differences []string `json:"differences"`
}

// ABTestingGroup represents a group of A/B testing variants
type ABTestingGroup struct {
	GroupName   string             `json:"group_name"`
	Variants    []ABTestingVariant `json:"variants"`
	TotalWeight int                `json:"total_weight"`
}

// ABTestingAnalysisResult represents the complete A/B testing analysis result
type ABTestingAnalysisResult struct {
	SearchID        int                          `json:"search_id"`
	SearchType      string                       `json:"search_type"`
	ABTestingGroups []ABTestingGroup             `json:"ab_testing_groups"`
	NormalResults   []config.RelatedConfigResult `json:"normal_results"`
	TotalResults    int                          `json:"total_results"`
}

// DetectABTestingVariants finds A/B testing variants of a config
func DetectABTestingVariants(sourceConfig *config.LenderConfig, allConfigs []*config.LenderConfig) []ABTestingVariant {
	var variants []ABTestingVariant

	for _, cfg := range allConfigs {
		if cfg.ID == sourceConfig.ID {
			continue
		}

		// Check if configs have same basic conditions but different UI flows
		if IsABTestingVariant(sourceConfig, cfg) {
			differences := FindUIFlowDifferences(sourceConfig.UIFlow, cfg.UIFlow)
			variants = append(variants, ABTestingVariant{
				ConfigID:    cfg.ID,
				Name:        cfg.Name,
				Weight:      cfg.Weight,
				UIFlow:      cfg.UIFlow,
				Differences: differences,
			})
		}
	}

	return variants
}

// IsABTestingVariant checks if 2 configs are A/B testing variants
func IsABTestingVariant(config1, config2 *config.LenderConfig) bool {
	// 1. Must have same name (same business logic)
	if config1.Name != config2.Name {
		return false
	}

	// 2. Must have same basic tags
	if !HasSameBasicTags(config1, config2) {
		return false
	}

	// 3. Must have different UI flows (this is the A/B test point)
	if AreUIFlowsIdentical(config1.UIFlow, config2.UIFlow) {
		return false
	}

	// 4. Usually have weight > 0 (for traffic distribution)
	if config1.Weight <= 0 || config2.Weight <= 0 {
		return false
	}

	return true
}

// HasSameBasicTags checks if 2 configs have same basic tags
func HasSameBasicTags(config1, config2 *config.LenderConfig) bool {
	tags1 := make(map[string][]string)
	tags2 := make(map[string][]string)

	// Collect tags from config1
	for _, tag := range config1.Tags {
		tags1[tag.Name] = append(tags1[tag.Name], tag.Value)
	}

	// Collect tags from config2
	for _, tag := range config2.Tags {
		tags2[tag.Name] = append(tags2[tag.Name], tag.Value)
	}

	// Check critical tags
	criticalTags := []string{"product_code", "lead_source", "telco_code", "flow_type", "esign_flow_type"}

	for _, tagName := range criticalTags {
		if !AreTagValuesEqual(tags1[tagName], tags2[tagName]) {
			// Special case for flow_type: if one config has esign_flow_type and the other has flow_type, consider equivalent
			if tagName == "flow_type" || tagName == "esign_flow_type" {
				// Get flow_type from both configs
				flowType1 := GetFlowTypeFromTagsMap(tags1)
				flowType2 := GetFlowTypeFromTagsMap(tags2)
				if flowType1 != flowType2 {
					return false
				}
			} else {
				return false
			}
		}
	}

	return true
}

// AreTagValuesEqual checks if 2 string slices are equal (ignoring order)
func AreTagValuesEqual(slice1, slice2 []string) bool {
	if len(slice1) != len(slice2) {
		return false
	}

	map1 := make(map[string]int)
	map2 := make(map[string]int)

	for _, v := range slice1 {
		map1[v]++
	}

	for _, v := range slice2 {
		map2[v]++
	}

	for k, v := range map1 {
		if map2[k] != v {
			return false
		}
	}

	return true
}

// AreUIFlowsIdentical checks if 2 UI flows are identical
func AreUIFlowsIdentical(flow1, flow2 []string) bool {
	if len(flow1) != len(flow2) {
		return false
	}

	for i, step := range flow1 {
		if step != flow2[i] {
			return false
		}
	}

	return true
}

// FindUIFlowDifferences finds differences between 2 UI flows
func FindUIFlowDifferences(flow1, flow2 []string) []string {
	var differences []string

	// Find steps that are different
	maxLen := len(flow1)
	if len(flow2) > maxLen {
		maxLen = len(flow2)
	}

	for i := 0; i < maxLen; i++ {
		var step1, step2 string

		if i < len(flow1) {
			step1 = flow1[i]
		}

		if i < len(flow2) {
			step2 = flow2[i]
		}

		if step1 != step2 {
			if step1 != "" && step2 != "" {
				differences = append(differences, fmt.Sprintf("Step %d: %s vs %s", i+1, step1, step2))
			} else if step1 != "" {
				differences = append(differences, fmt.Sprintf("Step %d: %s (missing in variant)", i+1, step1))
			} else {
				differences = append(differences, fmt.Sprintf("Step %d: %s (extra in variant)", i+1, step2))
			}
		}
	}

	return differences
}

// GetFlowTypeFromTagsMap gets flow_type from tags map (prioritizes esign_flow_type first)
func GetFlowTypeFromTagsMap(tagsMap map[string][]string) string {
	// Prioritize esign_flow_type first
	if esignFlowTypes, exists := tagsMap["esign_flow_type"]; exists && len(esignFlowTypes) > 0 {
		return esignFlowTypes[0]
	}

	// If no esign_flow_type, find flow_type
	if flowTypes, exists := tagsMap["flow_type"]; exists && len(flowTypes) > 0 {
		return flowTypes[0]
	}

	return "unknown"
}

// FindAllABTestingGroups finds all A/B testing groups in a set of configs
func FindAllABTestingGroups(allConfigs []*config.LenderConfig) []ABTestingGroup {
	var groups []ABTestingGroup
	processedConfigs := make(map[int]bool)

	for _, cfg := range allConfigs {
		if processedConfigs[cfg.ID] {
			continue
		}

		variants := DetectABTestingVariants(cfg, allConfigs)
		if len(variants) > 0 {
			// Create A/B testing group
			group := ABTestingGroup{
				GroupName:   cfg.Name,
				TotalWeight: cfg.Weight,
			}

			// Add source config as first variant
			sourceVariant := ABTestingVariant{
				ConfigID:    cfg.ID,
				Name:        cfg.Name,
				Weight:      cfg.Weight,
				UIFlow:      cfg.UIFlow,
				Differences: []string{"Original variant"},
			}
			group.Variants = append(group.Variants, sourceVariant)
			processedConfigs[cfg.ID] = true

			// Add all variants
			for _, variant := range variants {
				group.Variants = append(group.Variants, variant)
				group.TotalWeight += variant.Weight
				processedConfigs[variant.ConfigID] = true
			}

			groups = append(groups, group)
		}
	}

	return groups
}

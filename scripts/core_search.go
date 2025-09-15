package ui_version_check

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// Global path constants - configurable paths for lender configs
const (
	DJLenderConfigsPath = "submodules/digital_journey/migration/sync/vietnam/tpbank/lender_configs"
	ProjectDir          = "../"

	// Test results directory structure
	TestResultsBaseDir   = "out/test_results"
	TestResultsPumlDir   = "pumls"
	TestResultsImagesDir = "images"
)

// Helper functions to build paths with lender config ID
func GetConfigResultsDir(lenderConfigID int) string {
	return filepath.Join(TestResultsBaseDir, fmt.Sprintf("%d", lenderConfigID))
}

func GetConfigPumlDir(lenderConfigID int) string {
	return filepath.Join(GetConfigResultsDir(lenderConfigID), TestResultsPumlDir)
}

func GetConfigImagesDir(lenderConfigID int) string {
	return filepath.Join(GetConfigResultsDir(lenderConfigID), TestResultsImagesDir)
}

// Global SearchType constants
const (
	SearchTypeABTestingAnalysis              = "ab_testing_analysis"
	SearchTypeUIVersionAnalysis              = "ui_version_analysis"
	SearchTypeUserOnboardingWorkflowAnalysis = "user_onboarding_workflow_analysis"
	SearchTypeUserDropOffAnalysis            = "user_drop_off_analysis"
)

// ValidSearchTypes returns all valid SearchType constants
func ValidSearchTypes() []string {
	return []string{
		SearchTypeABTestingAnalysis,
		SearchTypeUIVersionAnalysis,
		SearchTypeUserOnboardingWorkflowAnalysis,
		SearchTypeUserDropOffAnalysis,
	}
}

// IsValidSearchType checks if a SearchType is valid
func IsValidSearchType(searchType string) bool {
	validTypes := ValidSearchTypes()
	for _, validType := range validTypes {
		if searchType == validType {
			return true
		}
	}
	return false
}

// CheckFile ensures the directory exists for a given filename
func CheckFile(filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}

// ListFilesContainingKeyword searches for files containing a specific keyword (ID)
func ListFilesContainingKeyword(path string, keyword int) [][]string {
	var matchingFiles [][]string

	err := filepath.Walk(path, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "archive") {
			return filepath.SkipDir
		}

		if !info.IsDir() && strings.Contains(info.Name(), fmt.Sprintf("%d", keyword)) {
			relPath, err := filepath.Rel(path, filePath)
			if err != nil {
				matchingFiles = append(matchingFiles, []string{info.Name(), filePath})
			} else {
				dirPath := filepath.Dir(relPath)
				if dirPath == "." {
					matchingFiles = append(matchingFiles, []string{info.Name(), path})
				} else {
					fullPath := filepath.Join(path, dirPath)
					matchingFiles = append(matchingFiles, []string{info.Name(), fullPath})
				}
			}
		}

		return nil
	})

	if err != nil {
		return [][]string{}
	}

	return matchingFiles
}

// SearchLenderConfigID finds the file name and path for a given lender config ID
func SearchLenderConfigID(lenderConfigID int) (string, string) {
	listFiles := ListFilesContainingKeyword(DJLenderConfigsPath, lenderConfigID)
	if len(listFiles) == 0 {
		fmt.Printf("Warning: No files found for Lender Config ID %d.\n", lenderConfigID)
		return "", ""
	}
	if len(listFiles) > 1 {
		fmt.Printf("Warning: Multiple files found for Lender Config ID %d. Using the first match.\n", lenderConfigID)
	}
	name := listFiles[0][0]
	path := listFiles[0][1]

	return name, path
}

// WriteSearchResultToJSON writes search results to a JSON file
func WriteSearchResultToJSON(result SearchResult, filename string) error {
	if err := CheckFile(filename); err != nil {
		return fmt.Errorf("failed to prepare file path: %w", err)
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal SearchResult to JSON: %w", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file %s: %w", filename, err)
	}

	fmt.Printf("Successfully wrote SearchResult to %s\n", filename)
	return nil
}

// ReadLenderConfig reads and parses a lender configuration file
func ReadLenderConfig(path string) (*LenderConfig, error) {
	data, _ := os.ReadFile(path)

	var result *LenderConfig
	err := json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return result, nil
}

// GetAllLenderConfigsFromPath reads all lender configs from a specific folder path
func GetAllLenderConfigsFromPath(folderPath string) []*LenderConfig {
	var configs []*LenderConfig

	// Determine full path
	var fullPath string
	if folderPath == "" {
		// If not specified, scan all
		fullPath = DJLenderConfigsPath
	} else if strings.HasPrefix(folderPath, "submodules/") || filepath.IsAbs(folderPath) {
		// Path is already complete (submodules or absolute)
		fullPath = folderPath
	} else {
		// Use relative path from project root
		fullPath = ProjectDir + folderPath
	}

	fmt.Printf("Scanning configs from path: %s\n", fullPath)

	// Scan specified directory
	err := filepath.Walk(fullPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		// Skip archive directories
		if info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "archive") {
			return filepath.SkipDir
		}

		// Process JSON files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			config, err := ReadLenderConfig(filePath)
			if err == nil && config != nil {
				configs = append(configs, config)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error scanning configs from %s: %v\n", fullPath, err)
	}

	fmt.Printf("Found %d configs in %s\n", len(configs), fullPath)
	return configs
}

// ============================================================================
// FLOW TYPE FUNCTIONS
// ============================================================================

// GetFlowTypeFromTags gets flow_type from tags (prioritizes esign_flow_type first, then flow_type)
func GetFlowTypeFromTags(tags []Tag) string {
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

// GetFlowTypeFromSourceTags gets flow_type from source tags map (prioritizes esign_flow_type)
func GetFlowTypeFromSourceTags(sourceTags map[string]string) string {
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

// ============================================================================
// CORE SEARCH FUNCTIONS
// ============================================================================

// SearchRelatedConfig finds related lender config IDs from an input ID, lead_source and folder path
func SearchRelatedConfig(lenderConfigID int, leadSource string, folderPath string) []int {
	// Read source config
	name, path := SearchLenderConfigID(lenderConfigID)
	if name == "" || path == "" {
		fmt.Printf("Cannot find lender config with ID: %d\n", lenderConfigID)
		return []int{}
	}

	sourceConfig, err := ReadLenderConfig(path + "/" + name)
	if err != nil {
		fmt.Printf("Error reading source config: %v\n", err)
		return []int{}
	}

	var relatedConfigIDs []int
	relatedMap := make(map[int]bool) // To avoid duplicates

	// Only use tag matching logic - this is the most accurate logic (and exclude same name)
	relatedByTags := FindConfigsByTagsWithLeadSourceAndPathAndName(sourceConfig.Tags, leadSource, folderPath, sourceConfig.Name)
	for _, configID := range relatedByTags {
		if configID != lenderConfigID && !relatedMap[configID] {
			relatedConfigIDs = append(relatedConfigIDs, configID)
			relatedMap[configID] = true
		}
	}

	// Remove decision tree and UI flow matching as they create too many unexpected results
	// Only use tag matching for accurate results

	return relatedConfigIDs
}

// FindConfigsByTagsWithLeadSourceAndPathAndName finds compatible configs and excludes same name
func FindConfigsByTagsWithLeadSourceAndPathAndName(sourceTags []Tag, targetLeadSource string, folderPath string, sourceName string) []int {
	var relatedConfigs []int

	// Get important tags for flow routing
	var leadSource, telcoCode, productCode, flowType string
	for _, tag := range sourceTags {
		switch tag.Name {
		case "lead_source":
			leadSource = tag.Value
		case "telco_code":
			telcoCode = tag.Value
		case "product_code":
			productCode = tag.Value
		case "flow_type":
			if flowType == "" { // Only set if not already set from esign_flow_type
				flowType = tag.Value
			}
		case "esign_flow_type":
			flowType = tag.Value // esign_flow_type has higher priority
		}
	}

	// If there's targetLeadSource, prioritize searching by that lead_source
	if targetLeadSource != "" {
		leadSource = targetLeadSource
	}

	// Scan configs in specific folder path
	allConfigs := GetAllLenderConfigsFromPath(folderPath)

	for _, config := range allConfigs {
		// Exclude configs with same name
		if config.Name == sourceName {
			continue
		}

		if IsCompatibleByTagsWithLeadSource(config, leadSource, telcoCode, productCode, flowType) {
			relatedConfigs = append(relatedConfigs, config.ID)
		}
	}

	return relatedConfigs
}

// IsCompatibleByTagsWithLeadSource checks if config has compatible tags with specific lead_source
func IsCompatibleByTagsWithLeadSource(config *LenderConfig, leadSource, telcoCode, productCode, flowType string) bool {
	var configLeadSources, configTelcoCodes []string
	var configProductCode, configFlowType string

	for _, tag := range config.Tags {
		switch tag.Name {
		case "lead_source":
			configLeadSources = append(configLeadSources, tag.Value)
		case "telco_code":
			configTelcoCodes = append(configTelcoCodes, tag.Value)
		case "product_code":
			configProductCode = tag.Value
		case "flow_type":
			if configFlowType == "" { // Only set if not already set from esign_flow_type
				configFlowType = tag.Value
			}
		case "esign_flow_type":
			configFlowType = tag.Value // esign_flow_type has higher priority
		}
	}

	// Search logic:
	// 1. Must have same product_code (required)
	if productCode != "" && configProductCode != productCode {
		return false
	}

	// 2. Must have at least one common lead_source (if leadSource is specified)
	hasLeadSourceMatch := leadSource == ""
	if leadSource != "" {
		for _, configLS := range configLeadSources {
			if configLS == leadSource {
				hasLeadSourceMatch = true
				break
			}
		}
	}
	if !hasLeadSourceMatch {
		return false
	}

	// 3. Must have at least one common telco_code (if telcoCode is specified)
	hasTelcoMatch := telcoCode == ""
	if telcoCode != "" {
		for _, configTC := range configTelcoCodes {
			if configTC == telcoCode {
				hasTelcoMatch = true
				break
			}
		}
	}
	if !hasTelcoMatch {
		return false
	}

	// 4. Accept all flow_type (no longer filter by flow_type)
	return true
}

// SearchRelatedConfigDetailed returns detailed results with match reasons
func SearchRelatedConfigDetailed(lenderConfigID int, leadSource string, folderPath string) []RelatedConfigResult {
	// Read source config
	name, path := SearchLenderConfigID(lenderConfigID)
	if name == "" || path == "" {
		fmt.Printf("Cannot find lender config with ID: %d\n", lenderConfigID)
		return []RelatedConfigResult{}
	}

	sourceConfig, err := ReadLenderConfig(path + "/" + name)
	if err != nil {
		fmt.Printf("Error reading source config: %v\n", err)
		return []RelatedConfigResult{}
	}

	var results []RelatedConfigResult
	resultMap := make(map[int]bool) // To avoid duplicates

	// Only use tag matching - most accurate logic
	allConfigs := GetAllLenderConfigsFromPath(folderPath)

	// Detect A/B testing variants first
	abVariants := DetectABTestingVariants(sourceConfig, allConfigs)
	var abVariantIDs []int
	for _, variant := range abVariants {
		abVariantIDs = append(abVariantIDs, variant.ConfigID)
	}

	// Get important tags from source config
	var sourceTags = make(map[string]string)
	for _, tag := range sourceConfig.Tags {
		sourceTags[tag.Name] = tag.Value
	}

	// If there's leadSource from input, override source lead_source
	if leadSource != "" {
		sourceTags["lead_source"] = leadSource
	}

	for _, config := range allConfigs {
		if config.ID == lenderConfigID || resultMap[config.ID] {
			continue
		}

		var matchedTags []Tag
		var matchReason string
		isABTesting := false
		abTestingGroup := ""

		// Check if this is an A/B testing variant
		for _, variant := range abVariants {
			if variant.ConfigID == config.ID {
				isABTesting = true
				abTestingGroup = fmt.Sprintf("A/B Test: %s", config.Name)
				matchReason = fmt.Sprintf("A/B Testing variant (Weight: %d, Differences: %s)",
					config.Weight, strings.Join(variant.Differences, "; "))
				break
			}
		}

		// Only check tag compatibility if not A/B variant (since A/B variants are already excluded)
		if !isABTesting && IsCompatibleByTagsDetailedWithLeadSourceAndName(config, sourceTags, sourceConfig.Name, &matchedTags, &matchReason) {
			results = append(results, RelatedConfigResult{
				ConfigID:       config.ID,
				Name:           config.Name,
				FlowType:       GetFlowTypeFromTags(config.Tags),
				UIVersion:      config.UIVersion,
				Weight:         config.Weight,
				MatchReason:    matchReason,
				MatchedTags:    matchedTags,
				IsABTesting:    false,
				ABTestingGroup: "",
				ABVariants:     []int{},
			})
			resultMap[config.ID] = true
		} else if isABTesting {
			// Include A/B testing variants as separate entries for information
			results = append(results, RelatedConfigResult{
				ConfigID:       config.ID,
				Name:           config.Name,
				FlowType:       GetFlowTypeFromTags(config.Tags),
				UIVersion:      config.UIVersion,
				Weight:         config.Weight,
				MatchReason:    matchReason,
				MatchedTags:    matchedTags,
				IsABTesting:    true,
				ABTestingGroup: abTestingGroup,
				ABVariants:     abVariantIDs,
			})
			resultMap[config.ID] = true
		}
	}

	return results
}

// IsCompatibleByTagsDetailedWithLeadSource checks tag compatibility with specific lead_source and detailed info
func IsCompatibleByTagsDetailedWithLeadSource(config *LenderConfig, sourceTags map[string]string, matchedTags *[]Tag, matchReason *string) bool {
	var configTags = make(map[string]string)
	var configLeadSources, configTelcoCodes []string

	for _, tag := range config.Tags {
		configTags[tag.Name] = tag.Value
		switch tag.Name {
		case "lead_source":
			configLeadSources = append(configLeadSources, tag.Value)
		case "telco_code":
			configTelcoCodes = append(configTelcoCodes, tag.Value)
		}
	}

	var matches []Tag
	var reasons []string

	// Check product_code match (required)
	if sourceTags["product_code"] != "" && configTags["product_code"] == sourceTags["product_code"] {
		matches = append(matches, Tag{Name: "product_code", Value: sourceTags["product_code"]})
		reasons = append(reasons, "same product_code")
	} else if sourceTags["product_code"] != "" {
		return false // No product_code match, exclude
	}

	// Check lead_source match (if specified)
	if sourceTags["lead_source"] != "" {
		hasLeadSourceMatch := false
		for _, configLS := range configLeadSources {
			if configLS == sourceTags["lead_source"] {
				matches = append(matches, Tag{Name: "lead_source", Value: sourceTags["lead_source"]})
				reasons = append(reasons, "same lead_source")
				hasLeadSourceMatch = true
				break
			}
		}
		if !hasLeadSourceMatch {
			return false
		}
	}

	// Check telco_code compatibility (can have multiple common telco_codes)
	sourceTelcoCodes := []string{}
	if sourceTags["telco_code"] != "" {
		sourceTelcoCodes = append(sourceTelcoCodes, sourceTags["telco_code"])
	}

	// Find common telco_code
	for _, sourceTC := range sourceTelcoCodes {
		for _, configTC := range configTelcoCodes {
			if sourceTC == configTC {
				matches = append(matches, Tag{Name: "telco_code", Value: sourceTC})
				reasons = append(reasons, "shared telco_code: "+sourceTC)
				break
			}
		}
	}

	// Check flow_type (prioritize esign_flow_type, then flow_type)
	sourceFlowType := GetFlowTypeFromSourceTags(sourceTags)
	configFlowType := GetFlowTypeFromTags(config.Tags)

	if sourceFlowType != "" && configFlowType != "" {
		if configFlowType == sourceFlowType {
			matches = append(matches, Tag{Name: "flow_type", Value: configFlowType})
			reasons = append(reasons, "same flow_type")
		} else {
			matches = append(matches, Tag{Name: "flow_type", Value: configFlowType})
			reasons = append(reasons, "different flow_type: "+configFlowType)
		}
	}

	*matchedTags = matches
	*matchReason = strings.Join(reasons, ", ")

	// Improved matching logic:
	// 1. Must have product_code match (already checked above)
	// 2. Must have lead_source match if specified (already checked above)
	// 3. Must have at least 1 match (product_code is sufficient)
	return len(matches) >= 1
}

// IsCompatibleByTagsDetailedWithLeadSourceAndName checks tag compatibility and excludes same name
func IsCompatibleByTagsDetailedWithLeadSourceAndName(config *LenderConfig, sourceTags map[string]string, sourceName string, matchedTags *[]Tag, matchReason *string) bool {
	// Exclude configs with same name (may be duplicates or different versions)
	if config.Name == sourceName {
		return false
	}

	return IsCompatibleByTagsDetailedWithLeadSource(config, sourceTags, matchedTags, matchReason)
}

// ============================================================================
// A/B TESTING FUNCTIONS
// ============================================================================

// DetectABTestingVariants finds A/B testing variants of a config
func DetectABTestingVariants(sourceConfig *LenderConfig, allConfigs []*LenderConfig) []ABTestingVariant {
	var variants []ABTestingVariant

	for _, config := range allConfigs {
		if config.ID == sourceConfig.ID {
			continue
		}

		// Check if configs have same basic conditions but different UI flows
		if IsABTestingVariant(sourceConfig, config) {
			differences := FindUIFlowDifferences(sourceConfig.UIFlow, config.UIFlow)
			variants = append(variants, ABTestingVariant{
				ConfigID:    config.ID,
				Name:        config.Name,
				Weight:      config.Weight,
				UIFlow:      config.UIFlow,
				Differences: differences,
			})
		}
	}

	return variants
}

// IsABTestingVariant checks if 2 configs are A/B testing variants
func IsABTestingVariant(config1, config2 *LenderConfig) bool {
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
func HasSameBasicTags(config1, config2 *LenderConfig) bool {
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

// FindAllABTestingGroups finds all A/B testing groups in folder path
func FindAllABTestingGroups(folderPath string) []ABTestingGroup {
	allConfigs := GetAllLenderConfigsFromPath(folderPath)
	var groups []ABTestingGroup
	processedConfigs := make(map[int]bool)

	for _, config := range allConfigs {
		if processedConfigs[config.ID] {
			continue
		}

		variants := DetectABTestingVariants(config, allConfigs)
		if len(variants) > 0 {
			// Create A/B testing group
			group := ABTestingGroup{
				GroupName:   config.Name,
				TotalWeight: config.Weight,
			}

			// Add source config as first variant
			sourceVariant := ABTestingVariant{
				ConfigID:    config.ID,
				Name:        config.Name,
				Weight:      config.Weight,
				UIFlow:      config.UIFlow,
				Differences: []string{"Original variant"},
			}
			group.Variants = append(group.Variants, sourceVariant)
			processedConfigs[config.ID] = true

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

// ============================================================================
// VISUALIZATION FUNCTIONS
// ============================================================================

// GenerateABTestingDiagram creates PlantUML diagram for A/B testing groups
func GenerateABTestingDiagram(groups []ABTestingGroup, filename string) error {
	var puml strings.Builder

	puml.WriteString("@startuml\n")
	puml.WriteString("title A/B Testing Groups Analysis\n\n")

	for i, group := range groups {
		puml.WriteString(fmt.Sprintf("package \"Group %d: %s\" {\n", i+1, group.GroupName))

		for j, variant := range groups[i].Variants {
			percentage := float64(variant.Weight) / float64(group.TotalWeight) * 100
			puml.WriteString(fmt.Sprintf("  rectangle \"Config %d\\nWeight: %d (%.1f%%)\" as config_%d_%d\n",
				variant.ConfigID, variant.Weight, percentage, i, j))
		}

		puml.WriteString("}\n\n")
	}

	puml.WriteString("@enduml\n")

	// Write to file
	if err := CheckFile(filename); err != nil {
		return fmt.Errorf("failed to prepare file path: %w", err)
	}

	err := os.WriteFile(filename, []byte(puml.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PlantUML file %s: %w", filename, err)
	}

	fmt.Printf("A/B Testing PlantUML diagram written to %s\n", filename)
	return nil
}

// ExportPlantUMLToPNG converts a PlantUML file to PNG using plantuml.jar
func ExportPlantUMLToPNG(pumlFilename string) error {
	// Look for plantuml.jar in the ui_version_check directory
	jarPaths := []string{
		"../plantuml.jar",
		"plantuml.jar",
		"../../plantuml.jar",
		"../../../plantuml.jar",
	}

	var jarPath string
	for _, path := range jarPaths {
		if _, err := os.Stat(path); err == nil {
			jarPath = path
			break
		}
	}

	if jarPath == "" {
		return fmt.Errorf("plantuml.jar not found in expected locations: %s", strings.Join(jarPaths, ", "))
	}

	// Generate PNG filename by replacing .puml with .png
	pngFilename := strings.Replace(pumlFilename, ".puml", ".png", 1)

	// Create the directory for PNG output if it doesn't exist
	pngDir := filepath.Dir(pngFilename)
	if err := os.MkdirAll(pngDir, 0755); err != nil {
		return fmt.Errorf("failed to create PNG output directory: %w", err)
	}

	// Check if Java is available
	if _, err := exec.LookPath("java"); err != nil {
		return fmt.Errorf("java not found in PATH, please install Java to export PNG diagrams")
	}

	// Run PlantUML command to convert to PNG
	// java -jar plantuml.jar -tpng input.puml -o output_dir
	args := []string{"-jar", jarPath, "-tpng", pumlFilename, "-o", pngDir}

	fmt.Printf("Converting PlantUML to PNG: java %s\n", strings.Join(args, " "))

	// Execute the command
	execCmd := exec.Command("java", args...)

	output, err := execCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to convert PlantUML to PNG: %w\nOutput: %s", err, string(output))
	}

	fmt.Printf("PNG diagram exported to %s\n", pngFilename)
	return nil
}

// ============================================================================
// JOURNEY GENERATION FUNCTIONS
// ============================================================================

// JourneyTemplate represents the template structure for journey generation
type JourneyTemplate struct {
	SearchValue      int64     `json:"search_value"`
	SearchType       string    `json:"search_type"`
	RelatedConfigIDs []int     `json:"related_config_ids"`
	Journeys         []Journey `json:"journeys"`
}

// ABTestingAnalysisResult represents the complete A/B testing analysis result
type ABTestingAnalysisResult struct {
	SearchID        int                   `json:"search_id"`
	SearchType      string                `json:"search_type"`
	ABTestingGroups []ABTestingGroup      `json:"ab_testing_groups"`
	NormalResults   []RelatedConfigResult `json:"normal_results"`
	TotalResults    int                   `json:"total_results"`
}

// GenerateJourneyFromTemplate creates a journey section based on template data
func GenerateJourneyFromTemplate(sourceConfigID int, targetConfigID int, flowType string, condition string, description string, steps []Step) Journey {
	journeyID := fmt.Sprintf("from_%d_to_%d", sourceConfigID, targetConfigID)

	return Journey{
		ID:                 journeyID,
		FlowType:           flowType,
		FromLenderConfigID: sourceConfigID,
		ToLenderConfigID:   targetConfigID,
		Active:             true,
		Condition:          condition,
		Description:        description,
		Steps:              steps,
	}
}

// GenerateStandardJourneySteps creates standard journey steps based on UI flow
func GenerateStandardJourneySteps(uiFlow []string, mainUIVersion string) []Step {
	var steps []Step

	for i, stepName := range uiFlow {
		step := Step{
			ID:                       i,
			Name:                     stepName,
			MainUIVersion:            mainUIVersion,
			SubUIVersion:             "",
			SubUIVersionByConditions: []SubUIVersionByCondition{},
		}
		steps = append(steps, step)
	}

	return steps
}

// GenerateFullJourneySteps creates complete journey steps combining source and target flows
func GenerateFullJourneySteps(sourceConfig, targetConfig *LenderConfig, flowType string) []Step {
	var steps []Step
	stepID := 0

	// For normal flow (self-loop), just use source config steps
	if sourceConfig.ID == targetConfig.ID {
		return GenerateStandardJourneySteps(sourceConfig.UIFlow, sourceConfig.UIVersion)
	}

	// Add initial steps from source config (common steps)
	commonSteps := []string{"otp", "app_form.basic_info"}

	// For rejection flows, add minimal steps
	if strings.Contains(flowType, "rejection") {
		for _, stepName := range commonSteps {
			if stepID < len(sourceConfig.UIFlow) && sourceConfig.UIFlow[stepID] == stepName {
				steps = append(steps, Step{
					ID:                       stepID,
					Name:                     stepName,
					MainUIVersion:            sourceConfig.UIVersion,
					SubUIVersion:             "",
					SubUIVersionByConditions: []SubUIVersionByCondition{},
				})
				stepID++
			}
		}
		// Add rejection-specific steps
		steps = append(steps, Step{
			ID:                       stepID,
			Name:                     "ekyc.selfie.flash",
			MainUIVersion:            targetConfig.UIVersion,
			SubUIVersion:             "",
			SubUIVersionByConditions: []SubUIVersionByCondition{},
		})
		stepID++
		steps = append(steps, Step{
			ID:                       stepID,
			Name:                     "failure",
			MainUIVersion:            targetConfig.UIVersion,
			SubUIVersion:             "",
			SubUIVersionByConditions: []SubUIVersionByCondition{},
		})
		return steps
	}

	// For automated flows (auto_pcb, auto_cic, semi), create full journey
	if strings.Contains(flowType, "auto") || strings.Contains(flowType, "semi") {
		// Add initial common steps from source
		initialSteps := []string{
			"otp", "app_form.basic_info", "appraising.quick_approval",
			"app_form.personal_info", "ekyc.selfie.active", "appraising.second_approval",
			"ekyc.id_card", "ekyc.confirm", "appraising.third_approval", "appraising.fourth_approval",
		}

		for _, stepName := range initialSteps {
			steps = append(steps, Step{
				ID:                       stepID,
				Name:                     stepName,
				MainUIVersion:            sourceConfig.UIVersion,
				SubUIVersion:             getSubUIVersionForStep(stepName, sourceConfig),
				SubUIVersionByConditions: getSubUIVersionConditions(stepName, sourceConfig),
			})
			stepID++
		}

		// Add automated flow specific steps
		automatedSteps := []string{
			"inform.success", "app_form.contact_info", "appraising.fifth_approval",
			"esign.intro", "esign.review", "esign.otp", "app_form.card_design",
			"app_form.personalize_reward", "ekyc.nfc_scan", "appraising.nfc_verify",
		}

		for _, stepName := range automatedSteps {
			subUIVersion := ""
			var subUIConditions []SubUIVersionByCondition

			// Add specific sub UI versions based on step and flow type
			switch stepName {
			case "inform.success":
				if strings.Contains(flowType, "semi") {
					subUIConditions = []SubUIVersionByCondition{
						{
							Condition:    "communication_call=success, lead_source=organic",
							SubUIVersion: "v1.1-semi",
						},
					}
				} else {
					subUIConditions = []SubUIVersionByCondition{
						{
							Condition:    "communication_call=success, lead_source=organic",
							SubUIVersion: "v1.1-auto",
						},
					}
				}
			case "app_form.contact_info", "appraising.fifth_approval", "esign.intro":
				subUIVersion = "v1.0-c1"
			case "esign.review":
				if strings.Contains(flowType, "semi") {
					subUIVersion = "v1.0-semi-nfc"
				} else {
					subUIVersion = "v1.0-auto-nfc"
				}
			}

			steps = append(steps, Step{
				ID:                       stepID,
				Name:                     stepName,
				MainUIVersion:            targetConfig.UIVersion,
				SubUIVersion:             subUIVersion,
				SubUIVersionByConditions: subUIConditions,
			})
			stepID++
		}
		return steps
	}

	// For CIF flows, add CIF-specific steps
	if strings.Contains(flowType, "cif") || strings.Contains(flowType, "diff") {
		// Add initial steps if needed, then CIF steps
		steps = append(steps, Step{
			ID:                       stepID,
			Name:                     "cif.confirm",
			MainUIVersion:            targetConfig.UIVersion,
			SubUIVersion:             "",
			SubUIVersionByConditions: []SubUIVersionByCondition{},
		})
		stepID++

		// Only add appraising.cif if not cif_no_branch
		if !strings.Contains(flowType, "no_branch") {
			steps = append(steps, Step{
				ID:                       stepID,
				Name:                     "appraising.cif",
				MainUIVersion:            targetConfig.UIVersion,
				SubUIVersion:             "",
				SubUIVersionByConditions: []SubUIVersionByCondition{},
			})
		}
		return steps
	}

	// Default: use target config's UI flow
	return GenerateStandardJourneySteps(targetConfig.UIFlow, targetConfig.UIVersion)
}

// Helper functions for step generation
func getSubUIVersionForStep(stepName string, config *LenderConfig) string {
	// Add logic to determine sub UI version based on step and config
	if stepName == "app_form.personal_info" {
		return "v1.0-c1"
	}
	return ""
}

func getSubUIVersionConditions(stepName string, config *LenderConfig) []SubUIVersionByCondition {
	// Add logic to determine sub UI version conditions
	return []SubUIVersionByCondition{}
}

// GenerateJourneyTemplate creates a complete journey template for a lender config
func GenerateJourneyTemplate(sourceConfigID int, relatedConfigs []RelatedConfigResult, folderPath string) (*JourneyTemplate, error) {
	// Read source config to get UI flow
	name, path := SearchLenderConfigID(sourceConfigID)
	if name == "" || path == "" {
		return nil, fmt.Errorf("cannot find lender config with ID: %d", sourceConfigID)
	}

	sourceConfig, err := ReadLenderConfig(path + "/" + name)
	if err != nil {
		return nil, fmt.Errorf("error reading source config: %w", err)
	}

	var relatedConfigIDs []int
	var journeys []Journey

	// Add self-loop journey (standard flow)
	standardSteps := GenerateStandardJourneySteps(sourceConfig.UIFlow, sourceConfig.UIVersion)
	standardJourney := GenerateJourneyFromTemplate(
		sourceConfigID,
		sourceConfigID,
		"normal",
		"",
		"Normal flow",
		standardSteps,
	)
	journeys = append(journeys, standardJourney)

	// Generate journeys for related configs
	for _, relatedConfig := range relatedConfigs {
		if relatedConfig.IsABTesting {
			continue // Skip A/B testing variants for journey generation
		}

		relatedConfigIDs = append(relatedConfigIDs, relatedConfig.ConfigID)

		// Read target config to get its UI flow
		targetName, targetPath := SearchLenderConfigID(relatedConfig.ConfigID)
		if targetName == "" || targetPath == "" {
			continue
		}

		targetConfig, err := ReadLenderConfig(targetPath + "/" + targetName)
		if err != nil {
			continue
		}

		// Generate journey based on flow type and match reason
		flowType := DetermineFlowType(sourceConfig, targetConfig, relatedConfig.MatchReason)
		condition := GenerateConditionFromMatchReason(relatedConfig.MatchReason)
		description := GenerateDescriptionFromFlowType(flowType, relatedConfig.Name)

		// Generate full journey steps combining source and target flows
		targetSteps := GenerateFullJourneySteps(sourceConfig, targetConfig, flowType)

		journey := GenerateJourneyFromTemplate(
			sourceConfigID,
			relatedConfig.ConfigID,
			flowType,
			condition,
			description,
			targetSteps,
		)

		journeys = append(journeys, journey)
	}

	template := &JourneyTemplate{
		SearchValue:      int64(sourceConfigID),
		SearchType:       "lender_config_id",
		RelatedConfigIDs: relatedConfigIDs,
		Journeys:         journeys,
	}

	return template, nil
}

// DetermineFlowType determines the flow type based on source and target configs
func DetermineFlowType(sourceConfig, targetConfig *LenderConfig, matchReason string) string {
	sourceFlowType := GetFlowTypeFromTags(sourceConfig.Tags)
	targetFlowType := GetFlowTypeFromTags(targetConfig.Tags)

	// If same flow type, it's a normal flow
	if sourceFlowType == targetFlowType {
		return "normal"
	}

	// Generate flow type based on transition
	return fmt.Sprintf("%s_to_%s", sourceFlowType, targetFlowType)
}

// GenerateConditionFromMatchReason creates a condition string based on match reason
func GenerateConditionFromMatchReason(matchReason string) string {
	// Parse match reason to generate appropriate conditions
	if strings.Contains(matchReason, "different flow_type") {
		return "flow_routing_condition == true"
	}
	if strings.Contains(matchReason, "same product_code") {
		return "product_eligibility == true"
	}
	if strings.Contains(matchReason, "same lead_source") {
		return "lead_source_match == true"
	}
	if strings.Contains(matchReason, "shared telco_code") {
		return "telco_compatibility == true"
	}

	return "routing_condition == true"
}

// GenerateDescriptionFromFlowType creates a human-readable description
func GenerateDescriptionFromFlowType(flowType, configName string) string {
	switch {
	case strings.Contains(flowType, "rejection"):
		return "Rejection flow"
	case strings.Contains(flowType, "auto"):
		return "Automated flow"
	case strings.Contains(flowType, "semi"):
		return "Semi-automated flow"
	case strings.Contains(flowType, "manual"):
		return "Manual review flow"
	case strings.Contains(flowType, "cif"):
		return "CIF verification flow"
	case strings.Contains(flowType, "diff"):
		return "Different information flow"
	case flowType == "normal":
		return "Normal flow"
	default:
		return fmt.Sprintf("Flow to %s", configName)
	}
}

// WriteJourneyTemplateToJSON exports journey template to JSON file
func WriteJourneyTemplateToJSON(template *JourneyTemplate, filename string) error {
	if err := CheckFile(filename); err != nil {
		return fmt.Errorf("failed to prepare file path: %w", err)
	}

	jsonData, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal journey template to JSON: %w", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file %s: %w", filename, err)
	}

	fmt.Printf("Journey template written to %s\n", filename)
	return nil
}

// ExportABTestingAnalysis exports A/B testing analysis to JSON file
func ExportABTestingAnalysis(lenderConfigID int, leadSource string, abGroups []ABTestingGroup, folderPath string) error {
	// Get detailed results for normal configs
	detailedResults := SearchRelatedConfigDetailed(lenderConfigID, leadSource, folderPath)

	// Separate normal results from A/B testing variants
	var normalResults []RelatedConfigResult
	for _, result := range detailedResults {
		if !result.IsABTesting {
			normalResults = append(normalResults, result)
		}
	}

	// Create analysis result
	analysisResult := ABTestingAnalysisResult{
		SearchID:        lenderConfigID,
		SearchType:      "ab_testing_analysis",
		ABTestingGroups: abGroups,
		NormalResults:   normalResults,
		TotalResults:    len(abGroups) + len(normalResults),
	}

	// Export to JSON
	filename := filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("ab_testing_analysis_%d_%s.json", lenderConfigID, leadSource))
	if err := CheckFile(filename); err != nil {
		return fmt.Errorf("failed to prepare file path: %w", err)
	}

	jsonData, err := json.MarshalIndent(analysisResult, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal A/B testing analysis to JSON: %w", err)
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("failed to write JSON file %s: %w", filename, err)
	}

	fmt.Printf("A/B testing analysis written to %s\n", filename)

	// Generate PlantUML diagram if there are A/B testing groups
	if len(abGroups) > 0 {
		pumlFilename := filepath.Join(GetConfigPumlDir(lenderConfigID), fmt.Sprintf("ab_testing_groups_%d_%s.puml", lenderConfigID, leadSource))
		err = GenerateABTestingDiagram(abGroups, pumlFilename)
		if err != nil {
			fmt.Printf("Warning: Failed to generate A/B testing PlantUML diagram: %v\n", err)
		} else {
			fmt.Printf("A/B testing PlantUML diagram written to %s\n", pumlFilename)

			// Export to PNG
			pngFilename := filepath.Join(GetConfigImagesDir(lenderConfigID), fmt.Sprintf("ab_testing_groups_%d_%s.png", lenderConfigID, leadSource))
			err = ExportPlantUMLToPNGCustomPath(pumlFilename, pngFilename)
			if err != nil {
				fmt.Printf("Warning: Failed to export A/B testing PNG (Java/PlantUML may not be available): %v\n", err)
			} else {
				fmt.Printf("A/B testing PNG diagram exported to %s\n", pngFilename)
			}
		}
	}

	return nil
}

// GenerateJourneyAnalysis performs complete journey analysis for a lender config
func GenerateJourneyAnalysis(lenderConfigID int, leadSource string, folderPath string) error {
	fmt.Printf("=== Generating Journey Analysis for Config %d ===\n", lenderConfigID)

	// Get related configs
	relatedConfigs := SearchRelatedConfigDetailed(lenderConfigID, leadSource, folderPath)
	fmt.Printf("Found %d related configs\n", len(relatedConfigs))

	// Generate journey template
	template, err := GenerateJourneyTemplate(lenderConfigID, relatedConfigs, folderPath)
	if err != nil {
		return fmt.Errorf("failed to generate journey template: %w", err)
	}

	fmt.Printf("Generated %d journeys:\n", len(template.Journeys))
	for i, journey := range template.Journeys {
		fmt.Printf("  %d. %s: %s -> %s (%d steps)\n",
			i+1, journey.ID, journey.FlowType, journey.Description, len(journey.Steps))
	}

	// Export to JSON
	filename := filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("journey_analysis_%d_%s.json", lenderConfigID, leadSource))
	err = WriteJourneyTemplateToJSON(template, filename)
	if err != nil {
		return fmt.Errorf("failed to write journey template: %w", err)
	}

	fmt.Printf("=== Journey Analysis Complete ===\n")
	return nil
}

// GenerateJourneyFlowDiagram creates a PlantUML diagram for journey flows
func GenerateJourneyFlowDiagram(template *JourneyTemplate, filename string) error {
	var puml strings.Builder

	puml.WriteString("@startuml\n")

	// Add Materia theme
	puml.WriteString("!$THEME = \"materia\"\n\n")
	puml.WriteString("!if %not(%variable_exists(\"$BGCOLOR\"))\n")
	puml.WriteString("!$BGCOLOR = \"$WHITE\"\n")
	puml.WriteString("!endif\n\n")
	puml.WriteString("skinparam backgroundColor $BGCOLOR\n")
	puml.WriteString("skinparam useBetaStyle false\n\n")

	// Define colors
	puml.WriteString("!$PRIMARY = \"#2196F3\"\n")
	puml.WriteString("!$SUCCESS = \"#4CAF50\"\n")
	puml.WriteString("!$WARNING = \"#ff9800\"\n")
	puml.WriteString("!$DANGER = \"#e51c23\"\n")
	puml.WriteString("!$INFO = \"#9C27B0\"\n")
	puml.WriteString("!$WHITE = \"#FFF\"\n")
	puml.WriteString("!$DARK = \"#222\"\n\n")

	// Apply styling
	puml.WriteString("skinparam rectangle {\n")
	puml.WriteString("  BackgroundColor $PRIMARY\n")
	puml.WriteString("  BorderColor $PRIMARY\n")
	puml.WriteString("  FontColor $WHITE\n")
	puml.WriteString("  BorderThickness 2\n")
	puml.WriteString("}\n\n")

	puml.WriteString("skinparam arrow {\n")
	puml.WriteString("  Color $DARK\n")
	puml.WriteString("  FontColor $DARK\n")
	puml.WriteString("  Thickness 2\n")
	puml.WriteString("}\n\n")

	puml.WriteString(fmt.Sprintf("title Journey Flow Analysis - Config %d\n\n", template.SearchValue))

	// Define the source config
	puml.WriteString(fmt.Sprintf("rectangle \"Config %d\\n(Source)\" as config_%d $PRIMARY\n",
		template.SearchValue, template.SearchValue))

	// Define target configs
	configMap := make(map[int]bool)
	for _, journey := range template.Journeys {
		if journey.ToLenderConfigID != int(template.SearchValue) && !configMap[journey.ToLenderConfigID] {
			configMap[journey.ToLenderConfigID] = true

			// Determine color based on flow type using theme colors
			color := "$SUCCESS"
			if strings.Contains(journey.FlowType, "rejection") {
				color = "$DANGER"
			} else if strings.Contains(journey.FlowType, "auto") {
				color = "$WARNING"
			} else if strings.Contains(journey.FlowType, "semi") {
				color = "$INFO"
			} else if strings.Contains(journey.FlowType, "cif") {
				color = "$PRIMARY"
			}

			puml.WriteString(fmt.Sprintf("rectangle \"Config %d\\n%s\" as config_%d %s\n",
				journey.ToLenderConfigID, journey.Description, journey.ToLenderConfigID, color))
		}
	}

	puml.WriteString("\n")

	// Add journey connections
	for _, journey := range template.Journeys {
		if journey.FromLenderConfigID != journey.ToLenderConfigID {
			// Connection to other configs only (skip self-loops for cleaner diagram)
			label := journey.FlowType
			if journey.Condition != "" {
				// Escape condition text for PlantUML
				label = journey.FlowType // Simplified label
			}

			puml.WriteString(fmt.Sprintf("config_%d --> config_%d : %s\n",
				journey.FromLenderConfigID, journey.ToLenderConfigID, label))
		}
	}

	// Add legend
	puml.WriteString("\nlegend right\n")
	puml.WriteString("  |Color|Flow Type|\n")
	puml.WriteString("  |<#lightblue>|Source Config|\n")
	puml.WriteString("  |<#lightgreen>|Normal Flow|\n")
	puml.WriteString("  |<#lightyellow>|Automated Flow|\n")
	puml.WriteString("  |<#lightpink>|Semi-Automated Flow|\n")
	puml.WriteString("  |<#lightcyan>|CIF Verification|\n")
	puml.WriteString("  |<#lightcoral>|Rejection Flow|\n")
	puml.WriteString("endlegend\n")

	puml.WriteString("\n@enduml\n")

	// Write to file
	if err := CheckFile(filename); err != nil {
		return fmt.Errorf("failed to prepare file path: %w", err)
	}

	err := os.WriteFile(filename, []byte(puml.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PlantUML file %s: %w", filename, err)
	}

	fmt.Printf("Journey flow PlantUML diagram written to %s\n", filename)
	return nil
}

// GenerateCompleteJourneyAnalysis performs complete journey analysis with visualization
func GenerateCompleteJourneyAnalysis(lenderConfigID int, leadSource string, folderPath string) error {
	fmt.Printf("=== Generating Complete Journey Analysis for Config %d ===\n", lenderConfigID)

	// Get related configs
	relatedConfigs := SearchRelatedConfigDetailed(lenderConfigID, leadSource, folderPath)
	fmt.Printf("Found %d related configs\n", len(relatedConfigs))

	// Generate journey template
	template, err := GenerateJourneyTemplate(lenderConfigID, relatedConfigs, folderPath)
	if err != nil {
		return fmt.Errorf("failed to generate journey template: %w", err)
	}

	fmt.Printf("Generated %d journeys\n", len(template.Journeys))

	// Export JSON
	jsonFilename := filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("journey_analysis_%d_%s.json", lenderConfigID, leadSource))
	err = WriteJourneyTemplateToJSON(template, jsonFilename)
	if err != nil {
		return fmt.Errorf("failed to write journey template: %w", err)
	}

	// Generate PlantUML diagram
	pumlFilename := filepath.Join(GetConfigPumlDir(lenderConfigID), fmt.Sprintf("journey_flow_%d_%s.puml", lenderConfigID, leadSource))
	err = GenerateJourneyFlowDiagram(template, pumlFilename)
	if err != nil {
		return fmt.Errorf("failed to generate journey flow diagram: %w", err)
	}

	// Export to PNG in images directory
	pngFilename := filepath.Join(GetConfigImagesDir(lenderConfigID), fmt.Sprintf("journey_flow_%d_%s.png", lenderConfigID, leadSource))
	err = ExportPlantUMLToPNGCustomPath(pumlFilename, pngFilename)
	if err != nil {
		fmt.Printf("Warning: Failed to export PNG (Java/PlantUML may not be available): %v\n", err)
	}

	// Export individual journey step diagrams
	err = ExportAllJourneysPlantUML(template, lenderConfigID, leadSource)
	if err != nil {
		fmt.Printf("Warning: Failed to export individual journey diagrams: %v\n", err)
	}

	fmt.Printf("=== Complete Journey Analysis Finished ===\n")
	return nil
}

// ExportPlantUMLToPNGCustomPath exports PlantUML file to PNG with custom output path
func ExportPlantUMLToPNGCustomPath(pumlFilename, pngFilename string) error {
	// Check if Java is available
	if _, err := exec.LookPath("java"); err != nil {
		return fmt.Errorf("java not found in PATH, please install Java to export PNG diagrams")
	}

	// Ensure output directory exists
	if err := CheckFile(pngFilename); err != nil {
		return fmt.Errorf("failed to prepare PNG output path: %w", err)
	}

	// Create a temporary directory for PlantUML output
	tempDir := filepath.Dir(pumlFilename)

	// Run PlantUML to convert to PNG (output to same directory as PUML file)
	cmd := exec.Command("java", "-jar", "../plantuml.jar", "-tpng", pumlFilename)
	fmt.Printf("Converting PlantUML to PNG: %s\n", strings.Join(cmd.Args, " "))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to convert PlantUML to PNG: %w\nOutput: %s", err, string(output))
	}

	// PlantUML creates PNG with same base name as PUML in same directory
	pumlBasename := filepath.Base(pumlFilename)
	pumlBasenameNoExt := strings.TrimSuffix(pumlBasename, filepath.Ext(pumlBasename))
	generatedPNG := filepath.Join(tempDir, pumlBasenameNoExt+".png")

	if _, err := os.Stat(generatedPNG); err == nil {
		if generatedPNG != pngFilename {
			err = os.Rename(generatedPNG, pngFilename)
			if err != nil {
				return fmt.Errorf("failed to move PNG file from %s to %s: %w", generatedPNG, pngFilename, err)
			}
		}
		fmt.Printf("PNG diagram exported to %s\n", pngFilename)
	} else {
		return fmt.Errorf("PNG file was not generated at expected location: %s", generatedPNG)
	}

	return nil
}

// GenerateCompleteAnalysis performs all analyses and writes all results for a lender config
func GenerateCompleteAnalysis(lenderConfigID int, leadSource string, folderPath string) error {
	fmt.Printf("=== Starting Complete Analysis for Lender Config %d ===\n", lenderConfigID)

	// 1. A/B Testing Analysis
	fmt.Printf("\n--- Step 1: A/B Testing Analysis ---\n")
	abGroups := FindAllABTestingGroups(folderPath)
	err := ExportABTestingAnalysis(lenderConfigID, leadSource, abGroups, folderPath)
	if err != nil {
		fmt.Printf("Warning: A/B Testing Analysis failed: %v\n", err)
	} else {
		fmt.Printf("✅ A/B Testing Analysis completed\n")
	}

	// 2. Journey Analysis
	fmt.Printf("\n--- Step 2: Journey Analysis ---\n")
	err = GenerateJourneyAnalysis(lenderConfigID, leadSource, folderPath)
	if err != nil {
		fmt.Printf("Warning: Journey Analysis failed: %v\n", err)
	} else {
		fmt.Printf("✅ Journey Analysis completed\n")
	}

	// 3. Complete Journey Analysis with Visualization
	fmt.Printf("\n--- Step 3: Journey Visualization ---\n")
	err = GenerateCompleteJourneyAnalysis(lenderConfigID, leadSource, folderPath)
	if err != nil {
		fmt.Printf("Warning: Journey Visualization failed: %v\n", err)
	} else {
		fmt.Printf("✅ Journey Visualization completed\n")
	}

	// 4. Generate Summary Report
	fmt.Printf("\n--- Step 4: Summary Report ---\n")
	err = GenerateSummaryReport(lenderConfigID, leadSource)
	if err != nil {
		fmt.Printf("Warning: Summary Report failed: %v\n", err)
	} else {
		fmt.Printf("✅ Summary Report completed\n")
	}

	fmt.Printf("\n=== Complete Analysis Finished for Config %d ===\n", lenderConfigID)
	return nil
}

// GenerateSummaryReport creates a comprehensive summary of all analyses
func GenerateSummaryReport(lenderConfigID int, leadSource string) error {
	var report strings.Builder

	report.WriteString(fmt.Sprintf("# Complete Analysis Report - Config %d\n\n", lenderConfigID))
	report.WriteString(fmt.Sprintf("**Lead Source:** %s\n", leadSource))
	report.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))

	// Read A/B Testing Analysis
	abFilename := filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("ab_testing_analysis_%d_%s.json", lenderConfigID, leadSource))
	if abData, err := os.ReadFile(abFilename); err == nil {
		var abAnalysis ABTestingAnalysisResult
		if json.Unmarshal(abData, &abAnalysis) == nil {
			report.WriteString("## A/B Testing Analysis\n\n")
			report.WriteString(fmt.Sprintf("- **Total A/B Testing Groups:** %d\n", len(abAnalysis.ABTestingGroups)))

			for i, group := range abAnalysis.ABTestingGroups {
				report.WriteString(fmt.Sprintf("- **Group %d:** %s (%d variants, total weight: %d)\n",
					i+1, group.GroupName, len(group.Variants), group.TotalWeight))

				for j, variant := range group.Variants {
					report.WriteString(fmt.Sprintf("  - Variant %d: Config %d (weight: %d, %d steps)\n",
						j+1, variant.ConfigID, variant.Weight, len(variant.UIFlow)))
				}
			}

			report.WriteString(fmt.Sprintf("- **Normal Results:** %d configs\n\n", len(abAnalysis.NormalResults)))
		}
	}

	// Read Journey Analysis
	journeyFilename := filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("journey_analysis_%d_%s.json", lenderConfigID, leadSource))
	if journeyData, err := os.ReadFile(journeyFilename); err == nil {
		var journeyTemplate JourneyTemplate
		if json.Unmarshal(journeyData, &journeyTemplate) == nil {
			report.WriteString("## Journey Analysis\n\n")
			report.WriteString(fmt.Sprintf("- **Total Journeys:** %d\n", len(journeyTemplate.Journeys)))
			report.WriteString(fmt.Sprintf("- **Related Config IDs:** %v\n\n", journeyTemplate.RelatedConfigIDs))

			// Group journeys by flow type
			flowTypes := make(map[string]int)
			for _, journey := range journeyTemplate.Journeys {
				flowTypes[journey.FlowType]++
			}

			report.WriteString("### Journey Flow Types:\n")
			for flowType, count := range flowTypes {
				report.WriteString(fmt.Sprintf("- **%s:** %d journeys\n", flowType, count))
			}
			report.WriteString("\n")
		}
	}

	// Generated Files Section
	report.WriteString("## Generated Files\n\n")

	files := []struct {
		name        string
		description string
	}{
		{filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("ab_testing_analysis_%d_%s.json", lenderConfigID, leadSource)), "A/B Testing Analysis (JSON)"},
		{filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("journey_analysis_%d_%s.json", lenderConfigID, leadSource)), "Journey Analysis (JSON)"},
		{filepath.Join(GetConfigPumlDir(lenderConfigID), fmt.Sprintf("journey_flow_%d_%s.puml", lenderConfigID, leadSource)), "Journey Flow Diagram (PlantUML)"},
		{filepath.Join(GetConfigImagesDir(lenderConfigID), fmt.Sprintf("journey_flow_%d_%s.png", lenderConfigID, leadSource)), "Journey Flow Diagram (PNG)"},
		{filepath.Join(GetConfigImagesDir(lenderConfigID), fmt.Sprintf("ab_testing_groups_%d_%s.png", lenderConfigID, leadSource)), "A/B Testing Groups Diagram (PNG)"},
	}

	for _, file := range files {
		if _, err := os.Stat(file.name); err == nil {
			report.WriteString(fmt.Sprintf("- ✅ **%s:** `%s`\n", file.description, file.name))
		} else {
			report.WriteString(fmt.Sprintf("- ❌ **%s:** `%s` (not generated)\n", file.description, file.name))
		}
	}

	// Write summary report
	summaryFilename := filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("summary_report_%d_%s.md", lenderConfigID, leadSource))
	if err := CheckFile(summaryFilename); err != nil {
		return fmt.Errorf("failed to prepare summary file path: %w", err)
	}

	err := os.WriteFile(summaryFilename, []byte(report.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write summary report %s: %w", summaryFilename, err)
	}

	fmt.Printf("Summary report written to %s\n", summaryFilename)
	return nil
}

// SearchLenderConfigComplete performs complete search and analysis for a lender config
func SearchLenderConfigComplete(lenderConfigID int, leadSource string, folderPath string) error {
	fmt.Printf("🔍 Starting Complete Lender Config Search for ID: %d\n", lenderConfigID)
	fmt.Printf("📊 Lead Source: %s\n", leadSource)
	fmt.Printf("📁 Search Path: %s\n\n", folderPath)

	// Validate that the config exists
	name, path := SearchLenderConfigID(lenderConfigID)
	if name == "" || path == "" {
		return fmt.Errorf("lender config with ID %d not found", lenderConfigID)
	}

	fmt.Printf("✅ Found lender config: %s at %s\n\n", name, path)

	// Perform complete analysis
	err := GenerateCompleteAnalysis(lenderConfigID, leadSource, folderPath)
	if err != nil {
		return fmt.Errorf("complete analysis failed: %w", err)
	}

	fmt.Printf("\n🎉 Complete search and analysis finished successfully!\n")
	fmt.Printf("📋 Check the summary report: %s/summary_report_%d_%s.md\n", GetConfigResultsDir(lenderConfigID), lenderConfigID, leadSource)

	return nil
}

// ExportJourneyStepsPlantUML exports PlantUML diagram for individual journey showing UI versions with branching
func ExportJourneyStepsPlantUML(journey Journey, filename string) error {
	var puml strings.Builder

	puml.WriteString("@startuml\n")

	// Add Materia theme
	puml.WriteString("!$THEME = \"materia\"\n\n")
	puml.WriteString("!if %not(%variable_exists(\"$BGCOLOR\"))\n")
	puml.WriteString("!$BGCOLOR = \"transparent\"\n")
	puml.WriteString("!endif\n\n")
	puml.WriteString("skinparam backgroundColor $BGCOLOR\n")
	puml.WriteString("skinparam useBetaStyle false\n\n")

	// Define colors
	puml.WriteString("!$BLUE = \"#2196F3\"\n")
	puml.WriteString("!$GREEN = \"#4CAF50\"\n")
	puml.WriteString("!$ORANGE = \"#fd7e14\"\n")
	puml.WriteString("!$RED = \"#e51c23\"\n")
	puml.WriteString("!$PRIMARY = \"#2196F3\"\n")
	puml.WriteString("!$SUCCESS = \"#4CAF50\"\n")
	puml.WriteString("!$WARNING = \"#ff9800\"\n")
	puml.WriteString("!$DANGER = \"#e51c23\"\n")
	puml.WriteString("!$WHITE = \"#FFF\"\n")
	puml.WriteString("!$DARK = \"#222\"\n\n")

	// Apply activity styling
	puml.WriteString("skinparam activity {\n")
	puml.WriteString("  BackgroundColor $PRIMARY\n")
	puml.WriteString("  BorderColor $BLUE\n")
	puml.WriteString("  FontColor $WHITE\n")
	puml.WriteString("  StartColor $SUCCESS\n")
	puml.WriteString("  EndColor $DANGER\n")
	puml.WriteString("  DiamondBackgroundColor $WARNING\n")
	puml.WriteString("  DiamondBorderColor $ORANGE\n")
	puml.WriteString("  DiamondFontColor $DARK\n")
	puml.WriteString("}\n\n")

	puml.WriteString("skinparam arrow {\n")
	puml.WriteString("  Color $PRIMARY\n")
	puml.WriteString("  FontColor $DARK\n")
	puml.WriteString("  Thickness 2\n")
	puml.WriteString("}\n\n")

	puml.WriteString(fmt.Sprintf("title Journey Steps - %s - %s\n\n", journey.ID, journey.Description))

	puml.WriteString("start\n")

	for i, step := range journey.Steps {
		stepLabel := fmt.Sprintf("Step %d: %s", step.ID, step.Name)

		// Check if this step has conditional UI versions
		if len(step.SubUIVersionByConditions) > 0 {
			// Create branching logic for conditional UI
			puml.WriteString(fmt.Sprintf(":%s;\n", stepLabel))

			// Create decision point
			for j, condition := range step.SubUIVersionByConditions {
				// Escape special characters for PlantUML
				conditionText := strings.ReplaceAll(condition.Condition, "==", "equals")
				conditionText = strings.ReplaceAll(conditionText, "&&", "and")
				conditionText = strings.ReplaceAll(conditionText, "||", "or")
				conditionText = strings.ReplaceAll(conditionText, ",", " and")

				if j == 0 {
					puml.WriteString(fmt.Sprintf("if (%s?) then (yes)\n", conditionText))
				}

				// Branch for conditional UI version
				puml.WriteString(fmt.Sprintf("  :Use UI Version\\n%s;\n", condition.SubUIVersion))

				puml.WriteString("else (no)\n")
				// Branch for main UI version
				var fallbackUIText string
				if step.SubUIVersion != "" {
					fallbackUIText = fmt.Sprintf("%s\\n(Main: %s)", step.SubUIVersion, step.MainUIVersion)
				} else {
					fallbackUIText = step.MainUIVersion
				}
				puml.WriteString(fmt.Sprintf("  :Use UI Version\\n%s;\n", fallbackUIText))
				puml.WriteString("endif\n")
			}
		} else {
			// Regular step without conditional UI
			var stepText string

			// If sub UI version exists, prioritize it as the main display
			if step.SubUIVersion != "" {
				stepText = fmt.Sprintf("%s\\nUI Version: %s\\n(Main: %s)", stepLabel, step.SubUIVersion, step.MainUIVersion)
			} else {
				stepText = fmt.Sprintf("%s\\nUI Version: %s", stepLabel, step.MainUIVersion)
			}

			puml.WriteString(fmt.Sprintf(":%s;\n", stepText))
		}

		// Add separator between steps (except for last step)
		if i < len(journey.Steps)-1 {
			puml.WriteString("\n")
		}
	}

	puml.WriteString("\nstop\n")

	// Add note with journey info
	puml.WriteString("\nnote right\n")
	puml.WriteString("Journey Information:\n")
	puml.WriteString(fmt.Sprintf("Flow Type: %s\n", journey.FlowType))
	puml.WriteString(fmt.Sprintf("From Config: %d\n", journey.FromLenderConfigID))
	puml.WriteString(fmt.Sprintf("To Config: %d\n", journey.ToLenderConfigID))
	if journey.Condition != "" {
		puml.WriteString(fmt.Sprintf("Condition: %s\n", journey.Condition))
	}
	puml.WriteString("\\nUI Version Legend:\\n")
	puml.WriteString("- Main UI: Primary version\n")
	puml.WriteString("- Sub UI: Secondary version\n")
	puml.WriteString("- Conditional: Dynamic based on conditions\n")
	puml.WriteString("end note\n")

	puml.WriteString("\n@enduml\n")

	// Write to file
	if err := CheckFile(filename); err != nil {
		return fmt.Errorf("failed to prepare file path: %w", err)
	}

	err := os.WriteFile(filename, []byte(puml.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PlantUML file %s: %w", filename, err)
	}

	fmt.Printf("Journey steps PlantUML diagram written to %s\n", filename)
	return nil
}

// ExportAllJourneysPlantUML exports individual PlantUML files for all journeys
func ExportAllJourneysPlantUML(template *JourneyTemplate, lenderConfigID int, leadSource string) error {
	fmt.Printf("=== Exporting Individual Journey PlantUML Files ===\n")

	for i, journey := range template.Journeys {
		// Create filename for each journey
		filename := filepath.Join(GetConfigPumlDir(lenderConfigID), fmt.Sprintf("journey_steps_%d_%s_%s.puml",
			lenderConfigID, leadSource, sanitizeFilename(journey.ID)))

		err := ExportJourneyStepsPlantUML(journey, filename)
		if err != nil {
			fmt.Printf("Warning: Failed to export journey %s: %v\n", journey.ID, err)
			continue
		}

		// Export to PNG
		pngFilename := filepath.Join(GetConfigImagesDir(lenderConfigID), fmt.Sprintf("journey_steps_%d_%s_%s.png",
			lenderConfigID, leadSource, sanitizeFilename(journey.ID)))
		err = ExportPlantUMLToPNGCustomPath(filename, pngFilename)
		if err != nil {
			fmt.Printf("Warning: Failed to export PNG for journey %s: %v\n", journey.ID, err)
		}

		fmt.Printf("  %d. %s (%d steps)\n", i+1, journey.ID, len(journey.Steps))
	}

	fmt.Printf("=== Individual Journey Export Complete ===\n")
	return nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(filename string) string {
	// Replace invalid characters with underscores
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")
	return filename
}

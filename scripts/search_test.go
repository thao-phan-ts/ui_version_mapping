package ui_version_check

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestABTestingAnalysis tests A/B testing detection and analysis functionality
func TestABTestingAnalysis(t *testing.T) {
	// Test configuration
	lenderConfigID := 9054
	leadSource := "organic"
	evoPath := "submodules/digital_journey/migration/sync/vietnam/tpbank/lender_configs/evo"

	// Step 1: Find all A/B testing groups in the evo folder
	fmt.Printf("=== STEP 1: Finding A/B Testing Groups ===\n")
	abGroups := FindAllABTestingGroups(evoPath)

	if len(abGroups) == 0 {
		fmt.Printf("No A/B testing groups found in %s\n", evoPath)
	} else {
		fmt.Printf("Found %d A/B testing groups:\n", len(abGroups))
		for i, group := range abGroups {
			fmt.Printf("  Group %d: %s (Total Weight: %d)\n", i+1, group.GroupName, group.TotalWeight)
			for j, variant := range group.Variants {
				percentage := float64(variant.Weight) / float64(group.TotalWeight) * 100
				fmt.Printf("    Variant %d: Config %d (Weight: %d = %.1f%%)\n",
					j+1, variant.ConfigID, variant.Weight, percentage)
				if len(variant.Differences) > 0 {
					fmt.Printf("      Differences: %v\n", variant.Differences)
				}
			}
		}
	}

	// Step 2: Search for related configs with detailed analysis
	fmt.Printf("\n=== STEP 2: Detailed Related Config Analysis ===\n")
	detailedResults := SearchRelatedConfigDetailed(lenderConfigID, leadSource, evoPath)

	// Separate A/B testing variants from normal results
	var normalResults []RelatedConfigResult
	var abTestingResults []RelatedConfigResult

	for _, result := range detailedResults {
		if result.IsABTesting {
			abTestingResults = append(abTestingResults, result)
		} else {
			normalResults = append(normalResults, result)
		}
	}

	fmt.Printf("Found %d related configs (%d A/B variants, %d normal matches)\n",
		len(detailedResults), len(abTestingResults), len(normalResults))

	// Step 3: Export results to JSON
	fmt.Printf("\n=== STEP 3: Exporting Analysis Results ===\n")
	analysisResult := struct {
		SearchConfigID  int                   `json:"search_config_id"`
		LeadSource      string                `json:"lead_source"`
		SearchPath      string                `json:"search_path"`
		ABTestingGroups []ABTestingGroup      `json:"ab_testing_groups"`
		NormalResults   []RelatedConfigResult `json:"normal_results"`
		SearchType      string                `json:"search_type"`
	}{
		SearchConfigID:  lenderConfigID,
		LeadSource:      leadSource,
		SearchPath:      evoPath,
		SearchType:      SearchTypeABTestingAnalysis,
		ABTestingGroups: abGroups,
		NormalResults:   normalResults,
	}

	filename := filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("ab_testing_analysis_%d_%s.json", lenderConfigID, leadSource))
	if err := CheckFile(filename); err != nil {
		t.Errorf("Failed to prepare file path: %v", err)
		return
	}

	jsonData, err := json.MarshalIndent(analysisResult, "", "  ")
	if err != nil {
		t.Errorf("Failed to marshal analysis result: %v", err)
		return
	}

	err = os.WriteFile(filename, jsonData, 0644)
	if err != nil {
		t.Errorf("Failed to write JSON file: %v", err)
		return
	}

	fmt.Printf("Analysis results exported to %s\n", filename)

	// Step 4: Generate PlantUML diagram
	fmt.Printf("\n=== STEP 4: Generating PlantUML Diagram ===\n")
	pumlFilename := filepath.Join(GetConfigPumlDir(lenderConfigID), fmt.Sprintf("ab_testing_groups_%d_%s.puml", lenderConfigID, leadSource))
	err = GenerateABTestingDiagram(abGroups, pumlFilename)
	if err != nil {
		t.Errorf("Failed to generate PlantUML diagram: %v", err)
		return
	}

	// Step 5: Export to PNG
	fmt.Printf("\n=== STEP 5: Converting to PNG ===\n")
	err = ExportPlantUMLToPNG(pumlFilename)
	if err != nil {
		fmt.Printf("Warning: Failed to export PNG (Java/PlantUML may not be available): %v\n", err)
	}

	fmt.Printf("\n=== A/B Testing Analysis Complete ===\n")
}

// TestCompleteSearch tests the complete lender config search that writes all results
func TestCompleteSearch(t *testing.T) {
	fmt.Printf("=== Testing Complete Lender Config Search ===\n")

	// Test configuration
	lenderConfigID := 9054
	leadSource := "organic"
	evoPath := "submodules/digital_journey/migration/sync/vietnam/tpbank/lender_configs/evo"

	// Perform complete search and analysis
	err := SearchLenderConfigComplete(lenderConfigID, leadSource, evoPath)
	if err != nil {
		t.Errorf("Failed to perform complete search: %v", err)
		return
	}

	// Verify all expected files were created
	expectedFiles := []string{
		filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("ab_testing_analysis_%d_%s.json", lenderConfigID, leadSource)),
		filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("journey_analysis_%d_%s.json", lenderConfigID, leadSource)),
		filepath.Join(GetConfigPumlDir(lenderConfigID), fmt.Sprintf("journey_flow_%d_%s.puml", lenderConfigID, leadSource)),
		filepath.Join(GetConfigImagesDir(lenderConfigID), fmt.Sprintf("journey_flow_%d_%s.png", lenderConfigID, leadSource)),
		filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("summary_report_%d_%s.md", lenderConfigID, leadSource)),
	}

	fmt.Printf("\n=== Verifying Generated Files ===\n")
	for _, filename := range expectedFiles {
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			t.Errorf("Expected file was not created: %s", filename)
		} else {
			fmt.Printf("✅ %s\n", filename)
		}
	}

	// Read and validate the summary report
	summaryFilename := filepath.Join(GetConfigResultsDir(lenderConfigID), fmt.Sprintf("summary_report_%d_%s.md", lenderConfigID, leadSource))
	summaryData, err := os.ReadFile(summaryFilename)
	if err != nil {
		t.Errorf("Failed to read summary report: %v", err)
		return
	}

	summaryContent := string(summaryData)

	// Validate summary report contains expected sections
	expectedSections := []string{
		"# Complete Analysis Report",
		"## A/B Testing Analysis",
		"## Journey Analysis",
		"## Generated Files",
	}

	for _, section := range expectedSections {
		if !strings.Contains(summaryContent, section) {
			t.Errorf("Summary report missing expected section: %s", section)
		}
	}

	fmt.Printf("\n=== Summary Report Preview ===\n")
	lines := strings.Split(summaryContent, "\n")
	for i, line := range lines {
		if i < 20 { // Show first 20 lines
			fmt.Printf("%s\n", line)
		}
	}
	if len(lines) > 20 {
		fmt.Printf("... (%d more lines)\n", len(lines)-20)
	}

	fmt.Printf("\n=== Complete Search Tests Finished ===\n")
}

// TestIndividualJourneyExport tests the individual journey PlantUML export functionality
func TestIndividualJourneyExport(t *testing.T) {
	fmt.Printf("=== Testing Individual Journey PlantUML Export ===\n")

	// Test configuration
	lenderConfigID := 9054
	leadSource := "organic"
	evoPath := "submodules/digital_journey/migration/sync/vietnam/tpbank/lender_configs/evo"

	// Generate journey template
	relatedConfigs := SearchRelatedConfigDetailed(lenderConfigID, leadSource, evoPath)
	template, err := GenerateJourneyTemplate(lenderConfigID, relatedConfigs, evoPath)
	if err != nil {
		t.Errorf("Failed to generate journey template: %v", err)
		return
	}

	fmt.Printf("Generated %d journeys for individual export\n", len(template.Journeys))

	// Export individual journey diagrams
	err = ExportAllJourneysPlantUML(template, lenderConfigID, leadSource)
	if err != nil {
		t.Errorf("Failed to export individual journey diagrams: %v", err)
		return
	}

	// Verify some key journey files were created
	keyJourneys := []string{
		"from_9054_to_9095", // auto_pcb - should have 20 steps with UI versions
		"from_9054_to_9097", // semi - should have conditional UI versions
		"from_9054_to_9048", // rejection - should have simple flow
	}

	fmt.Printf("\n=== Verifying Key Journey Files ===\n")
	for _, journeyID := range keyJourneys {
		// Check PlantUML file
		pumlFilename := filepath.Join(GetConfigPumlDir(lenderConfigID), fmt.Sprintf("journey_steps_%d_%s_%s.puml",
			lenderConfigID, leadSource, journeyID))
		if _, err := os.Stat(pumlFilename); os.IsNotExist(err) {
			t.Errorf("Expected PlantUML file was not created: %s", pumlFilename)
		} else {
			fmt.Printf("✅ %s\n", pumlFilename)

			// Read and show a preview of the PlantUML content
			content, err := os.ReadFile(pumlFilename)
			if err == nil {
				lines := strings.Split(string(content), "\n")
				fmt.Printf("   Preview (first 10 lines):\n")
				for i, line := range lines {
					if i < 10 {
						fmt.Printf("   %s\n", line)
					}
				}
				fmt.Printf("   ... (%d total lines)\n", len(lines))
			}
		}

		// Check PNG file
		pngFilename := filepath.Join(GetConfigImagesDir(lenderConfigID), fmt.Sprintf("journey_steps_%d_%s_%s.png",
			lenderConfigID, leadSource, journeyID))
		if _, err := os.Stat(pngFilename); os.IsNotExist(err) {
			fmt.Printf("⚠️  PNG not generated: %s (Java/PlantUML may not be available)\n", pngFilename)
		} else {
			fmt.Printf("✅ %s\n", pngFilename)
		}
		fmt.Printf("\n")
	}

	fmt.Printf("=== Individual Journey Export Tests Complete ===\n")
}

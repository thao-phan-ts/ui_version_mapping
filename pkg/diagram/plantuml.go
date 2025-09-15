package diagram

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tsocial/ui-version-mapping/pkg/analyzer"
	"github.com/tsocial/ui-version-mapping/pkg/journey"
)

// DiagramConfig represents configuration for PlantUML diagram generation
type DiagramConfig struct {
	Title      string
	UIFlow     []string
	ConfigID   int
	FlowType   string
	UIVersion  string
	OutputPath string
}

// ActivityDiagram represents a PlantUML Activity Diagram
type ActivityDiagram struct {
	Config   DiagramConfig
	Content  string
	FilePath string
}

// GenerateABTestingDiagram creates PlantUML diagram for A/B testing groups
func GenerateABTestingDiagram(groups []analyzer.ABTestingGroup, filename string) error {
	var puml strings.Builder

	puml.WriteString("@startuml\n")
	puml.WriteString("title A/B Testing Groups Analysis\n\n")

	for i, group := range groups {
		puml.WriteString(fmt.Sprintf("package \"Group %d: %s\" {\n", i+1, group.GroupName))

		for j, variant := range group.Variants {
			percentage := float64(variant.Weight) / float64(group.TotalWeight) * 100
			puml.WriteString(fmt.Sprintf("  rectangle \"Config %d\\nWeight: %d (%.1f%%)\" as config_%d_%d\n",
				variant.ConfigID, variant.Weight, percentage, i, j))
		}

		puml.WriteString("}\n\n")
	}

	puml.WriteString("@enduml\n")

	// Write to file
	if err := ensureDir(filename); err != nil {
		return fmt.Errorf("failed to prepare file path: %w", err)
	}

	err := os.WriteFile(filename, []byte(puml.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PlantUML file %s: %w", filename, err)
	}

	fmt.Printf("A/B Testing PlantUML diagram written to %s\n", filename)
	return nil
}

// GenerateJourneyFlowDiagram creates a PlantUML diagram for journey flows
func GenerateJourneyFlowDiagram(template *journey.JourneyTemplate, filename string) error {
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
	for _, j := range template.Journeys {
		if j.ToLenderConfigID != int(template.SearchValue) && !configMap[j.ToLenderConfigID] {
			configMap[j.ToLenderConfigID] = true

			// Determine color based on flow type using theme colors
			color := "$SUCCESS"
			if strings.Contains(j.FlowType, "rejection") {
				color = "$DANGER"
			} else if strings.Contains(j.FlowType, "auto") {
				color = "$WARNING"
			} else if strings.Contains(j.FlowType, "semi") {
				color = "$INFO"
			} else if strings.Contains(j.FlowType, "cif") {
				color = "$PRIMARY"
			}

			puml.WriteString(fmt.Sprintf("rectangle \"Config %d\\n%s\" as config_%d %s\n",
				j.ToLenderConfigID, j.Description, j.ToLenderConfigID, color))
		}
	}

	puml.WriteString("\n")

	// Add journey connections
	for _, j := range template.Journeys {
		if j.FromLenderConfigID != j.ToLenderConfigID {
			// Connection to other configs only (skip self-loops for cleaner diagram)
			label := j.FlowType
			if j.Condition != "" {
				// Escape condition text for PlantUML
				label = j.FlowType // Simplified label
			}

			puml.WriteString(fmt.Sprintf("config_%d --> config_%d : %s\n",
				j.FromLenderConfigID, j.ToLenderConfigID, label))
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
	if err := ensureDir(filename); err != nil {
		return fmt.Errorf("failed to prepare file path: %w", err)
	}

	err := os.WriteFile(filename, []byte(puml.String()), 0644)
	if err != nil {
		return fmt.Errorf("failed to write PlantUML file %s: %w", filename, err)
	}

	fmt.Printf("Journey flow PlantUML diagram written to %s\n", filename)
	return nil
}

// ExportPlantUMLToPNG converts a PlantUML file to PNG using plantuml.jar
func ExportPlantUMLToPNG(pumlFilename, pngFilename string) error {
	// Check if Java is available
	if _, err := exec.LookPath("java"); err != nil {
		return fmt.Errorf("java not found in PATH, please install Java to export PNG diagrams")
	}

	// Ensure output directory exists
	if err := ensureDir(pngFilename); err != nil {
		return fmt.Errorf("failed to prepare PNG output path: %w", err)
	}

	// Create a temporary directory for PlantUML output
	tempDir := filepath.Dir(pumlFilename)

	// Run PlantUML to convert to PNG (output to same directory as PUML file)
	cmd := exec.Command("java", "-jar", "plantuml.jar", "-tpng", pumlFilename)
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

// ensureDir ensures the directory exists for a given filename
func ensureDir(filename string) error {
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}
	return nil
}

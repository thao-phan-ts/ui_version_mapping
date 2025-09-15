# UI Version Check Tool

A comprehensive tool for analyzing lender configurations, detecting A/B testing variants, generating journey flows, and creating visual diagrams for Digital Journey projects.

## 🎯 Overview

This tool provides complete analysis capabilities for lender configurations including:
- **A/B Testing Detection**: Identifies variants and traffic distribution
- **Journey Flow Analysis**: Maps user journeys between configurations
- **Visual Diagrams**: Generates PlantUML and PNG diagrams
- **Comprehensive Reports**: Creates detailed analysis summaries

## 📋 Prerequisites

### Required Dependencies
1. **Go 1.19+**: For running the tool
2. **Java 8+**: Required for PlantUML PNG generation
3. **PlantUML JAR**: Automatically downloaded if not present

### Verify Prerequisites
```bash
# Check Go version
go version

# Check Java version
java -version

# The tool will automatically download plantuml.jar if needed
```

## 🚀 Quick Start

### Option 1: GitHub Actions (Recommended for Users)

**No local setup required!** Use the GitHub Actions workflows directly from your browser:

1. **Discover Available Options**:
   - Go to [Actions](../../actions) → "List Configuration Options"
   - Click "Run workflow" and select a config path
   - Review available lead sources and lender config IDs

2. **Run Analysis**:
   - Go to [Actions](../../actions) → "UI Version Check Tool"
   - Click "Run workflow" and fill in the parameters
   - Download the results from the workflow artifacts

📖 **[See detailed GitHub Actions guide](.github/workflows/README.md)**

### Option 2: Local Development

### 1. Build the Tool
```bash
# Build using Make
make build

# Or build manually
go build -o bin/ui-version-check ./cmd/ui-version-check
```

### 2. Setup Environment
```bash
# Auto-setup (recommended)
make setup

# Or setup specific mode
make setup-local    # Local with submodules
make setup-remote   # Remote GitHub API
```

### 3. Run Analysis
```bash
# Complete analysis with local configs
./bin/ui-version-check -config 9054 -lead-source organic

# Use remote GitHub API (no submodules needed)
./bin/ui-version-check -config 9054 -remote

# Different config paths
./bin/ui-version-check -config 9012 -config-path win

# Show help for all options
./bin/ui-version-check -help

# Or use Make shortcuts
make run-example
```

### 4. Run Tests (Legacy Scripts)
```bash
# Run all tests
make test

# Run specific tests
make test-complete
make test-ab
make test-journey
```



### 3. Check Generated Results
```bash
# View generated files
tree test_results/

# Example output structure:
# test_results/9054/
# ├── ab_testing_analysis_9054_organic.json
# ├── journey_analysis_9054_organic.json
# ├── summary_report_9054_organic.md
# ├── pumls/
# │   ├── ab_testing_groups_9054_organic.puml
# │   ├── journey_flow_9054_organic.puml
# │   └── journey_steps_*.puml
# └── images/
#     ├── ab_testing_groups_9054_organic.png
#     ├── journey_flow_9054_organic.png
#     └── journey_steps_*.png
```

## 🔍 Search and Analysis Functions

### Core Search Function
```go
// Perform complete lender config search and analysis
err := SearchLenderConfigComplete(lenderConfigID, leadSource, folderPath)
```

**Parameters:**
- `lenderConfigID`: Target configuration ID (e.g., 9054)
- `leadSource`: Lead source type (e.g., "organic", "paid")
- `folderPath`: Path to lender configs directory

### Individual Analysis Functions

#### 1. A/B Testing Analysis
```go
// Detect A/B testing groups and variants
abGroups := FindAllABTestingGroups(lenderConfigID, leadSource, folderPath)

// Export A/B testing analysis with PNG
err := ExportABTestingAnalysis(lenderConfigID, leadSource, abGroups, folderPath)
```

#### 2. Journey Analysis
```go
// Generate journey flows between configurations
err := GenerateJourneyAnalysis(lenderConfigID, leadSource, folderPath)
```

#### 3. Complete Analysis Pipeline
```go
// Run all analyses: A/B testing + Journey + Visualization + Summary
err := GenerateCompleteAnalysis(lenderConfigID, leadSource, folderPath)
```

## 📊 Generated Outputs

### 1. JSON Data Files
- **`ab_testing_analysis_*.json`**: A/B testing variants and traffic distribution
- **`journey_analysis_*.json`**: Journey flows and step sequences

### 2. PlantUML Source Files (`pumls/` directory)
- **`ab_testing_groups_*.puml`**: A/B testing diagram source
- **`journey_flow_*.puml`**: Overall journey flow diagram
- **`journey_steps_*.puml`**: Individual journey step diagrams

### 3. PNG Images (`images/` directory)
- **`ab_testing_groups_*.png`**: A/B testing visualization
- **`journey_flow_*.png`**: Journey flow diagram
- **`journey_steps_*.png`**: Detailed step-by-step diagrams

### 4. Summary Report
- **`summary_report_*.md`**: Comprehensive analysis summary

## 🎨 Visual Diagram Features

### A/B Testing Diagrams
- **Traffic Distribution**: Shows percentage split between variants
- **Variant Comparison**: Highlights differences between configurations
- **UI Version Mapping**: Details UI versions for each variant

### Journey Flow Diagrams
- **Flow Types**: Normal, rejection, automated, CIF verification flows
- **Configuration Relationships**: Visual mapping between configs
- **Decision Points**: Shows branching logic

### Journey Step Diagrams
- **Detailed Steps**: Each step with UI version information
- **Conditional Branching**: Shows different paths based on conditions
- **UI Version Priority**: Displays `sub_ui_version` prominently with `main_ui_version` as context

## 🛠️ Configuration

### Project Structure
```
ui_version_mapping/
├── cmd/ui-version-check/     # Command-line interface
├── pkg/                      # Public packages
│   ├── analyzer/            # A/B testing analysis
│   ├── config/              # Configuration types
│   ├── diagram/             # PlantUML generation
│   └── journey/             # Journey mapping
├── internal/                # Private packages
├── scripts/                 # Legacy scripts (still functional)
├── test_results/            # Generated analysis results
├── documents/               # Documentation
├── Makefile                 # Build automation
└── README.md
```

### Path Configuration
Default paths can be configured via command-line flags:
```bash
# Custom config path
./bin/ui-version-check -config-path ./my-configs

# Custom output path  
./bin/ui-version-check -output ./my-results
```

### Output Directory Structure
```
test_results/
└── <lender_config_id>/
    ├── *.json                    # Analysis data
    ├── *.md                      # Summary reports
    ├── pumls/
    │   └── *.puml               # PlantUML source files
    └── images/
        └── *.png                # Generated PNG diagrams
```

## 🖥️ Command Line Interface

### Basic Usage
```bash
# Complete analysis (default mode)
./bin/ui-version-check -config 9054 -lead-source organic

# A/B testing analysis only
./bin/ui-version-check -config 9054 -mode ab-testing

# Journey analysis only  
./bin/ui-version-check -config 9054 -mode journey

# Custom paths
./bin/ui-version-check \
  -config 9054 \
  -config-path ./my-configs \
  -output ./my-results
```

### Available Options
- `-config <id>`: Lender config ID to analyze (default: 9054)
- `-lead-source <src>`: Lead source type (default: "organic")  
- `-config-path <path>`: Path to lender configs directory
- `-output <path>`: Output directory for results
- `-mode <mode>`: Analysis mode (complete, ab-testing, journey)
- `-help`: Show help message

## 📝 Usage Examples

### Example 1: Analyze Specific Configuration
```go
func main() {
    lenderConfigID := 9054
    leadSource := "organic"
    folderPath := "../digital_journey/migration/sync/vietnam/tpbank/lender_configs/evo"
    
    err := SearchLenderConfigComplete(lenderConfigID, leadSource, folderPath)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Analysis complete! Check test_results/ directory")
}
```

### Example 2: Generate Only A/B Testing Analysis
```go
func analyzeABTesting() {
    lenderConfigID := 9054
    leadSource := "organic"
    folderPath := "../digital_journey/migration/sync/vietnam/tpbank/lender_configs/evo"
    
    // Find A/B testing groups
    abGroups := FindAllABTestingGroups(lenderConfigID, leadSource, folderPath)
    
    // Export with PNG generation
    err := ExportABTestingAnalysis(lenderConfigID, leadSource, abGroups, folderPath)
    if err != nil {
        log.Fatal(err)
    }
}
```

### Example 3: Custom Journey Analysis
```go
func customJourneyAnalysis() {
    // Generate journey template
    template, err := GenerateJourneyTemplate(9054, relatedConfigs, folderPath)
    if err != nil {
        log.Fatal(err)
    }
    
    // Export individual journey diagrams
    err = ExportAllJourneysPlantUML(template, 9054, "organic")
    if err != nil {
        log.Fatal(err)
    }
}
```

## 🧪 Testing

### Using Make (Recommended)
```bash
# Run all tests
make test

# Run specific tests
make test-complete    # Complete search functionality
make test-ab         # A/B testing detection  
make test-journey    # Journey generation

# Code quality checks
make check           # Run fmt, vet, lint, and test
make lint           # Run golangci-lint
make fmt            # Format code
make vet            # Run go vet
```

### Manual Testing (Legacy)
```bash
# Navigate to scripts directory
cd scripts

# Run specific tests
go test -v -run TestCompleteSearch
go test -v -run TestABTestingAnalysis
go test -v -run TestIndividualJourneyExport
```

### Test Output Verification
The tests automatically verify:
- ✅ JSON files are created with valid data
- ✅ PlantUML files are generated with correct syntax
- ✅ PNG files are exported successfully
- ✅ Summary reports contain expected sections

## 🎨 PlantUML Theming

The tool uses a custom "Materia" theme for all diagrams with:
- **Professional Color Palette**: Blue, green, red semantic colors
- **Modern Styling**: Rounded corners, clean typography
- **Semantic Coloring**: 
  - 🔵 Blue: Normal flows
  - 🟢 Green: Success paths
  - 🔴 Red: Rejection flows
  - 🟡 Yellow: Warning states

## 🔧 Troubleshooting

### Common Issues

#### 1. Java Not Found
```
Error: java not found in PATH
```
**Solution**: Install Java 8+ and ensure it's in your PATH

#### 2. PlantUML Errors
```
Some diagram description contains errors
```
**Solution**: Check the generated `.puml` files for syntax issues

#### 3. File Path Issues
```
Found 0 configs in path
```
**Solution**: Verify the `folderPath` parameter points to correct lender configs directory

#### 4. Permission Issues
```
failed to write file
```
**Solution**: Ensure write permissions for the `test_results/` directory

### Debug Mode
Enable verbose output by checking the test logs:
```bash
go test -v -run TestCompleteSearch 2>&1 | tee debug.log
```

## 📚 API Reference

### Core Functions

#### `SearchLenderConfigComplete(lenderConfigID int, leadSource string, folderPath string) error`
Performs complete search and analysis pipeline.

#### `FindAllABTestingGroups(lenderConfigID int, leadSource string, folderPath string) []ABTestingGroup`
Detects A/B testing variants and groups.

#### `GenerateJourneyAnalysis(lenderConfigID int, leadSource string, folderPath string) error`
Generates journey flow analysis.

#### `ExportABTestingAnalysis(lenderConfigID int, leadSource string, abGroups []ABTestingGroup, folderPath string) error`
Exports A/B testing analysis with PNG generation.

### Data Structures

#### `ABTestingGroup`
```go
type ABTestingGroup struct {
    GroupName    string              `json:"group_name"`
    Variants     []ABTestingVariant  `json:"variants"`
    TotalWeight  int                 `json:"total_weight"`
}
```

#### `Journey`
```go
type Journey struct {
    ID          string `json:"id"`
    FlowType    string `json:"flow_type"`
    Description string `json:"description"`
    Steps       []Step `json:"steps"`
}
```

#### `Step`
```go
type Step struct {
    StepNumber               int                        `json:"step_number"`
    StepName                 string                     `json:"step_name"`
    MainUIVersion           string                     `json:"main_ui_version"`
    SubUIVersion            string                     `json:"sub_ui_version,omitempty"`
    SubUIVersionByConditions []SubUIVersionByCondition `json:"sub_ui_version_by_conditions,omitempty"`
}
```

## 🤝 Contributing

1. **Fork the repository**
2. **Create feature branch**: `git checkout -b feature/new-analysis`
3. **Add tests**: Ensure new functionality is tested
4. **Update README**: Document new features
5. **Submit pull request**

## 📄 License

This tool is part of the Digital Journey testing framework.

---

**📞 Support**: For issues or questions, please check the troubleshooting section or contact the development team. 
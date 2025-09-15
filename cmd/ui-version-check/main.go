package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/tsocial/ui-version-mapping/pkg/analyzer"
	"github.com/tsocial/ui-version-mapping/pkg/config"
)

const (
	// Default paths
	DefaultConfigPath = "evo"
	DefaultOutputPath = "../../test_results"
)

func main() {
	var (
		configID   = flag.Int("config", 9054, "Lender config ID to analyze")
		leadSource = flag.String("lead-source", "organic", "Lead source (organic, paid, etc.)")
		configPath = flag.String("config-path", DefaultConfigPath, "Path to lender configs directory")
		outputPath = flag.String("output", DefaultOutputPath, "Output directory for results")
		mode       = flag.String("mode", "complete", "Analysis mode: complete, ab-testing, journey")
		remote     = flag.Bool("remote", false, "Use remote GitHub API instead of local files")
		help       = flag.Bool("help", false, "Show help message")
	)

	flag.Parse()

	if *help {
		showHelp()
		return
	}

	// Validate inputs
	if *configID <= 0 {
		log.Fatal("Config ID must be a positive integer")
	}

	if *leadSource == "" {
		log.Fatal("Lead source cannot be empty")
	}

	// Ensure output directory exists
	if err := os.MkdirAll(*outputPath, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	fmt.Printf("🔍 UI Version Check Tool\n")
	fmt.Printf("========================\n")
	fmt.Printf("Config ID: %d\n", *configID)
	fmt.Printf("Lead Source: %s\n", *leadSource)
	fmt.Printf("Config Path: %s\n", *configPath)
	fmt.Printf("Output Path: %s\n", *outputPath)
	fmt.Printf("Mode: %s\n", *mode)
	fmt.Printf("Remote: %t\n\n", *remote)

	// Create config provider
	var provider config.ConfigProvider
	if *remote {
		// Use remote GitHub API
		baseURL := os.Getenv("CONFIG_REMOTE_URL")
		if baseURL == "" {
			baseURL = "https://api.github.com/repos/tsocial/digital_journey"
		}
		token := os.Getenv("GITHUB_TOKEN")
		provider = config.NewRemoteConfigProvider(baseURL, token)
		fmt.Printf("Using remote config provider: %s\n", baseURL)
	} else {
		// Use smart provider selection
		provider = config.GetConfigProvider()
		fmt.Printf("Using automatic config provider\n")
	}

	// Create analyzer service
	analyzerService := analyzer.NewAnalyzerService(provider)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Run analysis based on mode
	switch *mode {
	case "ab-testing":
		err := runABTestingAnalysis(ctx, analyzerService, *configID, *leadSource, *configPath, *outputPath)
		if err != nil {
			log.Fatalf("A/B testing analysis failed: %v", err)
		}
	case "journey":
		err := runJourneyAnalysis(ctx, analyzerService, *configID, *leadSource, *configPath, *outputPath)
		if err != nil {
			log.Fatalf("Journey analysis failed: %v", err)
		}
	case "complete":
		err := runCompleteAnalysis(ctx, analyzerService, *configID, *leadSource, *configPath, *outputPath)
		if err != nil {
			log.Fatalf("Complete analysis failed: %v", err)
		}
	default:
		log.Fatalf("Unknown mode: %s", *mode)
	}

	fmt.Printf("\n🎉 Analysis completed successfully!\n")
}

func runABTestingAnalysis(ctx context.Context, service *analyzer.AnalyzerService, configID int, leadSource, configPath, outputPath string) error {
	fmt.Printf("=== Running A/B Testing Analysis ===\n")

	// Find A/B testing groups
	groups, err := service.FindABTestingGroups(ctx, configPath)
	if err != nil {
		return fmt.Errorf("failed to find A/B testing groups: %w", err)
	}

	fmt.Printf("Found %d A/B testing groups\n", len(groups))

	// Export results
	outputDir := filepath.Join(outputPath, fmt.Sprintf("%d", configID))
	filename := filepath.Join(outputDir, fmt.Sprintf("ab_testing_analysis_%d_%s.json", configID, leadSource))

	// TODO: Implement export logic using the new service
	fmt.Printf("Results would be exported to: %s\n", filename)

	return nil
}

func runJourneyAnalysis(ctx context.Context, service *analyzer.AnalyzerService, configID int, leadSource, configPath, outputPath string) error {
	fmt.Printf("=== Running Journey Analysis ===\n")

	// Find related configs
	relatedConfigs, err := service.SearchRelatedConfigs(ctx, configID, leadSource, configPath)
	if err != nil {
		return fmt.Errorf("failed to find related configs: %w", err)
	}

	fmt.Printf("Found %d related configs\n", len(relatedConfigs))

	// TODO: Implement journey generation using the new service
	outputDir := filepath.Join(outputPath, fmt.Sprintf("%d", configID))
	filename := filepath.Join(outputDir, fmt.Sprintf("journey_analysis_%d_%s.json", configID, leadSource))
	fmt.Printf("Results would be exported to: %s\n", filename)

	return nil
}

func runCompleteAnalysis(ctx context.Context, service *analyzer.AnalyzerService, configID int, leadSource, configPath, outputPath string) error {
	fmt.Printf("=== Running Complete Analysis ===\n")

	// Run A/B testing analysis
	if err := runABTestingAnalysis(ctx, service, configID, leadSource, configPath, outputPath); err != nil {
		return fmt.Errorf("A/B testing analysis failed: %w", err)
	}

	// Run journey analysis
	if err := runJourneyAnalysis(ctx, service, configID, leadSource, configPath, outputPath); err != nil {
		return fmt.Errorf("journey analysis failed: %w", err)
	}

	return nil
}

func showHelp() {
	fmt.Printf(`UI Version Check Tool - Enhanced Version

USAGE:
    ui-version-check [OPTIONS]

OPTIONS:
    -config <id>        Lender config ID to analyze (default: 9054)
    -lead-source <src>  Lead source type (default: "organic")
    -config-path <path> Path to lender configs directory (default: "evo")
    -output <path>      Output directory for results (default: "../../test_results")
    -mode <mode>        Analysis mode: complete, ab-testing, journey (default: "complete")
    -remote             Use remote GitHub API instead of local files
    -help               Show this help message

EXAMPLES:
    # Complete analysis with local files
    ui-version-check -config 9054 -lead-source organic

    # A/B testing analysis only
    ui-version-check -config 9054 -mode ab-testing

    # Use remote GitHub API
    ui-version-check -config 9054 -remote

    # Custom paths
    ui-version-check -config 9054 -config-path win -output ./results

ENVIRONMENT VARIABLES:
    CONFIG_REMOTE_URL   GitHub API base URL (default: https://api.github.com/repos/tsocial/digital_journey)
    GITHUB_TOKEN        GitHub token for API access (optional for public repos)

MODES:
    complete    - Full analysis including A/B testing, journey mapping, and visualization
    ab-testing  - A/B testing detection and analysis only
    journey     - Journey flow analysis and visualization only

FEATURES:
    ✅ Smart config provider selection (local/remote)
    ✅ No dependency on submodules or auto_sync
    ✅ GitHub API integration for remote configs
    ✅ Environment-based configuration
    ✅ Optimized service architecture

OUTPUT:
    The tool generates JSON data files, PlantUML diagrams, PNG images, and summary reports
    in the specified output directory.
`)
}

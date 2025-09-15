package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ConfigProvider interface cho việc load configs từ local filesystem
type ConfigProvider interface {
	LoadConfigs(ctx context.Context, path string) ([]*LenderConfig, error)
	LoadConfig(ctx context.Context, configID int, leadSource string) (*LenderConfig, error)
}

// LocalConfigProvider - load từ local filesystem
type LocalConfigProvider struct {
	BasePath string
}

// NewLocalConfigProvider tạo local provider
func NewLocalConfigProvider(basePath string) *LocalConfigProvider {
	return &LocalConfigProvider{
		BasePath: basePath,
	}
}

// LoadConfigs từ local filesystem
func (p *LocalConfigProvider) LoadConfigs(ctx context.Context, path string) ([]*LenderConfig, error) {
	var configs []*LenderConfig

	fullPath := filepath.Join(p.BasePath, path)

	err := filepath.Walk(fullPath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		// Skip archive directories
		if info.IsDir() && strings.Contains(strings.ToLower(info.Name()), "archive") {
			return filepath.SkipDir
		}

		// Process JSON files
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".json") {
			config, err := p.loadConfigFile(filePath)
			if err == nil && config != nil {
				configs = append(configs, config)
			}
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan configs from %s: %w", fullPath, err)
	}

	return configs, nil
}

// LoadConfig từ local filesystem
func (p *LocalConfigProvider) LoadConfig(ctx context.Context, configID int, leadSource string) (*LenderConfig, error) {
	// Search pattern: *{configID}*.json
	var foundConfig *LenderConfig

	err := filepath.Walk(p.BasePath, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if !info.IsDir() && strings.Contains(info.Name(), fmt.Sprintf("%d", configID)) && strings.HasSuffix(info.Name(), ".json") {
			config, err := p.loadConfigFile(filePath)
			if err == nil && config != nil && config.ID == configID {
				// Check lead source if specified
				if leadSource != "" {
					hasLeadSource := false
					for _, tag := range config.Tags {
						if tag.Name == "lead_source" && tag.Value == leadSource {
							hasLeadSource = true
							break
						}
					}
					if !hasLeadSource {
						return nil // Continue searching
					}
				}
				foundConfig = config
				return filepath.SkipAll // Found, stop searching
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to search for config %d: %w", configID, err)
	}

	if foundConfig == nil {
		return nil, fmt.Errorf("config %d not found", configID)
	}

	return foundConfig, nil
}

func (p *LocalConfigProvider) loadConfigFile(filePath string) (*LenderConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var config LenderConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetConfigProvider tạo provider dựa trên environment - chỉ sử dụng local files
func GetConfigProvider() ConfigProvider {
	// Check if vendor configs exist (preferred)
	if _, err := os.Stat("vendor/configs"); err == nil {
		return NewLocalConfigProvider("vendor/configs")
	}

	// Check if submodules exist
	if _, err := os.Stat("scripts/submodules/digital_journey"); err == nil {
		return NewLocalConfigProvider("scripts/submodules/digital_journey/migration/sync/vietnam/tpbank/lender_configs")
	}

	// Default to vendor structure
	return NewLocalConfigProvider("vendor/configs")
}

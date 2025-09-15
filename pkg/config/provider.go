package config

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ConfigProvider interface cho việc load configs từ nhiều nguồn
type ConfigProvider interface {
	LoadConfigs(ctx context.Context, path string) ([]*LenderConfig, error)
	LoadConfig(ctx context.Context, configID int, leadSource string) (*LenderConfig, error)
}

// LocalConfigProvider - load từ local filesystem
type LocalConfigProvider struct {
	BasePath string
}

// RemoteConfigProvider - load từ GitHub API
type RemoteConfigProvider struct {
	BaseURL    string
	Token      string // GitHub token nếu cần
	CacheDir   string
	CacheTTL   time.Duration
	httpClient *http.Client
}

// CachedConfigProvider - wrapper với cache
type CachedConfigProvider struct {
	Provider ConfigProvider
	CacheDir string
	TTL      time.Duration
}

// NewLocalConfigProvider tạo local provider
func NewLocalConfigProvider(basePath string) *LocalConfigProvider {
	return &LocalConfigProvider{
		BasePath: basePath,
	}
}

// NewRemoteConfigProvider tạo remote provider
func NewRemoteConfigProvider(baseURL, token string) *RemoteConfigProvider {
	return &RemoteConfigProvider{
		BaseURL:    baseURL,
		Token:      token,
		CacheDir:   ".cache/configs",
		CacheTTL:   time.Hour * 24, // Cache 24h
		httpClient: &http.Client{Timeout: 30 * time.Second},
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

// LoadConfigs từ GitHub API
func (p *RemoteConfigProvider) LoadConfigs(ctx context.Context, path string) ([]*LenderConfig, error) {
	// GitHub API URL format: https://api.github.com/repos/owner/repo/contents/path
	// Build full path: migration/sync/vietnam/tpbank/lender_configs/{path}
	fullPath := fmt.Sprintf("migration/sync/vietnam/tpbank/lender_configs/%s", path)

	return p.loadConfigsRecursive(ctx, fullPath)
}

// loadConfigsRecursive loads configs recursively from a directory
func (p *RemoteConfigProvider) loadConfigsRecursive(ctx context.Context, path string) ([]*LenderConfig, error) {
	url := fmt.Sprintf("%s/contents/%s", p.BaseURL, path)

	// Get directory listing
	files, err := p.fetchDirectoryListing(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch directory listing for %s: %w", path, err)
	}

	var configs []*LenderConfig
	for _, file := range files {
		if file.Type == "file" && strings.HasSuffix(file.Name, ".json") {
			// Load JSON file
			config, err := p.fetchConfigFile(ctx, file.DownloadURL)
			if err != nil {
				continue // Skip errors
			}
			if config != nil {
				configs = append(configs, config)
			}
		} else if file.Type == "dir" && file.Name != "archive" {
			// Recursively load from subdirectory
			subConfigs, err := p.loadConfigsRecursive(ctx, fmt.Sprintf("%s/%s", path, file.Name))
			if err != nil {
				continue // Skip errors in subdirectories
			}
			configs = append(configs, subConfigs...)
		}
	}

	return configs, nil
}

// LoadConfig từ GitHub API
func (p *RemoteConfigProvider) LoadConfig(ctx context.Context, configID int, leadSource string) (*LenderConfig, error) {
	// Search in different subdirectories
	searchPaths := []string{"evo", "win", "evo_native", "winback"}

	for _, subPath := range searchPaths {
		configs, err := p.LoadConfigs(ctx, subPath)
		if err != nil {
			continue // Try next path
		}

		// Find matching config
		for _, cfg := range configs {
			if cfg.ID == configID {
				// Check lead source if specified
				if leadSource != "" {
					hasLeadSource := false
					for _, tag := range cfg.Tags {
						if tag.Name == "lead_source" && tag.Value == leadSource {
							hasLeadSource = true
							break
						}
					}
					if !hasLeadSource {
						continue // Try next config
					}
				}
				return cfg, nil
			}
		}
	}

	return nil, fmt.Errorf("config %d not found in remote repository", configID)
}

type GitHubFile struct {
	Name        string `json:"name"`
	DownloadURL string `json:"download_url"`
	Type        string `json:"type"`
}

func (p *RemoteConfigProvider) fetchDirectoryListing(ctx context.Context, url string) ([]GitHubFile, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if p.Token != "" {
		req.Header.Set("Authorization", "token "+p.Token)
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close() // Ignore close error
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var files []GitHubFile
	if err := json.NewDecoder(resp.Body).Decode(&files); err != nil {
		return nil, err
	}

	return files, nil
}

func (p *RemoteConfigProvider) fetchConfigFile(ctx context.Context, url string) (*LenderConfig, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = resp.Body.Close() // Ignore close error
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch config file: status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var config LenderConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetConfigProvider tạo provider dựa trên environment
func GetConfigProvider() ConfigProvider {
	// Check environment variables
	if remoteURL := os.Getenv("CONFIG_REMOTE_URL"); remoteURL != "" {
		token := os.Getenv("GITHUB_TOKEN")
		return NewRemoteConfigProvider(remoteURL, token)
	}

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

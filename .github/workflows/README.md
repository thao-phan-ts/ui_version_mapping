# GitHub Actions for UI Version Check Tool

This directory contains GitHub Action workflows that allow you to run the UI Version Check Tool directly from the GitHub interface without needing to set up a local development environment.

## 🚀 Available Workflows

### 1. UI Version Check (`ui-version-check.yml`)

**Purpose**: Run comprehensive UI version analysis on lender configurations.

**How to use**:
1. Go to the [Actions tab](../../actions) in your repository
2. Select "UI Version Check Tool" from the workflow list
3. Click "Run workflow"
4. Fill in the required parameters:
   - **Lead Source**: Choose from organic, paid, referral, direct
   - **Config Path**: Choose the configuration directory (evo, win, evo_native, winback)
   - **Lender Config ID**: Enter the specific configuration ID to analyze
   - **Analysis Mode**: Choose complete, ab-testing, or journey
   - **Use Remote**: Whether to use GitHub API (recommended) or local files

**Outputs**:
- Complete analysis results including JSON data, PlantUML diagrams, and PNG images
- Summary report with key findings
- Downloadable artifacts with all generated files

### 2. List Configuration Options (`list-config-options.yml`)

**Purpose**: Discover available lead sources and lender config IDs before running analysis.

**How to use**:
1. Go to the [Actions tab](../../actions) in your repository
2. Select "List Configuration Options" from the workflow list
3. Click "Run workflow"
4. Choose the config path you want to scan
5. The workflow will display all available options in the logs

**Outputs**:
- List of available lead sources with counts
- List of available lender config IDs with names
- Usage examples
- Downloadable summary report

## 📋 Step-by-Step Guide

### First Time Setup

1. **Discover Available Options**:
   ```
   Actions → List Configuration Options → Run workflow
   ```
   - Choose config path (start with "evo")
   - Use remote: true (recommended)
   - Review the output to see available lead sources and config IDs

2. **Run Analysis**:
   ```
   Actions → UI Version Check Tool → Run workflow
   ```
   - Use the lead sources and config IDs discovered in step 1
   - Start with "complete" mode for full analysis

### Example Workflow

1. **List options for "evo" configs**:
   - Workflow: List Configuration Options
   - Config Path: `evo`
   - Use Remote: `true`
   - Result: See available lead sources (organic, paid, etc.) and config IDs (9054, 9012, etc.)

2. **Analyze config 9054 with organic lead source**:
   - Workflow: UI Version Check Tool
   - Lead Source: `organic`
   - Config Path: `evo`
   - Lender Config ID: `9054`
   - Analysis Mode: `complete`
   - Use Remote: `true`

3. **Download results**:
   - Go to the workflow run page
   - Download the artifact: `ui-version-analysis-results-9054-organic`
   - Extract and review the generated files

## 🔧 Configuration

### Environment Variables

The workflows use these environment variables (automatically configured):

- `GITHUB_TOKEN`: Automatically provided by GitHub Actions for API access
- `CONFIG_REMOTE_URL`: Set to `https://api.github.com/repos/tsocial/digital_journey`

### Permissions

The workflows require:
- `contents: read` - To checkout the repository
- `actions: read` - To access workflow artifacts
- Standard GitHub token permissions for API access

## 📊 Understanding the Results

### Generated Files

When you download the analysis artifacts, you'll find:

```
results/
└── {config_id}/
    ├── ab_testing_analysis_{config_id}_{lead_source}.json
    ├── journey_analysis_{config_id}_{lead_source}.json
    ├── summary_report_{config_id}_{lead_source}.md
    ├── pumls/
    │   ├── ab_testing_groups_{config_id}_{lead_source}.puml
    │   ├── journey_flow_{config_id}_{lead_source}.puml
    │   └── journey_steps_*.puml
    └── images/
        ├── ab_testing_groups_{config_id}_{lead_source}.png
        ├── journey_flow_{config_id}_{lead_source}.png
        └── journey_steps_*.png
```

### Key Files to Review

1. **`summary_report_*.md`**: Start here for an overview of findings
2. **`ab_testing_analysis_*.json`**: Detailed A/B testing variant information
3. **`journey_analysis_*.json`**: Journey flow and step analysis
4. **`images/*.png`**: Visual diagrams of the analysis results

## 🐛 Troubleshooting

### Common Issues

1. **"Config ID not found"**:
   - Run "List Configuration Options" first to see available IDs
   - Make sure you're using the correct config path
   - Verify the config ID exists in the selected path

2. **"No lead sources found"**:
   - The configuration might not have lead source tags
   - Try a different config path
   - Check if you're using the right remote repository

3. **"Analysis failed"**:
   - Check the workflow logs for detailed error messages
   - Verify your inputs are correct
   - Try with a simpler analysis mode (e.g., "ab-testing" instead of "complete")

4. **"Remote API timeout"**:
   - The GitHub API might be slow or rate-limited
   - Try again later
   - Consider using a different config path with fewer files

### Getting Help

1. Check the workflow logs for detailed error messages
2. Review the troubleshooting section in the main README
3. Run "List Configuration Options" to verify available choices
4. Try with different input parameters

## 🔄 Workflow Updates

These workflows are automatically updated when you push changes to the repository. The workflows will:

1. Use the latest version of the UI Version Check Tool
2. Automatically download dependencies
3. Cache build artifacts for faster execution
4. Generate results using the most current configuration data

## 📈 Performance Tips

1. **Use Remote API**: Generally faster and more reliable than local files
2. **Start Small**: Begin with "ab-testing" or "journey" mode before running "complete"
3. **Cache Awareness**: Subsequent runs are faster due to Go module caching
4. **Artifact Management**: Download artifacts promptly as they expire after 30 days

## 🔐 Security Notes

- The workflows use read-only access to the configuration repository
- GitHub tokens are automatically managed and scoped appropriately
- No sensitive data is stored in workflow artifacts
- All API calls use HTTPS and proper authentication 
package config

// DecisionEngine represents a decision engine configuration
type DecisionEngine struct {
	TreeUUID          string   `json:"tree_uuid"`
	CreditTreeUUID    string   `json:"credit_tree_uuid,omitempty"`
	RiskGradeTreeUUID string   `json:"risk_grade_tree_uuid,omitempty"`
	EvaluationType    string   `json:"evaluation_type"`
	MaxWaitSeconds    int      `json:"max_wait_seconds"`
	UseAddOnServices  []string `json:"use_add_on_services,omitempty"`
}

// Tag represents a tag with name and value
type Tag struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// LenderConfig represents the structure of a lender configuration JSON file
type LenderConfig struct {
	ID              int                       `json:"id"`
	Name            string                    `json:"name"`
	Tags            []Tag                     `json:"tags"`
	UIVersion       string                    `json:"ui_version"`
	UIFlow          []string                  `json:"ui_flow"`
	UIFlowSettings  map[string]interface{}    `json:"ui_flow_settings"`
	DecisionEngines map[string]DecisionEngine `json:"decision_engines,omitempty"`
	Weight          int                       `json:"weight"`
}

// ConfigInfo represents processed configuration information
type ConfigInfo struct {
	File           string
	ConfigID       int
	Name           string
	UIVersion      string
	UIFlow         []string
	UIFlowSettings map[string]interface{}
}

// RelatedConfigResult represents the result of finding related configs
type RelatedConfigResult struct {
	ConfigID       int    `json:"config_id"`
	Name           string `json:"name"`
	FlowType       string `json:"flow_type"`
	UIVersion      string `json:"ui_version"`
	Weight         int    `json:"weight"`
	MatchReason    string `json:"match_reason"`
	MatchedTags    []Tag  `json:"matched_tags,omitempty"`
	DecisionUUID   string `json:"decision_uuid,omitempty"`
	IsABTesting    bool   `json:"is_ab_testing,omitempty"`
	ABTestingGroup string `json:"ab_testing_group,omitempty"`
	ABVariants     []int  `json:"ab_variants,omitempty"`
}

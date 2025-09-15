package ui_version_check

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

// CSVRow represents a row in the CSV test_results
type CSVRow struct {
	STT               int
	LenderConfigID    string
	FlowType          string
	UIVersion         string
	UIFlow            string
	SubUIVersion      string
	SubUIVersionState string
	ActiveStatus      string
	TreeUUID          string
	Weight            string
	Path              string
}

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

// StepInfo represents information about a UI flow step
type StepInfo struct {
	Name           string `json:"name"`
	UIVersion      string `json:"ui_version"`
	SubVersion     string `json:"sub_version,omitempty"`
	Purpose        string `json:"purpose"`
	IsDecision     bool   `json:"is_decision"`
	TreeUUID       string `json:"tree_uuid,omitempty"`
	RiskTreeUUID   string `json:"risk_tree_uuid,omitempty"`
	CreditTreeUUID string `json:"credit_tree_uuid,omitempty"`
}

// DecisionOutcome represents a possible outcome from a decision step
type DecisionOutcome struct {
	Name        string `json:"name"`
	Condition   string `json:"condition"`
	NextStep    string `json:"next_step"`
	ConfigID    string `json:"config_id,omitempty"`
	UIVersion   string `json:"ui_version,omitempty"`
	Description string `json:"description"`
}

// DetailedFlowConfig represents the complete configuration for generating detailed flow puml
type DetailedFlowConfig struct {
	ConfigID      int                          `json:"config_id"`
	Name          string                       `json:"name"`
	UIVersion     string                       `json:"ui_version"`
	FlowType      string                       `json:"flow_type"`
	Steps         []StepInfo                   `json:"steps"`
	DecisionLogic map[string][]DecisionOutcome `json:"decision_logic"`
}

// LenderConfigSearchResult represents the result of searching configurations by lender_id
type LenderConfigSearchResult struct {
	SearchValue      int64             `json:"search_value"`
	SearchType       string            `json:"search_type"`
	FlowRouting      []FlowRoutingInfo `json:"flow_routing"`
	RelatedConfigIDs []int64           `json:"related_config_ids"`
	FlowConfigs      []FlowConfigInfo  `json:"flow_configs"`
	JourneyPath      []JourneyStep     `json:"journey_path"`
}

// FlowRoutingInfo represents information about flow routing from one lender to another
type FlowRoutingInfo struct {
	FromLenderID int64  `json:"from_lender_id"`
	ToLenderID   int64  `json:"to_lender_id"`
	ConfigID     int    `json:"config_id"`
	FlowType     string `json:"flow_type"`
	Condition    string `json:"condition"`
	DecisionStep string `json:"decision_step"`
	Description  string `json:"description"`
}

// FlowConfigInfo represents configuration information for flow routing
type FlowConfigInfo struct {
	ConfigID  int    `json:"config_id"`
	Name      string `json:"name"`
	LenderID  int64  `json:"lender_id"`
	UIVersion string `json:"ui_version"`
	FlowType  string `json:"flow_type"`
	File      string `json:"file"`
}

// JourneyStep represents a step in the journey from initial config to target lender
type JourneyStep struct {
	StepOrder    int    `json:"step_order"`
	LenderID     int64  `json:"lender_id"`
	ConfigID     int    `json:"config_id"`
	ConfigName   string `json:"config_name"`
	UIVersion    string `json:"ui_version"`
	FlowType     string `json:"flow_type"`
	DecisionStep string `json:"decision_step,omitempty"`
	Condition    string `json:"condition,omitempty"`
}

// RealConfig represents the actual structure of lender config files
type RealConfig struct {
	ID              int64                  `json:"id"`
	Name            string                 `json:"name"`
	LenderID        int64                  `json:"lender_id"`
	UIVersion       string                 `json:"ui_version"`
	UIFlow          []string               `json:"ui_flow"`
	UIFlowSettings  map[string]interface{} `json:"ui_flow_settings"`
	DecisionEngines map[string]interface{} `json:"decision_engines"`
	Active          bool                   `json:"active"`
}

type SearchResult struct {
	SearchValue    interface{} `json:"search_value"`
	SearchType     string      `json:"search_type"`
	RelatedConfigs []int       `json:"related_config_ids"`
	Journeys       []*Journey  `json:"journeys"`
}

type Journey struct {
	ID                 string `json:"id"`
	FlowType           string `json:"flow_type"`
	FromLenderConfigID int    `json:"from_lender_config_id"`
	ToLenderConfigID   int    `json:"to_lender_config_id"`
	Active             bool   `json:"active"`
	Condition          string `json:"condition"`
	Description        string `json:"description"`
	Steps              []Step `json:"steps"`
}

type Step struct {
	ID                       int                       `json:"id"`
	Name                     string                    `json:"name"`
	MainUIVersion            string                    `json:"main_ui_version"`
	SubUIVersion             string                    `json:"sub_ui_version"`
	SubUIVersionByConditions []SubUIVersionByCondition `json:"sub_ui_version_by_conditions"`
}

type SubUIVersionByCondition struct {
	Condition    string `json:"condition"`
	SubUIVersion string `json:"sub_ui_version"`
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

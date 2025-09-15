package journey

// Journey represents a user journey between configurations
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

// Step represents a step in a journey
type Step struct {
	ID                       int                       `json:"id"`
	Name                     string                    `json:"name"`
	MainUIVersion            string                    `json:"main_ui_version"`
	SubUIVersion             string                    `json:"sub_ui_version"`
	SubUIVersionByConditions []SubUIVersionByCondition `json:"sub_ui_version_by_conditions"`
}

// SubUIVersionByCondition represents conditional UI version logic
type SubUIVersionByCondition struct {
	Condition    string `json:"condition"`
	SubUIVersion string `json:"sub_ui_version"`
}

// JourneyTemplate represents the template structure for journey generation
type JourneyTemplate struct {
	SearchValue      int64     `json:"search_value"`
	SearchType       string    `json:"search_type"`
	RelatedConfigIDs []int     `json:"related_config_ids"`
	Journeys         []Journey `json:"journeys"`
}

// SearchResult represents the result of a search operation
type SearchResult struct {
	SearchValue    interface{} `json:"search_value"`
	SearchType     string      `json:"search_type"`
	RelatedConfigs []int       `json:"related_config_ids"`
	Journeys       []*Journey  `json:"journeys"`
}

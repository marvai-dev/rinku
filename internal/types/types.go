package types

type LibsFile struct {
	Libs map[string]Library `json:"libs"`
}

type Library struct {
	URL    string `json:"url"`
	Lang   string `json:"lang"`
	Unsafe string `json:"unsafe,omitempty"`
}

type MappingsFile struct {
	Mappings []Mapping `json:"mappings"`
}

type Mapping struct {
	Source     string   `json:"source"`
	Targets    []string `json:"targets"`
	Category   string   `json:"category,omitempty"`
	Confidence float64  `json:"confidence,omitempty"`
}

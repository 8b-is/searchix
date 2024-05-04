package server

type NixValue struct {
	Type string `json:"_type"`
	Text string `json:"text"`
}

type Option struct {
	Declarations []string
	Default      NixValue
	Description  string
	Example      NixValue
	ReadOnly     bool
	Type         string `json:"type"`
}

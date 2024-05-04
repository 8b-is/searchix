package server

type NixValue struct {
	Kind  string `json:"_type"`
	Value string `json:"text"`
}

type Option struct {
	Declarations []string
	Default      NixValue
	Description  string
	Example      NixValue
	ReadOnly     bool
	Kind         string `json:"type"`
}

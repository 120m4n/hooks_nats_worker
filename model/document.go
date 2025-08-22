package model

type HookEventMetadata struct {
	Source    string `json:"source"`
	Version   string `json:"version"`
	Processed string `json:"processed"`
}

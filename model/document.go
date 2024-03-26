package model

type Document struct {
	UniqueId     string    `json:"unique_id"`
	UserId       string    `json:"user_id"`
	Fleet        string    `json:"fleet"`
	Location     MongoLocation  `json:"location"`
	OriginIp     string    `json:"ip_origin"`
	LastModified int64     `json:"last_modified"`
}

type MongoLocation struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}
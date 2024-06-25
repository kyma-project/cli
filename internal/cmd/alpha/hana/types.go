package hana

type HanaAPIParameters struct {
	TechnicalUser bool `json:"technicalUser"`
}

type HanaMapping struct {
	Platform  string `json:"platform"`
	PrimaryID string `json:"primaryID"`
}

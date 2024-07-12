package hana

type HanaAPIParameters struct {
	TechnicalUser bool `json:"technicalUser"`
}

type HanaInstanceParameters struct {
	Data HanaInstanceParametersData `json:"data"`
}

type HanaInstanceParametersData struct {
	Memory                 int      `json:"memory"`
	Vcpu                   int      `json:"vcpu"`
	WhitelistIPs           []string `json:"whitelistIPs"`
	GenerateSystemPassword bool     `json:"generateSystemPassword"`
	Edition                string   `json:"edition"`
}

type HanaBindingParameters struct {
	Scope           string `json:"scope"`
	CredentialsType string `json:"credential-type"`
}

type HanaMapping struct {
	Platform  string `json:"platform"`
	PrimaryID string `json:"primaryID"`
}

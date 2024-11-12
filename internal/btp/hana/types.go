package hana

import "github.com/kyma-project/cli.v3/internal/btp/auth"

type HanaInstance struct {
	Data []HanaInstanceData `json:"data"`
}

type HanaInstanceData struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	ServicePlanName string `json:"servicePlanName"`
}

type HanaAdminCredentials struct {
	BaseURL string   `json:"baseurl"`
	UAA     auth.UAA `json:"uaa"`
}

type HanaMapping struct {
	Platform  string `json:"platform"`
	PrimaryID string `json:"primaryID"`
}

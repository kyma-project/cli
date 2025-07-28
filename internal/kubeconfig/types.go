package kubeconfig

import "k8s.io/apimachinery/pkg/runtime/schema"

var OpenIdConnectGVR = schema.GroupVersionResource{
	Group:    "authentication.gardener.cloud",
	Version:  "v1alpha1",
	Resource: "openidconnects",
}

type OpenIDConnect struct {
	Spec OpenIDConnectSpec `json:"spec"`
}

type OpenIDConnectSpec struct {
	ClientID  string `json:"clientID"`
	IssuerURL string `json:"issuerURL"`
}

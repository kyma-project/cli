package registry

import (
	"github.com/google/go-containerregistry/pkg/authn"
)

type basicAuth struct {
	username, password string
}

func NewBasicAuth(username, password string) authn.Authenticator {
	return &basicAuth{
		username: username,
		password: password,
	}
}

func (ba *basicAuth) Authorization() (*authn.AuthConfig, error) {
	return &authn.AuthConfig{
		Username: ba.username,
		Password: ba.password,
	}, nil
}

package hana

import (
	"encoding/json"
	"github.com/kyma-project/cli.v3/internal/clierror"
	"os"
)

func ReadCredentialsFromFile(path string) (*HanaAdminCredentials, clierror.Error) {
	bytes, err := os.ReadFile(path)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to read credentials file"))
	}
	credentials := &HanaAdminCredentials{}

	err = json.Unmarshal(bytes, credentials)
	if err != nil {
		return nil, clierror.Wrap(err, clierror.New("failed to parse credentials file"))
	}
	return credentials, nil
}

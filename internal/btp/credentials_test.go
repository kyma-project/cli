package btp

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	// nolint:gosec
	correctCredentials = `{
"grant_type": "test",
"uaa": {
	"url": "http://test.com",
	"clientid": "client-id",
	"clientsecret": "client-secret"
	}
}`
)

var (
	correctCredentialsStruct = CISCredentials{
		GrantType: "test",
		UAA: UAA{
			URL:          "http://test.com",
			ClientID:     "client-id",
			ClientSecret: "client-secret",
		},
	}
)

func TestLoadCISCredentials(t *testing.T) {
	tmpDirPath := t.TempDir()

	t.Run("load credentials file", func(t *testing.T) {
		filename := fmt.Sprintf("%s/creds.txt", tmpDirPath)

		err := os.WriteFile(filename, []byte(correctCredentials), os.ModePerm)
		require.NoError(t, err)

		credentials, err := LoadCISCredentials(filename)
		require.NoError(t, err)
		require.Equal(t, correctCredentialsStruct, *credentials)
	})

	t.Run("incorrect credentials file error", func(t *testing.T) {
		filename := fmt.Sprintf("%s/wrong-creds.txt", tmpDirPath)

		err := os.WriteFile(filename, []byte("\n{\n"), os.ModePerm)
		require.NoError(t, err)

		credentials, err := LoadCISCredentials(filename)
		require.Error(t, err)
		require.Nil(t, credentials)
	})

	t.Run("incorrect credentials file error", func(t *testing.T) {
		filename := fmt.Sprintf("%s/doesnotexist-creds.txt", tmpDirPath)

		credentials, err := LoadCISCredentials(filename)
		require.Error(t, err)
		require.Nil(t, credentials)
	})
}

package deploy

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/kyma-project/cli/internal/cli"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

const (
	defaultTLSCrtEnc = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURQVENDQWlXZ0F3SUJBZ0lSQVByWW0wbGhVdUdkeVNCTHo4d3g5VGd3RFFZSktvWklodmNOQVFFTEJRQXcKTURFVk1CTUdBMVVFQ2hNTVkyVnlkQzF0WVc1aFoyVnlNUmN3RlFZRFZRUURFdzVzYjJOaGJDMXJlVzFoTFdSbApkakFlRncweU1EQTNNamt3T1RJek5UTmFGdzB6TURBM01qY3dPVEl6TlROYU1EQXhGVEFUQmdOVkJBb1RER05sCmNuUXRiV0Z1WVdkbGNqRVhNQlVHQTFVRUF4TU9iRzlqWVd3dGEzbHRZUzFrWlhZd2dnRWlNQTBHQ1NxR1NJYjMKRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFDemE4VEV5UjIyTFRKN3A2aXg0M2E3WTVVblovRkNicGNOQkdEbQpxaDRiRGZLcjFvMm1CYldWdUhDbTVBdTBkeHZnbUdyd0tvZzJMY0N1bEd5UXVlK1JLQ0RIVFBJVjdqZEJwZHJhCkNZMXQrNjlJMkJWV0xiblFNVEZmOWw3Vy8yZFFFU0ExZHZQajhMZmlrcEQvUEQ5ekdHR0FQa2hlenVNRU80dUwKaUxXSloyYmpYK1dtaGZXb0lrOG5oak5YNVBFN2l4alMvNnB3QU56eXk2NW95NDJPaHNuYXlDR1grbmhFVk5SRApUejEraEMvdjJaOS9lRG1OdHdjT1hJSk4relZtUTJ4VHh2Sm0rbDUwYzlnenZTY3YzQXg0dUJsOTk3UnVlcUszCmdZMVRmVklFQ0FOTE9hb29jRG5kcW1FY1lBb25SeGJKK0M2U1RJYlhuUVAyMmYxQkFnTUJBQUdqVWpCUU1BNEcKQTFVZER3RUIvd1FFQXdJRm9EQVRCZ05WSFNVRUREQUtCZ2dyQmdFRkJRY0RBVEFNQmdOVkhSTUJBZjhFQWpBQQpNQnNHQTFVZEVRUVVNQktDRUNvdWJHOWpZV3d1YTNsdFlTNWtaWFl3RFFZSktvWklodmNOQVFFTEJRQURnZ0VCCkFBUnVOd0VadW1PK2h0dDBZSWpMN2VmelA3UjllK2U4NzJNVGJjSGtyQVhmT2hvQWF0bkw5cGhaTHhKbVNpa1IKY0tJYkJneDM3RG5ka2dPY3doNURTT2NrdHBsdk9sL2NwMHMwVmFWbjJ6UEk4Szk4L0R0bEU5bVAyMHRLbE90RwpaYWRhdkdrejhXbDFoRzhaNXdteXNJNWlEZHNpajVMUVJ6Rk04YmRGUUJiRGkxbzRvZWhIRTNXbjJjU3NTUFlDCkUxZTdsM00ySTdwQ3daT2lFMDY1THZEeEszWFExVFRMR2oxcy9hYzRNZUxCaXlEN29qb25MQmJNYXRiaVJCOUIKYlBlQS9OUlBaSHR4TDArQ2Nvb1JndmpBNEJMNEtYaFhxZHZzTFpiQWlZc0xTWk0yRHU0ZWZ1Q25SVUh1bW1xNQpVNnNOOUg4WXZxaWI4K3B1c0VpTUttND0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	defaultTLSKeyEnc = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBczJ2RXhNa2R0aTB5ZTZlb3NlTjJ1Mk9WSjJmeFFtNlhEUVJnNXFvZUd3M3lxOWFOCnBnVzFsYmh3cHVRTHRIY2I0SmhxOENxSU5pM0FycFJza0xudmtTZ2d4MHp5RmU0M1FhWGEyZ21OYmZ1dlNOZ1YKVmkyNTBERXhYL1plMXY5blVCRWdOWGJ6NC9DMzRwS1EvencvY3hoaGdENUlYczdqQkR1TGk0aTFpV2RtNDEvbApwb1gxcUNKUEo0WXpWK1R4TzRzWTB2K3FjQURjOHN1dWFNdU5qb2JKMnNnaGwvcDRSRlRVUTA4OWZvUXY3OW1mCmYzZzVqYmNIRGx5Q1RmczFaa05zVThieVp2cGVkSFBZTTcwbkw5d01lTGdaZmZlMGJucWl0NEdOVTMxU0JBZ0QKU3ptcUtIQTUzYXBoSEdBS0owY1d5Zmd1a2t5RzE1MEQ5dG45UVFJREFRQUJBb0lCQUJwVmYvenVFOWxRU3UrUgpUUlpHNzM5VGYybllQTFhtYTI4eXJGSk90N3A2MHBwY0ZGQkEyRVVRWENCeXFqRWpwa2pSdGlobjViUW1CUGphCnVoQ0g2ZHloU2laV2FkWEVNQUlIcU5hRnZtZGRJSDROa1J3aisvak5yNVNKSWFSbXVqQXJRMUgxa3BockZXSkEKNXQwL1o0U3FHRzF0TnN3TGk1QnNlTy9TOGVvbnJ0Q3gzSmxuNXJYdUIzT1hSQnMyVGV6dDNRRlBEMEJDY2c3cgpBbEQrSDN6UjE0ZnBLaFVvb0J4S0VacmFHdmpwVURFeThSSy9FemxaVzBxMDB1b2NhMWR0c0s1V1YxblB2aHZmCjBONGRYaUxuOE5xY1k0c0RTMzdhMWhYV3VJWWpvRndZa0traFc0YS9LeWRKRm5acmlJaDB0ZU81Q0I1ZnpaVnQKWklOYndyMENnWUVBd0gzeksvRTdmeTVpd0tJQnA1M0YrUk9GYmo1a1Y3VUlkY0RIVjFveHhIM2psQzNZUzl0MQo3Wk9UUHJ6eGZ4VlB5TVhnOEQ1clJybkFVQk43cE5xeWxHc3FMOFA1dnZlbVNwOGNKU0REQWN4RFlqeEJLams5CldtOXZnTGpnaERSUFN1Um50QXNxQVVqcWhzNmhHUzQ4WUhMOVI2QlI5dmY2U2xWLzN1NWMvTXNDZ1lFQTdwM1UKRDBxcGNlM1liaiszZmppVWFjcTRGcG9rQmp1MTFVTGNvREUydmZFZUtEQldsM3BJaFNGaHYvbnVqaUg2WWJpWApuYmxKNVRlSnI5RzBGWEtwcHNLWW9vVHFkVDlKcFp2QWZGUzc2blZZaUJvMHR3VzhwMGVCS3QyaUFyejRYRmxUCnpRSnNOS1dsRzBzdGJmSzNqdUNzaWJjYTBUd09lbTdSdjdHV0dLTUNnWUJjZmFoVVd1c2RweW9vS1MvbVhEYisKQVZWQnJaVUZWNlVpLzJoSkhydC9FSVpEY3V2Vk56UW8zWm9Jc1R6UXRXcktxOW56VmVxeDV4cnkzd213SXExZwpCMFlVQVhTRlAvV1ZNWEtTbkhWVzdkRUs2S3pmSHZYTitIRjVSbHdLNmgrWGVyd2hsS093VGxyeVAyTEUrS1JtCks1cHJ5aXJZSWpzUGNKbXFncG9IbFFLQmdCVWVFcTVueFNjNERYZDBYQ0Rua1BycjNlN2lKVjRIMnNmTTZ3bWkKVVYzdUFPVTlvZXcxL2tVSjkwU3VNZGFTV3o1YXY5Qk5uYVNUamJQcHN5NVN2NERxcCtkNksrWEVmQmdUK0swSQpNcmxGT1ZpU09TZ1pjZUM4QzBwbjR2YXJFcS9abC9rRXhkN0M2aUhJUFhVRmpna3ZDUllIQm5DT0NCbjl4TUphClRSWlJBb0dBWS9QYSswMFo1MHYrUU40cVhlRHFrS2FLZU80OFUzOHUvQUJMeGFMNHkrZkJpUStuaXh5ZFUzOCsKYndBR3JtMzUvSU5VRTlkWE44d21XRUlXVUZ3YVR2dHY5NXBpcWNKL25QZkFiY2pDeU8wU3BJWCtUYnFRSkljbgpTVjlrKzhWUFNiRUJ5YXRKVTdIQ3FaNUNTWEZuUnRNanliaWNYYUFKSWtBQm4zVjJ3OFk9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="
)

var (
	crtFile = filepath.Join(".", "testCrt.crt")
	keyFile = filepath.Join(".", "testCrt.key")
)

func TestMain(m *testing.M) {
	setUp()
	retCode := m.Run()
	tearDown()
	os.Exit(retCode)
}

func setUp() {
	deleteCrtFiles()
	crt, err := base64.StdEncoding.DecodeString(defaultTLSCrtEnc)
	if err != nil {
		panic(errors.Wrap(err, "Failed to decode base64-encoded certificate key string"))
	}
	if err := ioutil.WriteFile(crtFile, crt, 0600); err != nil {
		panic(errors.Wrap(err, fmt.Sprintf("Failed to write certificate file '%s'", crtFile)))
	}

	key, err := base64.StdEncoding.DecodeString(defaultTLSKeyEnc)
	if err != nil {
		panic(errors.Wrap(err, "Failed to decode base64-encoded certificate string"))
	}
	if err := ioutil.WriteFile(keyFile, key, 0600); err != nil {
		panic(errors.Wrap(err, fmt.Sprintf("Failed to write certificate key file '%s'", keyFile)))
	}
}

func tearDown() {
	deleteCrtFiles()
}

func deleteCrtFiles() {
	os.Remove(crtFile)
	os.Remove(keyFile)
}

func TestCertAsFile(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: crtFile,
			TLSKeyFile: keyFile,
		}
		tlsCrt, err := opts.tlsCrtEnc()
		require.NoError(t, err)
		require.Equal(t, defaultTLSCrtEnc, tlsCrt)

		tlsKey, err := opts.tlsKeyEnc()
		require.NoError(t, err)
		require.Equal(t, defaultTLSKeyEnc, tlsKey)
	})
	t.Run("Invalid cert file", func(t *testing.T) {
		crtFileInvalid := fmt.Sprintf("%s.xyz", crtFile)
		opts := &Options{
			TLSCrtFile: crtFileInvalid,
			TLSKeyFile: keyFile,
		}
		_, err := opts.tlsCrtEnc()
		require.True(t, os.IsNotExist(err))
	})
	t.Run("Invalid key file", func(t *testing.T) {
		keyFileInvalid := fmt.Sprintf("%s.xyz", keyFile)
		opts := &Options{
			TLSCrtFile: crtFile,
			TLSKeyFile: keyFileInvalid,
		}
		_, err := opts.tlsKeyEnc()
		require.True(t, os.IsNotExist(err))
	})
	t.Run("Ensure proper base64 encoding", func(t *testing.T) {
		ex, err := os.Executable()
		require.NoError(t, err)

		fileContent, err := ioutil.ReadFile(ex)
		require.NoError(t, err)
		expected := base64.StdEncoding.EncodeToString(fileContent)

		require.NoError(t, err)
		opts := &Options{
			TLSCrtFile: ex,
			TLSKeyFile: ex,
		}
		keyEnc, err := opts.tlsKeyEnc()
		require.NoError(t, err)
		require.Equal(t, expected, keyEnc)

		crtEnc, err := opts.tlsCrtEnc()
		require.NoError(t, err)
		require.Equal(t, expected, crtEnc)
	})
}

func TestOptsValidation(t *testing.T) {
	t.Run("Happy path", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: crtFile,
			TLSKeyFile: keyFile,
		}
		err := opts.validateFlags()
		require.NoError(t, err)
	})
	t.Run("Key is missing", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: crtFile,
		}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "key is empty")
	})
	t.Run("Cert is missing", func(t *testing.T) {
		opts := &Options{
			TLSKeyFile: keyFile,
		}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "certificate is empty")
	})
	t.Run("Wrong cert path", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: "/abc/test.yaml",
			TLSKeyFile: keyFile,
		}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
	t.Run("Wrong key path", func(t *testing.T) {
		opts := &Options{
			TLSCrtFile: crtFile,
			TLSKeyFile: "/do/not/exist.key",
		}
		err := opts.validateFlags()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
	t.Run(`Only one of "components-file" and "component" flags can be provided`, func(t *testing.T) {
		opts := &Options{
			TLSCrtFile:     crtFile,
			TLSKeyFile:     keyFile,
			ComponentsFile: "path/to/componentFile",
			Components:     []string{"comp1", "comp2"},
		}
		err := opts.validateFlags()
		require.Error(t, err)
	})
}

func TestComponentFile(t *testing.T) {
	t.Run("Test with non-changed workspace path", func(t *testing.T) {
		opts := NewOptions(cli.NewOptions())
		opts.WorkspacePath = defaultWorkspacePath
		compFile, err := opts.ResolveComponentsFile()
		require.NoError(t, err)
		require.Equal(t, defaultComponentsFile, compFile)
	})
	t.Run("Test with non-existing component file", func(t *testing.T) {
		opts := NewOptions(cli.NewOptions())
		opts.WorkspacePath = defaultWorkspacePath
		opts.ComponentsFile = "/xyz/comp.yaml"
		_, err := opts.ResolveComponentsFile()
		require.Error(t, err)
		require.Contains(t, err.Error(), "not found")
	})
	t.Run("Test with non-existing workspace path", func(t *testing.T) {
		opts := NewOptions(cli.NewOptions())
		opts.WorkspacePath = "/xyz/abc"
		compFile, err := opts.ResolveComponentsFile()
		require.NoError(t, err)
		require.Equal(t, compFile, "/xyz/abc/installation/resources/components.yaml")
	})
	t.Run("Test with changed workspace path and componets file path", func(t *testing.T) {
		_, currFile, _, _ := runtime.Caller(1)
		opts := NewOptions(cli.NewOptions())
		opts.WorkspacePath = filepath.Dir(currFile)
		opts.ComponentsFile = currFile
		compFile, err := opts.ResolveComponentsFile()
		require.NoError(t, err)
		require.Equal(t, currFile, compFile)
	})
}

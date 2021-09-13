package overrides

import (
	"fmt"
	"github.com/kyma-incubator/hydroform/parallel-install/pkg/deployment/k3d"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// interceptor which is replacing a value
type replaceOverrideInterceptor struct {
}

func (roi *replaceOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (roi *replaceOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return "intercepted", nil
}

func (roi *replaceOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

//stringerOverrideInterceptor hides the value of an override when the value is converted to a string
type stringerOverrideInterceptor struct {
}

func (i *stringerOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("string-%v", value)
}

func (i *stringerOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return value, nil
}

func (i *stringerOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

// interceptor which is failing
type failingOverrideInterceptor struct {
}

func (roi *failingOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (roi *failingOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return nil, fmt.Errorf("Interceptor failed")
}

func (roi *failingOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return nil
}

// interceptor which is returning a value for an undefined key
type undefinedOverrideInterceptor struct {
}

func (roi *undefinedOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (roi *undefinedOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return value, nil
}

func (roi *undefinedOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	return fmt.Errorf("This value was missing")
}

func Test_InterceptValue(t *testing.T) {
	t.Run("Test interceptor without failures", func(t *testing.T) {
		builder := Builder{}
		err := builder.AddFile("testdata/deployment-overrides-intercepted.yaml")
		require.NoError(t, err)
		builder.AddInterceptor([]string{"chart.key2.key2-1", "chart.key4"}, &replaceOverrideInterceptor{})

		// read expected result
		data, err := ioutil.ReadFile("testdata/deployment-overrides-intercepted-result.yaml")
		require.NoError(t, err)
		var expected map[string]interface{}
		err = yaml.Unmarshal(data, &expected)
		require.NoError(t, err)

		// verify merge result with expected data
		overrides, err := builder.Build()
		require.NoError(t, err)
		require.Equal(t, expected, overrides.Map())
	})

	t.Run("Test interceptor with failure", func(t *testing.T) {
		builder := Builder{}
		err := builder.AddFile("testdata/deployment-overrides-intercepted.yaml")
		require.NoError(t, err)
		builder.AddInterceptor([]string{"chart.key1"}, &failingOverrideInterceptor{})
		overrides, err := builder.Build()
		require.Empty(t, overrides.Map())
		require.Error(t, err)
	})
}

func Test_InterceptStringer(t *testing.T) {
	builder := Builder{}
	err := builder.AddFile("testdata/deployment-overrides-intercepted.yaml")
	require.NoError(t, err)
	builder.AddInterceptor([]string{"chart.key1", "chart.key3"}, &stringerOverrideInterceptor{})
	overrides, err := builder.Build()
	require.NoError(t, err)
	require.Equal(t,
		"map[chart:map[key1:string-value1yaml key2:map[key2-1:value2.1yaml key2-2:value2.2yaml] key3:string-value3yaml key4:value4yaml]]",
		fmt.Sprint(overrides))
}

func Test_InterceptUndefined(t *testing.T) {
	builder := Builder{}
	err := builder.AddFile("testdata/deployment-overrides-intercepted.yaml")
	require.NoError(t, err)
	builder.AddInterceptor([]string{"I.dont.exist"}, &undefinedOverrideInterceptor{})
	overrides, err := builder.Build()
	require.Empty(t, overrides.Map())
	require.Error(t, err)
	require.Equal(t, "This value was missing", err.Error())
}

func Test_FallbackInterceptor(t *testing.T) {
	builder := Builder{}
	err := builder.AddFile("testdata/deployment-overrides-intercepted.yaml")
	require.NoError(t, err)

	t.Run("Test FallbackInterceptor happy path", func(t *testing.T) {
		builder.AddInterceptor([]string{"I.dont.exist"}, NewFallbackOverrideInterceptor("I am the fallback"))
		overrides, err := builder.Build()
		require.NotEmpty(t, overrides)
		require.NoError(t, err)
		require.Equal(t, "I am the fallback", overrides.Map()["I"].(map[string]interface{})["dont"].(map[string]interface{})["exist"])
	})

	t.Run("Test FallbackInterceptor with sub-key which is not a map", func(t *testing.T) {
		builder.AddInterceptor([]string{"chart.key3.xyz"}, NewFallbackOverrideInterceptor("Use me as fallback"))
		overrides, err := builder.Build()
		require.Empty(t, overrides.Map())
		require.Error(t, err)
	})
}

func Test_GlobalOverridesInterceptionForLocalCluster(t *testing.T) {
	ob := Builder{}
	kubeClient := fake.NewSimpleClientset()

	newDomainNameOverrideInterceptor := NewDomainNameOverrideInterceptor(kubeClient)
	newDomainNameOverrideInterceptor.isLocalCluster = isLocalClusterFunc(true)

	newCertificateOverrideInterceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
	newCertificateOverrideInterceptor.isLocalCluster = isLocalClusterFunc(true)

	ob.AddInterceptor([]string{"global.isLocalEnv", "global.environment.gardener"}, NewFallbackOverrideInterceptor(false))
	ob.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, newDomainNameOverrideInterceptor)
	ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, newCertificateOverrideInterceptor)

	// read expected result
	data, err := ioutil.ReadFile("testdata/deployment-global-overrides.yaml")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = yaml.Unmarshal(data, &expected)
	require.NoError(t, err)

	// verify global overrides
	overrides, err := ob.Build()
	require.NotEmpty(t, overrides)
	require.NoError(t, err)
	require.Equal(t, expected, overrides.Map())
}

func Test_GlobalOverridesInterceptionForNonGardenerCluster(t *testing.T) {
	ob := Builder{}
	kubeClient := fake.NewSimpleClientset()

	newDomainNameOverrideInterceptor := NewDomainNameOverrideInterceptor(kubeClient)
	newDomainNameOverrideInterceptor.isLocalCluster = isLocalClusterFunc(false)

	newCertificateOverrideInterceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
	newCertificateOverrideInterceptor.isLocalCluster = isLocalClusterFunc(false)

	ob.AddInterceptor([]string{"global.isLocalEnv", "global.environment.gardener"}, NewFallbackOverrideInterceptor(false))
	ob.AddInterceptor([]string{"global.domainName", "global.ingress.domainName"}, newDomainNameOverrideInterceptor)
	ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, newCertificateOverrideInterceptor)

	// read expected result
	data, err := ioutil.ReadFile("testdata/deployment-global-overrides-for-remote-cluster.yaml")
	require.NoError(t, err)
	var expected map[string]interface{}
	err = yaml.Unmarshal(data, &expected)
	require.NoError(t, err)

	expectedKeys := extractKeys(expected)
	require.Equal(t, 6, len(expectedKeys))
	require.Contains(t, expectedKeys, "global.isLocalEnv")
	require.Contains(t, expectedKeys, "global.domainName")
	require.Contains(t, expectedKeys, "global.tlsCrt")
	require.Contains(t, expectedKeys, "global.tlsKey")
	require.Contains(t, expectedKeys, "global.environment.gardener")
	require.Contains(t, expectedKeys, "global.ingress.domainName")

	// verify global overrides
	overrides, err := ob.Build()
	require.NotEmpty(t, overrides)
	require.NoError(t, err)
	require.Equal(t, expected, overrides.Map())
}

func Test_DomainNameOverrideInterceptor(t *testing.T) {
	ob := Builder{}

	gardenerCM := fakeGardenerCM()

	mockNewDomainNameOverrideInterceptor := func(kubeClient kubernetes.Interface, isLocal bool) *DomainNameOverrideInterceptor {
		newDomainNameOverrideInterceptor := NewDomainNameOverrideInterceptor(kubeClient)
		newDomainNameOverrideInterceptor.isLocalCluster = isLocalClusterFunc(isLocal)
		return newDomainNameOverrideInterceptor
	}

	t.Run("test default domain for local cluster", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset()
		ob.AddInterceptor([]string{"global.domainName"}, mockNewDomainNameOverrideInterceptor(kubeClient, true))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Equal(t, 1, len(extractKeys(overrides.Map())))
		require.Equal(t, localKymaDevDomain, getOverride(overrides.Map(), "global.domainName"))
	})

	t.Run("test default domain for remote non-gardener cluster", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset()
		ob.AddInterceptor([]string{"global.domainName"}, mockNewDomainNameOverrideInterceptor(kubeClient, false))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Equal(t, 1, len(extractKeys(overrides.Map())))
		require.Equal(t, defaultRemoteKymaDomain, getOverride(overrides.Map(), "global.domainName"))
	})

	t.Run("test valid domain for a gardener cluster", func(t *testing.T) {
		//givenOverrides
		kubeClient := fake.NewSimpleClientset(gardenerCM)
		ob.AddInterceptor([]string{"global.domainName"}, NewDomainNameOverrideInterceptor(kubeClient))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Equal(t, 1, len(extractKeys(overrides.Map())))
		require.Equal(t, "gardener.domain", getOverride(overrides.Map(), "global.domainName"))
	})

	t.Run("test user-provided domain is overridden on gardener cluster", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset(gardenerCM)

		ob := Builder{}
		domainNameOverrides := make(map[string]interface{})
		domainNameOverrides["domainName"] = "user.domain"
		err := ob.AddOverrides(map[string]interface{}{
			"global": domainNameOverrides,
		})
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.domainName"}, NewDomainNameOverrideInterceptor(kubeClient))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Equal(t, 1, len(extractKeys(overrides.Map())))
		require.Equal(t, "gardener.domain", getOverride(overrides.Map(), "global.domainName"))
	})

	t.Run("test user-provided domain is not overridden on local cluster", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset()

		ob := Builder{}
		domainNameOverrides := make(map[string]interface{})
		domainNameOverrides["domainName"] = "user.domain"
		err := ob.AddOverrides(map[string]interface{}{
			"global": domainNameOverrides,
		})
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.domainName"}, mockNewDomainNameOverrideInterceptor(kubeClient, true))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Equal(t, 1, len(extractKeys(overrides.Map())))
		require.Equal(t, "user.domain", getOverride(overrides.Map(), "global.domainName"))
	})

	t.Run("test user-provided domain is not overridden on remote cluster", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset()

		ob := Builder{}
		domainNameOverrides := make(map[string]interface{})
		domainNameOverrides["domainName"] = "user.domain"
		err := ob.AddOverrides(map[string]interface{}{
			"global": domainNameOverrides,
		})
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.domainName"}, mockNewDomainNameOverrideInterceptor(kubeClient, false))

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides)
		require.Equal(t, 1, len(extractKeys(overrides.Map())))
		require.Equal(t, "user.domain", getOverride(overrides.Map(), "global.domainName"))
	})
}

func Test_CertificateOverridesInterception(t *testing.T) {

	testFakeCrt := "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUUrakNDQXVJQ0NRQ09EVk1VNHBqUFBUQU5CZ2txaGtpRzl3MEJBUXNGQURBL01Rc3dDUVlEVlFRR0V3SlEKVERFTk1Bc0dBMVVFQ2d3RVMzbHRZVEVOTUFzR0ExVUVDd3dFUzNsdFlURVNNQkFHQTFVRUF3d0phM2x0WVM1MApaWE4wTUI0WERUSXhNRFF5TVRFMU1UZ3hNRm9YRFRNeE1EUXhPVEUxTVRneE1Gb3dQekVMTUFrR0ExVUVCaE1DClVFd3hEVEFMQmdOVkJBb01CRXQ1YldFeERUQUxCZ05WQkFzTUJFdDViV0V4RWpBUUJnTlZCQU1NQ1d0NWJXRXUKZEdWemREQ0NBaUl3RFFZSktvWklodmNOQVFFQkJRQURnZ0lQQURDQ0Fnb0NnZ0lCQU40SUl4QnovN3dnd01SQQoxbVRlbWxhZzllaitSdDZvSUo5UTlUTWtDU1dDTk1tRFMzUW5UZDcyVVQxY09YandRVTcvMWEvbGVjY1BNdHpnCmlmckVmUVNMSjd0M3F2U01iN0ZJMmlYdFhqRjliV1oycGxXMGNlMFkxalIydmNRTEMvTDltcW5SSFZGaHRvRWMKYUhrU3Zmb2xkemxkbmdGNlN4UnFLZGQ3eVYvbHVXTkF3SWZ3KzRwMGNuZXQ0emExQUY4VkxXTGRXZThUM1c1eQprVmFHOXo1QUkwMURjSTlYUVJINHZIUktzcnJXUk5iNnUrcTVKRFFra1pENG5YN3NVNTdtcmQzWE1rZ2tpMjR5CmNPdUk5U2NxMHpIU2duWmFKUDVVbjFmaHg0MytweXgvSnB3SFgvREU0THN0cWc0SWxENTFoU1RsZGRnekZVYnUKL0tMdUZReTV5SnA5TUtFakx0L1RMWVhxK1BZQmVsdDFSZXB1V21ES2Y1aWJ4S0ZFWG9yc0V3QnN2aGR4VDVjcgpDelZvUmE2Z0dkbjZDaThDUVlUbmhYREVhemlvdWtaU2gxKytpT2NaVHB0eC95TTU5S1dYSHZEQzFwR2VWQk80Cm9pdFRRYkRUTjJTS3c0K1BhYTluQzBjVXJGa2hTUzJXYUNvSmNvWmdoZWJJQkNQb3FzNWpTeUVwNUsxM1dwbXIKVlZQVVhNQi8weDBjWUJ5R1cwNnpiNlJOUzdxR3g4YTBlQjU3MXI0YUljaThFVnlKa1BUbk16MG9JcVVQZHcxWApmcTdDMm1acm5yN1lQSEhnMEVCY0lFbHllL1Z0ZlcyU3M0cDZheUJrcTFpbDJNTE5ISVEzUThuU3lWL290MnpOCjFxUURENUtRbFJpcmxYcjBBV3psUklPZG5rV3JBZ01CQUFFd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dJQkFCL2cKNWhiOENqZVlrZkdlek9BMjFpYVpuZTVTV3c3eXNacHFiUGk5NHB2eFJIcklTbFFyM3lGWTBTS1JYcFR6Sk54RApEbWpNMjlJaDBnUXA2UkhLYUtFc094VE5NMWVBNW5MckV3THNPZDhDNDJibmFiSlFtbFlubnAwOXorMWJTTnJBCmZuNWNDek52MVJ5aVNDNHZ4SUpBdUIzOVpFODgvMHhMQXRDd01GS1hwNkJ2b2FxTTJPblRqcE5XZCt1Z2E5eGcKUkQvVVY4UXNTNTJSL083azlWaGI1YXZjUkorOG8vdTU3NVBHdzNHRDVHckwwQWplVXljaEFMVlB6MHVURXl4aQpQK0oyR0ZYRDJGZFl4aVVwR3V5eld5ajlWTEtyV0ljbEIzYytMTDZNMW9UN0gvNUdsQXI2d0htVEhobXBCekJLCnEwMC9EV2NzSHNvUzVKNG8wTUtJbEkxUnV2REhqK1Y2YU5NUWVXR08vQm43V1hiZEhld1BJTFBUVHFqOEJUU2YKUGhXcEdXRmRSRWNsUldmVU93R05MYkY1U01ycVNLOG1lNFdyZmMvTW85WVpHbFg2cDlmMXZMTWlhc3lDOGFXNgpUSHZyVkttcEtjTUpBUVJlUTQrb3hGc3pMSStQZTRVcHdrcEdnMmRPSm9zb09CcUphTkRJV0ttazZVV29sLzJoCnF4QXh6cy8zTW9OUzVOSWNMZkloQXhoeXlzbXNGZ1E1MkQvaWtFeGN6UzBzVC9aQWtuMTZjTzlTYzhKWVRKY2oKVEdLWlRyVWFmby9RS0Y3MjZyOWtsNmN4NmxramJiaVFCQzFxczgyb1FVREpNQnJWdFlha0huL1FXNGtDOWdBSQpkWGtEeTNMeWlhR2VBMm9BeFFPZytHTnhNV2pCWEU1aytGRndsSUhVCi0tLS0tRU5EIENFUlRJRklDQVRFLS0tLS0K"
	testFakeKey := "LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0tCk1JSUpSQUlCQURBTkJna3Foa2lHOXcwQkFRRUZBQVNDQ1M0d2dna3FBZ0VBQW9JQ0FRRGVDQ01RYy8rOElNREUKUU5aazNwcFdvUFhvL2tiZXFDQ2ZVUFV6SkFrbGdqVEpnMHQwSjAzZTlsRTlYRGw0OEVGTy85V3Y1WG5IRHpMYwo0SW42eEgwRWl5ZTdkNnIwakcreFNOb2w3VjR4ZlcxbWRxWlZ0SEh0R05ZMGRyM0VDd3Z5L1pxcDBSMVJZYmFCCkhHaDVFcjM2SlhjNVhaNEJla3NVYWluWGU4bGY1YmxqUU1DSDhQdUtkSEozcmVNMnRRQmZGUzFpM1ZudkU5MXUKY3BGV2h2YytRQ05OUTNDUFYwRVIrTHgwU3JLNjFrVFcrcnZxdVNRMEpKR1ErSjErN0ZPZTVxM2QxekpJSkl0dQpNbkRyaVBVbkt0TXgwb0oyV2lUK1ZKOVg0Y2VOL3Fjc2Z5YWNCMS93eE9DN0xhb09DSlErZFlVazVYWFlNeFZHCjd2eWk3aFVNdWNpYWZUQ2hJeTdmMHkyRjZ2ajJBWHBiZFVYcWJscGd5bitZbThTaFJGNks3Qk1BYkw0WGNVK1gKS3dzMWFFV3VvQm5aK2dvdkFrR0U1NFZ3eEdzNHFMcEdVb2Rmdm9qbkdVNmJjZjhqT2ZTbGx4N3d3dGFSbmxRVAp1S0lyVTBHdzB6ZGtpc09QajJtdlp3dEhGS3haSVVrdGxtZ3FDWEtHWUlYbXlBUWo2S3JPWTBzaEtlU3RkMXFaCnExVlQxRnpBZjlNZEhHQWNobHRPczIra1RVdTZoc2ZHdEhnZWU5YStHaUhJdkJGY2laRDA1ek05S0NLbEQzY04KVjM2dXd0cG1hNTYrMkR4eDROQkFYQ0JKY252MWJYMXRrck9LZW1zZ1pLdFlwZGpDelJ5RU4wUEowc2xmNkxkcwp6ZGFrQXcrU2tKVVlxNVY2OUFGczVVU0RuWjVGcXdJREFRQUJBb0lDQVFDWWZmZ3ZMYXcvdGpNTzF3VW9wQ1pXClJ4aDkzRjRxUUVpZmd3ZlZCdlB0T2Y4dFE2cUg3Ukt6aG5NSGRKYllkQkkyd3NrdGxLck54NmVFUWdjaUh0OUsKUnBTVVViMHRWbUxEM1NoT2tqZDJRQkhxSktWYkNhS1JWOVNPbGRzQmtTQzAwKzdzb1AzRXpocDlsRmhBaDFuSgpPd0FtZXlDeEhTQUJ0bVJrWmRWSnNzcGYyN0lmNjZlblVSRHBGNW1ORWtWZUNIcHlnMXBvTkRtSnlNLy9JSlVnCndRWTk0NHFrT0NZdHhLc1NKOWVYTU9CNDBoNU1PTG9md2Rua09veFpCdERydXIxQk5yS0hEK3BmVmU5dUpWTlMKZ2p3bzVNN0xvRi8xK1lLeTVoT2JkNEd6c3VSK2x5WVNnL1ZoT1J5cHBNVEVIUXpENllERmExZzZycHIvQUF1ZAp1Um5xVEx0eEdzaXpTeE1MNVdYaGY5YVFZeFduR0tWVDlZeUE4d3lGWnpzTUtvMkJpRlhKT2lXdGNXSVl4NEY0CjJRcmhmK1F5cmRpRmgxUFA5SzkrUi9xTzhSVWgxSlgvYkYzMEhuRklhSnplOEdWdzRqdEdxTWxaK2NjV21kaGwKUnM5akpodEdtQTkzWDlSc2dpRnllMWVPMmh2T3haT1lxUzNBVW1MS2dqb0JXSGtNV0ZWNjZnLzVBQ3lFd0xVYwpUeGFneVQ1dmtSMm5MeU5JdGVncldFdXpVZDBNbU5rQ3pwYWZGc0lxYVFNOU0xRmpXZmlScVFjOENlcmhrMnRCCmluMUpYdHpObDRSZndUYVloclhMQmdwWjhkbjlkOXpxTGR3MFhEV09jbmxkVWRmL3JWUVhnbU80Y09iVUF5N1oKQ2d2c09ObjMzY1RXa2JQd1E2ZE9ZUUtDQVFFQStJbFBkTjVveGFOVHYvam1rL2QxWnB1N0xuN24rQitqL1pDcgpsTE9IaE1JZHl2cXg5UU5ZTFhsa3ZGd3NrZTdnL3dlSng5ckU4MGFlU0dCd1Rod2gxYzlXMUk1bGlDS05UcEI2Ck55cnlOTXUxdi9PSnFDR3A3MDRXNlZBdEdDYXNSSDRleEp6QlNibElCZ2ZsdXUxRm9LKzdyeWQwR0FwME9ocTEKTmJma1NReDh0YWZqSk1ydmx6MXAwbUQ1eG92NlY2eEt6cU13ZndtQ1ZRYXlvSzdJTTJRL2drbVNrUDRERkM3RAppYW1SYWMyaVpEbmV3REhjQzFKbE9jY05URVQ2TkNqa2xTa1BzMm1IZGlCeXJxVXBZNEFEZGo3T2ZwV3k4N1QxCmplRFFtR0xIc3k0MjlvL3BMazY2bmNCUzIvSWF1NFJhYWp0RGltWFJpaGx5SzV3QU93S0NBUUVBNUxNUXh3ZTQKZHJKQzg3NytaZS9UZ2UzNXNRODNrZXRoMEFmRmtJbGlCdERpc1RDSTV1YTFVdGk0Q1h1bHBOUDExN0lnL3l6KwpiWlJBTWVjZXhnL3JGWmZhWFhUcmU4VnR1WDJJdS8xYnd5WXlzNE13RFJ0Z2xwS2hIWUdHN0xLY21EVHB6MS9SCnBoQTY0a3dBaEl1SnpENHVwb3JYbmRJcUJaT20wdnlmeUMzQ25MbjdYeVB0Tlp2d0tlbkFMV3FCemFpQXUvbzkKYjdIN3pFdWhyRDU3UXRDdDJGV1ZPbmZYSGJjRTN0dytlS0hUMGNWakRQZ3Z6d0xjZC9rQmRmYjJxenI1OVNVRwpCdFF1cS93aDFTSUc3eHZRYzJzUVVaeEJxK3E0aEhzVnNLbXk1YStGR2E5VkdkYVVXUUhYcmR2Q1dFVkJHOTF5CkduTnZJM0J4eHQ1cFVRS0NBUUVBaTh1NFhMVkpXM205L3VwQzBCSE9BSFF5T2puNXdyQVJidXYwQndWZ2djVXEKT3VUK09pR3lkSW1tcHVoMXpYUC9MSlFSNU05aUhyQ25FWERsV3BvcVVmaDVEOEEwemZrWllJcVZvL2hOR25ORwovUHhBZnNqSXJDbFJhOVRFT0tSd0cycVJaZWdDTkxTNkZXSlZ6dW50VXkvbHN1VFBRVUtJRTdLNElNb1o1eGpXCkFOdTVRUlhBNUdJUDV0elRRZUcwWTZJdXhjSTI0ZzM0T0ZrM0duaVZkWXE2eWs4VjJPWjMxdDlpNzBqbzJRbG8KZ1ZXbnZKV08vdk5PcXN3UzU4YVlzY1FhcHVmY3cvN2t5Z1lBVzhuYzJQSEZnTHBkTGdpSUN0ckxrQTFYWjQrUQpZbkhwU3BDeUNYRVJPUEJYNncxb0NmZXRYN21NQ1FteWJpcFg5TDJmeHdLQ0FRQjIwbHBMTGtXMjFkTlhWTXBVCktCQ2FGd3g5NDh3WmNsUTFnM1F3TGxEUi9jRnFFaTl1MkRzcE9oUVVTVHU4c2F1dlQ0czVTU1UveGFDOHpMbisKYWRMWU96ZG5DeEkyRWxONTVqRWVpdm9jSUVLRFpndVhJN3hCUHhtYWZPdWZHd1dsUndpYmg4c2pIcGVaYjZkdApOaHA0Rlc2amRNdWw0Y1dYZENsZXdZWTZ1UnU5MWhzMlNUSTdnak43YzBrM3ozaDFZN0RPK2FybDEzRmRxWVhzCk9lSk15cU1vSFA4Vmk2SW1mQ3A1cDdDRmVIN1hKRmpjS2k2Y3ZYM1NqM3NrMFJWRHpiYUVtYUhSOW5meFAyUk0KbWd0RVBBMUhpajdHU0FzT3lUcnBDaEl3NFZwalg1Z2x5aVRLOGVQTmd0bU9LUG1HWnlUMjEwMHJWUUpQUldLMwowbUtoQW9JQkFRREFDbGlVTTlhUlNWS2FkMnQweHExT2d4VXd4YmcyK1UyZWVYcnc3Nm5sbmc3V3lKbm9QVTNKClhJMjhIN1pHTmQzM1FJTXNjMWJRcGdseFRHbU1sTjNuV1RBSDE0bE44Y0lSTXhCaFh1WU16ejM0RkFqNmRSdzgKRWtjV1daM3pGZUYrK09zbngxc3dxeldMbHJZV1A1bVg3TEx5VzBNOEptN2hyUTZHMzJIRGg5cnlzay96Yk5IZApadHFtYXdySVpBU1FVaGREUUFPS1hIL2hLdDBGY2tFR0NabStac2FNdHY3QVlCelJ6QWZjL0FDdXp4VVhRV2J2CkdDT0xZYkR6N3RxQ3JqdGFwVWVlNTFSd3pjSTNaeEhORVlRTjNsQ3NUQWdMQ1FUOU5iWFpUUFBRVy9SYkIrY2MKZlFtRFV6MHNjVXI2Z0FVWVdoSHJXQWVlZ0xYRFlhSmEKLS0tLS1FTkQgUFJJVkFURSBLRVktLS0tLQo="

	gardenerCM := fakeGardenerCM()

	t.Run("test default cert for local cluster", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset()
		interceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
		interceptor.isLocalCluster = isLocalClusterFunc(true)

		ob := Builder{}

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, interceptor)

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
		require.Equal(t, 2, len(extractKeys(overrides.Map())))
		require.Equal(t, getOverride(overrides.Map(), "global.tlsCrt"), defaultLocalTLSCrtEnc)
		require.Equal(t, getOverride(overrides.Map(), "global.tlsKey"), defaultLocalTLSKeyEnc)
	})

	t.Run("test default cert for remote non-gardener cluster", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset()
		interceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
		interceptor.isLocalCluster = isLocalClusterFunc(false)

		ob := Builder{}

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, interceptor)

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
		require.Equal(t, 2, len(extractKeys(overrides.Map())))
		require.Equal(t, getOverride(overrides.Map(), "global.tlsCrt"), defaultRemoteTLSCrtEnc)
		require.Equal(t, getOverride(overrides.Map(), "global.tlsKey"), defaultRemoteTLSKeyEnc)
	})

	t.Run("test default cert is not set for a gardener cluster even if isLocalCluster returns true", func(t *testing.T) {
		kubeClient := fake.NewSimpleClientset(gardenerCM)

		// givenOverrides
		interceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
		interceptor.isLocalCluster = isLocalClusterFunc(true)

		ob := Builder{}

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, interceptor)

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.Empty(t, overrides.Map())
	})

	t.Run("test user-provided cert is reset to an empty string for a gardener cluster", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset(gardenerCM)
		interceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
		interceptor.isLocalCluster = isLocalClusterFunc(true)

		ob := Builder{}

		tlsOverrides := make(map[string]interface{})
		tlsOverrides["tlsCrt"] = defaultLocalTLSCrtEnc
		tlsOverrides["tlsKey"] = defaultLocalTLSKeyEnc
		err := ob.AddOverrides(map[string]interface{}{
			"global": tlsOverrides,
		})
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, interceptor)

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		// then user-provided overrides are replaced by empty strings
		require.Equal(t, 2, len(extractKeys(overrides.Map())))
		require.Equal(t, getOverride(overrides.Map(), "global.tlsCrt"), "")
		require.Equal(t, getOverride(overrides.Map(), "global.tlsKey"), "")
	})

	t.Run("test user-provided cert is preserved for a local cluster", func(t *testing.T) {

		// givenOverrides
		kubeClient := fake.NewSimpleClientset()
		interceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
		interceptor.isLocalCluster = isLocalClusterFunc(true)

		ob := Builder{}

		tlsOverrides := make(map[string]interface{})
		tlsOverrides["tlsCrt"] = testFakeCrt
		tlsOverrides["tlsKey"] = testFakeKey
		err := ob.AddOverrides(map[string]interface{}{
			"global": tlsOverrides,
		})
		require.NoError(t, err)

		// Ensure user provides values different than defaults for local domain
		require.NotEqual(t, tlsOverrides["tlsCrt"], defaultLocalTLSCrtEnc)
		require.NotEqual(t, tlsOverrides["tlsKey"], defaultLocalTLSKeyEnc)

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, interceptor)

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.Equal(t, 2, len(extractKeys(overrides.Map())))
		require.Equal(t, getOverride(overrides.Map(), "global.tlsCrt"), testFakeCrt)
		require.Equal(t, getOverride(overrides.Map(), "global.tlsKey"), testFakeKey)
	})

	t.Run("test user-provided cert is preserved for a remote non-gardener cluster", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset()
		interceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
		interceptor.isLocalCluster = isLocalClusterFunc(false)

		tlsOverrides := make(map[string]interface{})
		tlsOverrides["tlsCrt"] = testFakeCrt
		tlsOverrides["tlsKey"] = testFakeKey
		ob := Builder{}
		err := ob.AddOverrides(map[string]interface{}{
			"global": tlsOverrides,
		})
		require.NoError(t, err)

		// Ensure user provides values different than defaults for remote domain
		require.NotEqual(t, tlsOverrides["tlsCrt"], defaultRemoteTLSCrtEnc)
		require.NotEqual(t, tlsOverrides["tlsKey"], defaultRemoteTLSKeyEnc)

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, interceptor)

		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.Equal(t, 2, len(extractKeys(overrides.Map())))
		require.Equal(t, getOverride(overrides.Map(), "global.tlsCrt"), testFakeCrt)
		require.Equal(t, getOverride(overrides.Map(), "global.tlsKey"), testFakeKey)
	})

	t.Run("test invalid crt key pair", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset()
		interceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
		interceptor.isLocalCluster = isLocalClusterFunc(true)

		tlsOverrides := make(map[string]interface{})
		tlsOverrides["tlsCrt"] = testFakeCrt
		tlsOverrides["tlsKey"] = defaultLocalTLSKeyEnc
		ob := Builder{}
		err := ob.AddOverrides(map[string]interface{}{
			"global": tlsOverrides,
		})
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, interceptor)
		// when
		overrides, err := ob.Build()

		// then
		require.Error(t, err)
		require.Contains(t, err.Error(), "private key does not match public key")
		require.Empty(t, overrides.Map())
	})

	t.Run("test invalid key format", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset()
		interceptor := NewCertificateOverrideInterceptor("global.tlsCrt", "global.tlsKey", kubeClient)
		interceptor.isLocalCluster = isLocalClusterFunc(true)

		tlsOverrides := make(map[string]interface{})
		tlsOverrides["tlsCrt"] = testFakeCrt
		tlsOverrides["tlsKey"] = "V2VkIEFwciAyMSAxNzoyNTowOCBDRVNUIDIwMjEK"
		ob := Builder{}
		err := ob.AddOverrides(map[string]interface{}{
			"global": tlsOverrides,
		})
		require.NoError(t, err)

		ob.AddInterceptor([]string{"global.tlsCrt", "global.tlsKey"}, interceptor)
		// when
		overrides, err := ob.Build()

		// then
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to find any PEM data in key")
		require.Empty(t, overrides.Map())
	})
}

func Test_RegistryEnableOverrideInterception(t *testing.T) {
	k3dNode := k3d.FakeK3dNode()
	generalNode := fakeNode()

	t.Run("test disable internal registry for k3d cluster", func(t *testing.T) {
		// givenOverrides
		dockerRegistryOverrides := make(map[string]interface{})
		dockerRegistryOverrides["enableInternal"] = "true"
		serverlessOverrides := make(map[string]interface{})
		serverlessOverrides["dockerRegistry"] = dockerRegistryOverrides
		ob := Builder{}
		err := ob.AddOverrides(map[string]interface{}{
			"serverless": serverlessOverrides,
		})
		require.NoError(t, err)

		kubeClient := fake.NewSimpleClientset(k3dNode)
		ob.AddInterceptor([]string{"serverless.dockerRegistry.enableInternal"}, NewRegistryDisableInterceptor(kubeClient))
		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
		require.Equal(t, "false", getOverride(overrides.Map(), "serverless.dockerRegistry.enableInternal"))
	})

	t.Run("test preserve internal registry for non k3d cluster", func(t *testing.T) {
		// givenOverrides
		dockerRegistryOverrides := make(map[string]interface{})
		dockerRegistryOverrides["enableInternal"] = "true"
		serverlessOverrides := make(map[string]interface{})
		serverlessOverrides["dockerRegistry"] = dockerRegistryOverrides
		ob := Builder{}
		err := ob.AddOverrides(map[string]interface{}{
			"serverless": serverlessOverrides,
		})
		require.NoError(t, err)

		kubeClient := fake.NewSimpleClientset(generalNode)
		ob.AddInterceptor([]string{"serverless.dockerRegistry.enableInternal"}, NewRegistryDisableInterceptor(kubeClient))
		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
		require.Equal(t, "true", getOverride(overrides.Map(), "serverless.dockerRegistry.enableInternal"))
	})

}

func Test_RegistryOverridesInterception(t *testing.T) {
	k3dNode := k3d.FakeK3dNode()
	generalNode := fakeNode()

	t.Run("test getting k3d cluster name", func(t *testing.T) {
		// givenOverrides
		kubeClient := fake.NewSimpleClientset(k3dNode)

		// when
		clusterName, err := k3d.K3dClusterName(kubeClient)

		// then
		require.NoError(t, err)
		require.Equal(t, "kyma", clusterName)
	})

	t.Run("test registry address for k3d cluster if undefined", func(t *testing.T) {
		// givenOverrides
		dockerRegistryOverrides := make(map[string]interface{})
		dockerRegistryOverrides["enableInternal"] = "true"
		serverlessOverrides := make(map[string]interface{})
		serverlessOverrides["dockerRegistry"] = dockerRegistryOverrides
		ob := Builder{}
		err := ob.AddOverrides(map[string]interface{}{
			"serverless": serverlessOverrides,
		})
		require.NoError(t, err)

		kubeClient := fake.NewSimpleClientset(k3dNode)
		ob.AddInterceptor([]string{"serverless.dockerRegistry.internalServerAddress", "serverless.dockerRegistry.serverAddress", "serverless.dockerRegistry.registryAddress"}, NewRegistryInterceptor(kubeClient))
		// when
		overrides, err := ob.Build()
		require.NoError(t, err)

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
		require.Equal(t, "k3d-kyma-registry:5000", getOverride(overrides.Map(), "serverless.dockerRegistry.serverAddress"))
		require.Equal(t, "k3d-kyma-registry:5000", getOverride(overrides.Map(), "serverless.dockerRegistry.internalServerAddress"))
		require.Equal(t, "k3d-kyma-registry:5000", getOverride(overrides.Map(), "serverless.dockerRegistry.registryAddress"))
	})

	t.Run("test registry address for k3d cluster if overridden", func(t *testing.T) {
		// givenOverrides
		dockerRegistryOverrides := make(map[string]interface{})
		dockerRegistryOverrides["enableInternal"] = "true"
		dockerRegistryOverrides["serverAddress"] = "serverAddress"
		dockerRegistryOverrides["internalServerAddress"] = "internalServerAddress"
		dockerRegistryOverrides["registryAddress"] = "registryAddress"
		serverlessOverrides := make(map[string]interface{})
		serverlessOverrides["dockerRegistry"] = dockerRegistryOverrides
		ob := Builder{}
		err := ob.AddOverrides(map[string]interface{}{
			"serverless": serverlessOverrides,
		})
		require.NoError(t, err)

		kubeClient := fake.NewSimpleClientset(k3dNode)
		ob.AddInterceptor([]string{"serverless.dockerRegistry.internalServerAddress", "serverless.dockerRegistry.serverAddress", "serverless.dockerRegistry.registryAddress"}, NewRegistryInterceptor(kubeClient))
		// when
		overrides, err := ob.Build()
		require.NoError(t, err)

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
		require.Equal(t, "serverAddress", getOverride(overrides.Map(), "serverless.dockerRegistry.serverAddress"))
		require.Equal(t, "internalServerAddress", getOverride(overrides.Map(), "serverless.dockerRegistry.internalServerAddress"))
		require.Equal(t, "registryAddress", getOverride(overrides.Map(), "serverless.dockerRegistry.registryAddress"))
	})

	t.Run("test preserve registry address for non k3d cluster", func(t *testing.T) {
		// givenOverrides
		dockerRegistryOverrides := make(map[string]interface{})
		dockerRegistryOverrides["enableInternal"] = "false"
		dockerRegistryOverrides["serverAddress"] = "serverAddress"
		dockerRegistryOverrides["internalServerAddress"] = "internalServerAddress"
		dockerRegistryOverrides["registryAddress"] = "registryAddress"
		serverlessOverrides := make(map[string]interface{})
		serverlessOverrides["dockerRegistry"] = dockerRegistryOverrides
		ob := Builder{}
		err := ob.AddOverrides(map[string]interface{}{
			"serverless": serverlessOverrides,
		})
		require.NoError(t, err)

		kubeClient := fake.NewSimpleClientset(generalNode)
		ob.AddInterceptor([]string{"serverless.dockerRegistry.internalServerAddress", "serverless.dockerRegistry.serverAddress", "serverless.dockerRegistry.registryAddress"}, NewRegistryInterceptor(kubeClient))
		// when
		overrides, err := ob.Build()

		// then
		require.NoError(t, err)
		require.NotEmpty(t, overrides.Map())
		require.Equal(t, getOverride(overrides.Map(), "serverless.dockerRegistry.serverAddress"), "serverAddress")
		require.Equal(t, getOverride(overrides.Map(), "serverless.dockerRegistry.internalServerAddress"), "internalServerAddress")
		require.Equal(t, getOverride(overrides.Map(), "serverless.dockerRegistry.registryAddress"), "registryAddress")
	})
}

func fakeNode() *v1.Node {
	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "my-node-0",
		},
	}

	return node
}

func fakeGardenerCM() *v1.ConfigMap {
	domainData := make(map[string]string)
	domainData["domain"] = "gardener.domain"

	gardenerCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "shoot-info",
			Namespace: "kube-system",
		},
		Data: domainData,
	}

	return gardenerCM
}

func isLocalClusterFunc(val bool) func() (bool, error) {
	return func() (bool, error) {
		return val, nil
	}
}

func getOverride(overrides map[string]interface{}, key string) string {
	keys := strings.Split(key, ".")
	if len(keys) == 0 {
		panic("no key")
	}

	res := ""

	if len(keys) == 1 {
		res = overrides[keys[0]].(string)
	} else {
		val := overrides
		var i int
		for i = 0; i < len(keys)-1; i++ {
			val = val[keys[i]].(map[string]interface{})
		}

		res = val[keys[i]].(string)
	}

	return res
}

func extractKeys(overrides map[string]interface{}) []string {
	keys := []string{}

	for k := range overrides {
		if subMap, isMap := overrides[k].(map[string]interface{}); isMap {
			subKeys := extractKeys(subMap)
			for si := 0; si < len(subKeys); si++ {
				keys = append(keys, k+"."+subKeys[si])
			}
		} else {
			keys = append(keys, k)
		}
	}

	return keys
}

package values

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/kyma-project/cli/internal/clusterinfo"
	"github.com/stretchr/testify/require"
)

func TestMerge(t *testing.T) {
	testCases := []struct {
		summary     string
		values      []string
		valueFiles  []string
		domain      string
		tlsCrt      string
		tlsKey      string
		expected    Values
		expectedErr bool
	}{
		{
			summary: "single value",
			values:  []string{"component.key=foo"},
			expected: map[string]interface{}{
				"component": map[string]interface{}{
					"key": "foo",
				},
			},
		},
		{
			summary: "single value comma separated",
			values:  []string{"component.key=foo,component.inner.key=bar"},
			expected: map[string]interface{}{
				"component": map[string]interface{}{
					"key": "foo",
					"inner": map[string]interface{}{
						"key": "bar",
					},
				},
			},
		},
		{
			summary:     "invalid value",
			values:      []string{"component.key^foo"},
			expectedErr: true,
		},
		{
			summary: "multiple values",
			values:  []string{"component.key=foo", "component.inner.key=bar"},
			expected: map[string]interface{}{
				"component": map[string]interface{}{
					"key": "foo",
					"inner": map[string]interface{}{
						"key": "bar",
					},
				},
			},
		},
		{
			summary:    "multiple values with single value file",
			values:     []string{"component.key=foo", "component.inner.key=bar"},
			valueFiles: []string{"testdata/valid-values-1.yaml"},
			expected: map[string]interface{}{
				"component": map[string]interface{}{
					"key": "foo", //value wins
					"inner": map[string]interface{}{
						"key": "bar",
					},
					"outer": map[string]interface{}{
						"inner": map[string]interface{}{
							"key": "baz",
						},
					},
				},
			},
		},
		{
			summary:    "multiple values with multiple value files",
			values:     []string{"component.key=foo", "component.inner.key=bar"},
			valueFiles: []string{"testdata/valid-values-1.yaml", "testdata/valid-values-2.yaml"},
			expected: map[string]interface{}{
				"component": map[string]interface{}{
					"key": "foo", //value wins
					"inner": map[string]interface{}{
						"key": "bar",
					},
					"outer": map[string]interface{}{
						"inner": map[string]interface{}{
							"key": "bzz", //value file testdata/valid-values-2.yaml wins
						},
					},
				},
			},
		},
		{
			summary:     "non existing value file",
			valueFiles:  []string{"testdata/non-existing.yaml"},
			expectedErr: true,
		},
		{
			summary:     "not supported value file",
			valueFiles:  []string{"testdata/dummy.txt"},
			expectedErr: true,
		},
		{
			summary:     "corrupted value file",
			valueFiles:  []string{"testdata/corrupted-values.yaml"},
			expectedErr: true,
		},
		{
			summary: "custom tls key and cert",
			tlsCrt:  "testdata/fake.crt",
			tlsKey:  "testdata/fake.key",
			values:  []string{"global.tlsCrt=foo", "global.tlsKey.key=bar"},
			expected: map[string]interface{}{
				"global": map[string]interface{}{
					"tlsCrt": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUMxVENDQWIyZ0F3SUJBZ0lKQUwrVzNYMUtxdmNETUEwR0NTcUdTSWIzRFFFQkJRVUFNQm94R0RBV0JnTlYKQkFNVEQzZDNkeTVsZUdGdGNHeGxMbU52YlRBZUZ3MHlNVEE1TVRZd09USTVNek5hRncwek1UQTVNVFF3T1RJNQpNek5hTUJveEdEQVdCZ05WQkFNVEQzZDNkeTVsZUdGdGNHeGxMbU52YlRDQ0FTSXdEUVlKS29aSWh2Y05BUUVCCkJRQURnZ0VQQURDQ0FRb0NnZ0VCQUwyWitvQzRobDg5L01rTG1iODNtd2tsR2d2RGhwRklsb0hJcExHYzB3VUIKOTRJdGlvNEJnT3FaUFdXVWdGWGU0UlkyaEZVRUF6djVGWVZSR01OeVp1WEE3bEhWczFGLyt3WXAzbHM2VStkQwoxU053eDRnSUlRN2FWbld5aDFlZDNsckEzd1NLZFBEbUYzaDhUWGREUFVJZm45N0tpYUVVbDBHSU8yVkZUNisxCjJjS3BlaGJRMGhlcDZzdUFNQWU3SGJ4N3Y1WVJzQlhrVDFiVEN5Mkp1R3Y4TEwwM1FIc1VMdVEzN2EybzE4Q2gKVSsrWkZubjVYR2N3OGpVQS8zeDlhb2d3Sk5yZHFSNEZUdExBYnM0djM3d3FDc01zZVkwYjFLMzBYNC95Z3RFZgpwSGV3emFkMUlLamR6UzgveWpTcmlwdCs2Z3k2dVFSNXZhSzRweWxMQ1BFQ0F3RUFBYU1lTUJ3d0dnWURWUjBSCkJCTXdFWUlQZDNkM0xtVjRZVzF3YkdVdVkyOXRNQTBHQ1NxR1NJYjNEUUVCQlFVQUE0SUJBUUFKNVd4TFVzMCsKQ2NQNnJRY3N6Nk15RkwwaEJDYkU3akZ6ODdrY0dTRnhyZ2t1MDN4WTVHb2FOaTllS3laY3hOUmxmZ1U2ODB4Nwo4bmZWNExSVk4xWmwyemRLNWxDQnF6VCtiTVZLamE2dGg3V202ZGJLdnU4RkExUHRMT1c0aXJrdzl4Z0tYRFRGCnlWOE05ai9semIxNm8vRWdTUlZ1R1RJZW4wRWlCZXZaY0VnaXczM2xBaUpvQTZON25IUEdjL3puMm9naThrQngKbWNxOTNpYlMwMWt5cHgzUjlza3o3N1U3TE9weUpVZXJScXRnZFNJREpzQ0Yyb3dQbkh2L29iVHNya3M3QUxvMQpUOWNDTm1yeXNmaDBxNWRHeUZUa0V2bWRTRXBDWkxJUythaHZ2SmxBQkorUXp1SzBoVXM1MHFDWGF0cHhhMCsyCkVaKzhDME5yZEduMQotLS0tLUVORCBDRVJUSUZJQ0FURS0tLS0tCg==",
					"tlsKey": "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb3dJQkFBS0NBUUVBdlpuNmdMaUdYejM4eVF1WnZ6ZWJDU1VhQzhPR2tVaVdnY2lrc1p6VEJRSDNnaTJLCmpnR0E2cGs5WlpTQVZkN2hGamFFVlFRRE8va1ZoVkVZdzNKbTVjRHVVZFd6VVgvN0JpbmVXenBUNTBMVkkzREgKaUFnaER0cFdkYktIVjUzZVdzRGZCSXAwOE9ZWGVIeE5kME05UWgrZjNzcUpvUlNYUVlnN1pVVlByN1had3FsNgpGdERTRjZucXk0QXdCN3Nkdkh1L2xoR3dGZVJQVnRNTExZbTRhL3dzdlRkQWV4UXU1RGZ0cmFqWHdLRlQ3NWtXCmVmbGNaekR5TlFEL2ZIMXFpREFrMnQycEhnVk8wc0J1emkvZnZDb0t3eXg1alJ2VXJmUmZqL0tDMFIra2Q3RE4KcDNVZ3FOM05Mei9LTkt1S20zN3FETHE1QkhtOW9yaW5LVXNJOFFJREFRQUJBb0lCQUVVaEdqUEtrN3V3SnpYSwpWQUZqTGRUVXdUMWV5ZmE0eDUrRVg0QWUxTld6bE9IUzV2ekYwWkkzMHlueFRpV0JBUUtQV0FxRFR3YVQxK1BtCjRLZUtVN2diY3dsRmFIOGpzWXZhd2liekNscDhoS2ZLWEFYZUtPZDRkaU90dHlrYjkxR1JsdjdaMks5b3hVLzUKeW1qY2pENUt0NGlNd2tlSDhXcEVXSnVnL04vc1JVNEFjSlZ6d1RQZVBmUkpvWjhLc1pVL2tDdXorc1VIVGpHUApVRkp0dWllaWZoODg0ekdmY3o3YThINVYxWUxLR2dhSCtGQjlZY2R5MWxRMFBuUlBvOTM3MGUzSlhhT211RU5zCkVSR2pmQVplSTBoM0ZDRjdvclFFTkNZNkpwWDRxenJudHhCWDc0ZWJPRU1sS0wwaDIwL3lMT3VOM2ZHYW9BeEMKM0RiakFNRUNnWUVBOXFWOVFsaFpSL3lQNTIvcUQ0SEtuenRvRi9MTU1mbmhSUDhnTzlONFRkVTBQWUdkSEt2RAphaXh0aWhMNGtlUGh5cFErUlpTTVFndk1BS2Z4eWw0NG93d05GVGl3NTF3UnBrYTFsZTJZUlJvcHZiWnhiVkMzClRvb3pmYlNMMEpsdFVBbFpqS0t5cG9UL0hwWjZTdzdqakt2Q0lLNEpxeVV0aHZMbmdINlhwQ2tDZ1lFQXhNcXUKWkVWSVVXRzRjbU1mVzVSbytNK0NUM3ZqQ25UcGVidVFqS0Z6cWZva1lsMEhPekVueE9XVUNScWpDRmxreGYvWgp1REpTREdOc1FLaktaZW1XMVJMZWZCSEdUSG54TjNRU21UaDY2ckpaS2EySmVFbGhScG4wK0Q1a0d3RmdVTVdFCmVlVnBjREV3Z0xJd3I5Wmw4cFVOaG5kOEtyVkRMTlRnU0wvUWw0a0NnWUE2dGNuTE1SeVBkaDhMQ0NpKzZEWkQKRVBFR1FsVTQwREkvS2p1U0FoUnc4bjhzNU4xeEpiR3VaRVR1eVBWQ1JPeEtQRjlXVUxYU1F0eWNpMTJTdmpyZApGTkZJYStZd0xFcEhPaTJmTXA4OFU2Mzc2cUcxVTdGT2tMY1JCUmtDM29LV3VxTUdSdlFmanlqckx3YU5OMDRTCi9nK0hsK1hWUjFRKyt6TC84eUpGZ1FLQmdRQ29OdWdpNWVZUFNveXptbTh2aFFqRnhmc0pua2hRbytiL0c0bFAKN0tKRjVZQThaSEROOUJLZWgrK21hSko3akk1TGdZdkZtNTN1NFAyanQ2UnF3T1VoZFdPZ2drRVRGaGxPNFhVVQphK2NGdnpYZ0htcW4yM0cvTzlMZWI5WjZEdzhaZS96bGhXZy9jb3lYTmJuUVZHQUluOGhUN01iQ2F2YmsxNEp3CkxTWk1vUUtCZ0Vqa3Job1V2a283Y1RxaW1Vc3ViWTlVWWdMOVlLaFNpL3EyN0tGU0hwdzhtdjJDdW5XQVVGVDgKVzI1SVZSNVdub0tJZVE1enoyVUF4L3orMmRlZHlwN05wRkNHNkRIQ21YRFlDK1RpdmY4aW5WYXp0YXZ2cFhOQQpyTE5USWhFY3YyNTJNWTM4bkFRQmFGbDJPVHp6d1lyNUxkbkp2THNZSTBYTXJRRWhwMzFVCi0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg==",
				},
			},
		},
		{
			summary: "custom domain",
			domain:  "github.com",
			values:  []string{"global.domainName=example.com"},
			expected: map[string]interface{}{
				"global": map[string]interface{}{
					"domainName": "github.com",
				},
			},
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.summary, func(t *testing.T) {
			t.Parallel()

			opts := Sources{
				Values:     tc.values,
				ValueFiles: tc.valueFiles,
				Domain:     tc.domain,
				TLSCrtFile: tc.tlsCrt,
				TLSKeyFile: tc.tlsKey,
			}
			actual, err := Merge(opts, "testdata", clusterinfo.Unrecognized{})

			if tc.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Truef(t, reflect.DeepEqual(tc.expected, actual), "want: %#v\n got: %#v\n", tc.expected, actual)
			}
		})
	}

	t.Run("values file URL", func(t *testing.T) {
		fakeServer := httptest.NewServer(http.FileServer(http.Dir("testdata")))
		defer fakeServer.Close()

		opts := Sources{
			ValueFiles: []string{fmt.Sprintf("%s:/%s", fakeServer.URL, "valid-values-1.yaml")},
		}
		actual, err := Merge(opts, "testdata", clusterinfo.Unrecognized{})

		expected := Values{
			"component": map[string]interface{}{
				"key": "baz",
				"outer": map[string]interface{}{
					"inner": map[string]interface{}{
						"key": "baz",
					},
				},
			},
		}

		require.NoError(t, err)
		require.Truef(t, reflect.DeepEqual(expected, actual), "want: %#v\n got: %#v\n", expected, actual)
	})

	t.Run("k3d", func(t *testing.T) {
		t.Run("default values", func(t *testing.T) {
			actual, err := Merge(Sources{}, "testdata", clusterinfo.K3d{ClusterName: "foo"})

			expected := Values{
				"global": map[string]interface{}{
					"domainName": "local.kyma.dev",
					"tlsCrt":     defaultLocalTLSCrtEnc,
					"tlsKey":     defaultLocalTLSKeyEnc,
				},
				"serverless": map[string]interface{}{
					"dockerRegistry": map[string]interface{}{
						"enableInternal":        false,
						"internalServerAddress": "k3d-foo-registry:5000",
						"serverAddress":         "k3d-foo-registry:5000",
						"registryAddress":       "k3d-foo-registry:5000",
					},
					"containers": map[string]interface{}{
						"manager": map[string]interface{}{
							"envs": map[string]interface{}{
								"functionBuildExecutorArgs": map[string]interface{}{
									"value": "--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--force",
								},
							},
						},
					},
				},
				"istio": map[string]interface{}{
					"helmValues": map[string]interface{}{
						"cni": map[string]string{
							"cniConfDir": "/var/lib/rancher/k3s/agent/etc/cni/net.d",
							"cniBinDir":  "/bin",
						},
					},
				},
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(expected, actual), "want: %#v\n got: %#v\n", expected, actual)
		})

		t.Run("Serverless registry overrides", func(t *testing.T) {
			src := Sources{ValueFiles: []string{"./testdata/registry-overrides.yaml"}}
			actual, err := Merge(src, "testdata", clusterinfo.K3d{ClusterName: "foo"})

			expected := Values{
				"global": map[string]interface{}{
					"domainName": "local.kyma.dev",
					"tlsCrt":     defaultLocalTLSCrtEnc,
					"tlsKey":     defaultLocalTLSKeyEnc,
				},
				"serverless": map[string]interface{}{
					"dockerRegistry": map[string]interface{}{
						"enableInternal":        true,
						"password":              "secret password",
						"internalServerAddress": "internal-address",
						"serverAddress":         "external-address",
						"registryAddress":       "external-push-address",
					},
					"containers": map[string]interface{}{
						"manager": map[string]interface{}{
							"envs": map[string]interface{}{
								"functionBuildExecutorArgs": map[string]interface{}{
									"value": "--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--force",
								},
							},
						},
					},
				},
				"istio": map[string]interface{}{
					"helmValues": map[string]interface{}{
						"cni": map[string]string{
							"cniConfDir": "/var/lib/rancher/k3s/agent/etc/cni/net.d",
							"cniBinDir":  "/bin",
						},
					},
				},
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(expected, actual), "want: %#v\n got: %#v\n", expected, actual)
		})

		t.Run("default values overridden", func(t *testing.T) {
			actual, err := Merge(Sources{
				Values: []string{
					"global.domainName=github.com",
					"global.tlsCrt=github_tls_crt",
					"global.tlsKey=github_tls_key",
					"serverless.dockerRegistry.enableInternal=true",
				},
			}, "testdata", clusterinfo.K3d{ClusterName: "foo"})

			expected := Values{
				"global": map[string]interface{}{
					"domainName": "github.com",
					"tlsCrt":     "github_tls_crt",
					"tlsKey":     "github_tls_key",
				},
				"serverless": map[string]interface{}{
					"dockerRegistry": map[string]interface{}{
						"enableInternal":        true,
						"internalServerAddress": "k3d-foo-registry:5000",
						"serverAddress":         "k3d-foo-registry:5000",
						"registryAddress":       "k3d-foo-registry:5000",
					},
					"containers": map[string]interface{}{
						"manager": map[string]interface{}{
							"envs": map[string]interface{}{
								"functionBuildExecutorArgs": map[string]interface{}{
									"value": "--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--force",
								},
							},
						},
					},
				},
				"istio": map[string]interface{}{
					"helmValues": map[string]interface{}{
						"cni": map[string]string{
							"cniConfDir": "/var/lib/rancher/k3s/agent/etc/cni/net.d",
							"cniBinDir":  "/bin",
						},
					},
				},
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(expected, actual), "want: %#v\n got: %#v\n", expected, actual)
		})

		t.Run("custom domain", func(t *testing.T) {
			actual, err := Merge(Sources{
				Domain: "hello.io",
			}, "testdata", clusterinfo.K3d{ClusterName: "foo"})

			expected := Values{
				"global": map[string]interface{}{
					"domainName": "hello.io",
					"tlsCrt":     defaultLocalTLSCrtEnc,
					"tlsKey":     defaultLocalTLSKeyEnc,
				},
				"serverless": map[string]interface{}{
					"dockerRegistry": map[string]interface{}{
						"enableInternal":        false,
						"internalServerAddress": "k3d-foo-registry:5000",
						"serverAddress":         "k3d-foo-registry:5000",
						"registryAddress":       "k3d-foo-registry:5000",
					},
					"containers": map[string]interface{}{
						"manager": map[string]interface{}{
							"envs": map[string]interface{}{
								"functionBuildExecutorArgs": map[string]interface{}{
									"value": "--insecure,--skip-tls-verify,--skip-unused-stages,--log-format=text,--cache=true,--force",
								},
							},
						},
					},
				},
				"istio": map[string]interface{}{
					"helmValues": map[string]interface{}{
						"cni": map[string]string{
							"cniConfDir": "/var/lib/rancher/k3s/agent/etc/cni/net.d",
							"cniBinDir":  "/bin",
						},
					},
				},
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(expected, actual), "want: %#v\n got: %#v\n", expected, actual)
		})
	})

	t.Run("gardener", func(t *testing.T) {
		t.Run("default values", func(t *testing.T) {
			actual, err := Merge(Sources{}, "testdata", clusterinfo.Gardener{Domain: "foo.gardener.com"})

			expected := Values{
				"global": map[string]interface{}{
					"domainName": "foo.gardener.com",
				},
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(expected, actual), "want: %#v\n got: %#v\n", expected, actual)
		})

		t.Run("custom domain via values", func(t *testing.T) {
			actual, err := Merge(Sources{
				Values: []string{"global.domainName=github.com"},
			}, "testdata", clusterinfo.Gardener{Domain: "foo.gardener.com"})

			expected := Values{
				"global": map[string]interface{}{
					"domainName": "github.com",
				},
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(expected, actual), "want: %#v\n got: %#v\n", expected, actual)
		})

		t.Run("custom domain", func(t *testing.T) {
			actual, err := Merge(Sources{
				Domain: "github.com",
			}, "testdata", clusterinfo.Gardener{Domain: "foo.gardener.com"})

			expected := Values{
				"global": map[string]interface{}{
					"domainName": "github.com",
				},
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(expected, actual), "want: %#v\n got: %#v\n", expected, actual)
		})
	})
	
	t.Run("gke", func(t *testing.T) {
		t.Run("default values", func(t *testing.T) {
			actual, err := Merge(Sources{}, "testdata", clusterinfo.GKE{})

			expected := Values{
				"istio": map[string]interface{}{
					"helmValues": map[string]interface{}{
						"cni": map[string]string{
							"cniBinDir":  "/home/kubernetes/bin",
						},
					},
				},
			}

			require.NoError(t, err)
			require.Truef(t, reflect.DeepEqual(expected, actual), "want: %#v\n got: %#v\n", expected, actual)
		})
	})
}

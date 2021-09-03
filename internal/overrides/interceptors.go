package overrides

import (
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"github.com/kyma-project/cli/internal/gardener"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

const (
	// Default values for local clusters
	localKymaDevDomain    = "local.kyma.dev"
	defaultLocalTLSCrtEnc = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURQVENDQWlXZ0F3SUJBZ0lSQVByWW0wbGhVdUdkeVNCTHo4d3g5VGd3RFFZSktvWklodmNOQVFFTEJRQXcKTURFVk1CTUdBMVVFQ2hNTVkyVnlkQzF0WVc1aFoyVnlNUmN3RlFZRFZRUURFdzVzYjJOaGJDMXJlVzFoTFdSbApkakFlRncweU1EQTNNamt3T1RJek5UTmFGdzB6TURBM01qY3dPVEl6TlROYU1EQXhGVEFUQmdOVkJBb1RER05sCmNuUXRiV0Z1WVdkbGNqRVhNQlVHQTFVRUF4TU9iRzlqWVd3dGEzbHRZUzFrWlhZd2dnRWlNQTBHQ1NxR1NJYjMKRFFFQkFRVUFBNElCRHdBd2dnRUtBb0lCQVFDemE4VEV5UjIyTFRKN3A2aXg0M2E3WTVVblovRkNicGNOQkdEbQpxaDRiRGZLcjFvMm1CYldWdUhDbTVBdTBkeHZnbUdyd0tvZzJMY0N1bEd5UXVlK1JLQ0RIVFBJVjdqZEJwZHJhCkNZMXQrNjlJMkJWV0xiblFNVEZmOWw3Vy8yZFFFU0ExZHZQajhMZmlrcEQvUEQ5ekdHR0FQa2hlenVNRU80dUwKaUxXSloyYmpYK1dtaGZXb0lrOG5oak5YNVBFN2l4alMvNnB3QU56eXk2NW95NDJPaHNuYXlDR1grbmhFVk5SRApUejEraEMvdjJaOS9lRG1OdHdjT1hJSk4relZtUTJ4VHh2Sm0rbDUwYzlnenZTY3YzQXg0dUJsOTk3UnVlcUszCmdZMVRmVklFQ0FOTE9hb29jRG5kcW1FY1lBb25SeGJKK0M2U1RJYlhuUVAyMmYxQkFnTUJBQUdqVWpCUU1BNEcKQTFVZER3RUIvd1FFQXdJRm9EQVRCZ05WSFNVRUREQUtCZ2dyQmdFRkJRY0RBVEFNQmdOVkhSTUJBZjhFQWpBQQpNQnNHQTFVZEVRUVVNQktDRUNvdWJHOWpZV3d1YTNsdFlTNWtaWFl3RFFZSktvWklodmNOQVFFTEJRQURnZ0VCCkFBUnVOd0VadW1PK2h0dDBZSWpMN2VmelA3UjllK2U4NzJNVGJjSGtyQVhmT2hvQWF0bkw5cGhaTHhKbVNpa1IKY0tJYkJneDM3RG5ka2dPY3doNURTT2NrdHBsdk9sL2NwMHMwVmFWbjJ6UEk4Szk4L0R0bEU5bVAyMHRLbE90RwpaYWRhdkdrejhXbDFoRzhaNXdteXNJNWlEZHNpajVMUVJ6Rk04YmRGUUJiRGkxbzRvZWhIRTNXbjJjU3NTUFlDCkUxZTdsM00ySTdwQ3daT2lFMDY1THZEeEszWFExVFRMR2oxcy9hYzRNZUxCaXlEN29qb25MQmJNYXRiaVJCOUIKYlBlQS9OUlBaSHR4TDArQ2Nvb1JndmpBNEJMNEtYaFhxZHZzTFpiQWlZc0xTWk0yRHU0ZWZ1Q25SVUh1bW1xNQpVNnNOOUg4WXZxaWI4K3B1c0VpTUttND0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	defaultLocalTLSKeyEnc = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFb2dJQkFBS0NBUUVBczJ2RXhNa2R0aTB5ZTZlb3NlTjJ1Mk9WSjJmeFFtNlhEUVJnNXFvZUd3M3lxOWFOCnBnVzFsYmh3cHVRTHRIY2I0SmhxOENxSU5pM0FycFJza0xudmtTZ2d4MHp5RmU0M1FhWGEyZ21OYmZ1dlNOZ1YKVmkyNTBERXhYL1plMXY5blVCRWdOWGJ6NC9DMzRwS1EvencvY3hoaGdENUlYczdqQkR1TGk0aTFpV2RtNDEvbApwb1gxcUNKUEo0WXpWK1R4TzRzWTB2K3FjQURjOHN1dWFNdU5qb2JKMnNnaGwvcDRSRlRVUTA4OWZvUXY3OW1mCmYzZzVqYmNIRGx5Q1RmczFaa05zVThieVp2cGVkSFBZTTcwbkw5d01lTGdaZmZlMGJucWl0NEdOVTMxU0JBZ0QKU3ptcUtIQTUzYXBoSEdBS0owY1d5Zmd1a2t5RzE1MEQ5dG45UVFJREFRQUJBb0lCQUJwVmYvenVFOWxRU3UrUgpUUlpHNzM5VGYybllQTFhtYTI4eXJGSk90N3A2MHBwY0ZGQkEyRVVRWENCeXFqRWpwa2pSdGlobjViUW1CUGphCnVoQ0g2ZHloU2laV2FkWEVNQUlIcU5hRnZtZGRJSDROa1J3aisvak5yNVNKSWFSbXVqQXJRMUgxa3BockZXSkEKNXQwL1o0U3FHRzF0TnN3TGk1QnNlTy9TOGVvbnJ0Q3gzSmxuNXJYdUIzT1hSQnMyVGV6dDNRRlBEMEJDY2c3cgpBbEQrSDN6UjE0ZnBLaFVvb0J4S0VacmFHdmpwVURFeThSSy9FemxaVzBxMDB1b2NhMWR0c0s1V1YxblB2aHZmCjBONGRYaUxuOE5xY1k0c0RTMzdhMWhYV3VJWWpvRndZa0traFc0YS9LeWRKRm5acmlJaDB0ZU81Q0I1ZnpaVnQKWklOYndyMENnWUVBd0gzeksvRTdmeTVpd0tJQnA1M0YrUk9GYmo1a1Y3VUlkY0RIVjFveHhIM2psQzNZUzl0MQo3Wk9UUHJ6eGZ4VlB5TVhnOEQ1clJybkFVQk43cE5xeWxHc3FMOFA1dnZlbVNwOGNKU0REQWN4RFlqeEJLams5CldtOXZnTGpnaERSUFN1Um50QXNxQVVqcWhzNmhHUzQ4WUhMOVI2QlI5dmY2U2xWLzN1NWMvTXNDZ1lFQTdwM1UKRDBxcGNlM1liaiszZmppVWFjcTRGcG9rQmp1MTFVTGNvREUydmZFZUtEQldsM3BJaFNGaHYvbnVqaUg2WWJpWApuYmxKNVRlSnI5RzBGWEtwcHNLWW9vVHFkVDlKcFp2QWZGUzc2blZZaUJvMHR3VzhwMGVCS3QyaUFyejRYRmxUCnpRSnNOS1dsRzBzdGJmSzNqdUNzaWJjYTBUd09lbTdSdjdHV0dLTUNnWUJjZmFoVVd1c2RweW9vS1MvbVhEYisKQVZWQnJaVUZWNlVpLzJoSkhydC9FSVpEY3V2Vk56UW8zWm9Jc1R6UXRXcktxOW56VmVxeDV4cnkzd213SXExZwpCMFlVQVhTRlAvV1ZNWEtTbkhWVzdkRUs2S3pmSHZYTitIRjVSbHdLNmgrWGVyd2hsS093VGxyeVAyTEUrS1JtCks1cHJ5aXJZSWpzUGNKbXFncG9IbFFLQmdCVWVFcTVueFNjNERYZDBYQ0Rua1BycjNlN2lKVjRIMnNmTTZ3bWkKVVYzdUFPVTlvZXcxL2tVSjkwU3VNZGFTV3o1YXY5Qk5uYVNUamJQcHN5NVN2NERxcCtkNksrWEVmQmdUK0swSQpNcmxGT1ZpU09TZ1pjZUM4QzBwbjR2YXJFcS9abC9rRXhkN0M2aUhJUFhVRmpna3ZDUllIQm5DT0NCbjl4TUphClRSWlJBb0dBWS9QYSswMFo1MHYrUU40cVhlRHFrS2FLZU80OFUzOHUvQUJMeGFMNHkrZkJpUStuaXh5ZFUzOCsKYndBR3JtMzUvSU5VRTlkWE44d21XRUlXVUZ3YVR2dHY5NXBpcWNKL25QZkFiY2pDeU8wU3BJWCtUYnFRSkljbgpTVjlrKzhWUFNiRUJ5YXRKVTdIQ3FaNUNTWEZuUnRNanliaWNYYUFKSWtBQm4zVjJ3OFk9Ci0tLS0tRU5EIFJTQSBQUklWQVRFIEtFWS0tLS0tCg=="

	// Default vales for remote (non-gardener) clusters
	defaultRemoteKymaDomain = "kyma.example.com"
	defaultRemoteTLSCrtEnc  = "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSURSVENDQWkyZ0F3SUJBZ0lSQU1LTlNRWEZmRFd4WU5ZeUNJcTR4S1V3RFFZSktvWklodmNOQVFFTEJRQXcKTWpFVk1CTUdBMVVFQ2hNTVkyVnlkQzF0WVc1aFoyVnlNUmt3RndZRFZRUURFeEJyZVcxaExXVjRZVzF3YkdVdApZMjl0TUNBWERUSXhNRFF5TVRBNU16YzBNbG9ZRHpJd05qSXdOVEUyTURrek56UXlXakF5TVJVd0V3WURWUVFLCkV3eGpaWEowTFcxaGJtRm5aWEl4R1RBWEJnTlZCQU1URUd0NWJXRXRaWGhoYlhCc1pTMWpiMjB3Z2dFaU1BMEcKQ1NxR1NJYjNEUUVCQVFVQUE0SUJEd0F3Z2dFS0FvSUJBUURHZDFzcUQzMmZOM3grNHlYMTZOMnd6N1BJVjdLagpPdlRWWUJqeW9BRVQ0VlBJbDlOME9Ocm8rTld3T0dSYnU5Y2hjTk9mZ1p0dE1ydjYxU3pGa04rQ2NsM1hjaFNICjRpU3NYTHlud3R0YVAxaXNma2Z3VzZpZ3FFN3pIYXE2SkdjcndHcjE2UDFlS0xUekgyazNacm8xb1A0TUtiNHEKa25mYWFENG1pU2tZRzlsUWxBck1NZ1lmdVlDVmhMSXhkNGhDUHltV2xKRmJXUWt1dUs4b0JVL1ZPbkhvOC91dQpzQVdQOWNVVXBoek5td0FIbUxXUjVUcFdlbGlRaVpyQjNsRWlJSlhSSzNZbmVON1VBZ25JRWxtRUMzOThydDJoCkNQREVudG5oZ1ppNnVVTFhUcE9aNnJwUTZxblIzUyszK2FCbUkwUW1WV1FWZFFTNkVhVzVqYmhCQWdNQkFBR2oKVkRCU01BNEdBMVVkRHdFQi93UUVBd0lGb0RBVEJnTlZIU1VFRERBS0JnZ3JCZ0VGQlFjREFUQU1CZ05WSFJNQgpBZjhFQWpBQU1CMEdBMVVkRVFRV01CU0NFaW91YTNsdFlTNWxlR0Z0Y0d4bExtTnZiVEFOQmdrcWhraUc5dzBCCkFRc0ZBQU9DQVFFQXdua1MxUCt6OThiU0F6SFk5bFlFV1JQck9oaFBXWmpQOU1qZysvS1l4dkE1QmI2SUhIMGoKSGkyREdBQS9QS3BQcGhVQ1c1Zld5anBrSWhXL0VZVE5qNGRMTkc2YXRaN3JjdVZWaTRzcGdJMTdDQmN4ekdyMwphUW1XTkVNYzBhSXRGdUwrQVhscmRzUzJPL2w3aG9YYmZlaUxodk1UTUNiNzRhRlM5R3A4S3RYZjdvWGtpREcvCnllem9kOXh4aTlyMXQ1ekdCSmxHNVZRZmJOYlE2QzdVd0lQMU5BbEljb0xRODdIWm9ReHo1RTdiSWlCRTI0S24KeEJwZ09oUTFYMGdkTEUxRitOZ0JwQmd3Y1RwM2lKMjVvNE92aXEvckkwbGM1RDlwTDBZamVjcFYxS1BPMDhxUQpFaVVoczdTSlFmcm1UVVJ4MU1zUXBFUDEyWWZaMWVmT1VnPT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo="
	defaultRemoteTLSKeyEnc  = "LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcEFJQkFBS0NBUUVBeG5kYktnOTluemQ4ZnVNbDllamRzTSt6eUZleW96cjAxV0FZOHFBQkUrRlR5SmZUCmREamE2UGpWc0Roa1c3dlhJWERUbjRHYmJUSzcrdFVzeFpEZmduSmQxM0lVaCtJa3JGeThwOExiV2o5WXJINUgKOEZ1b29LaE84eDJxdWlSbks4QnE5ZWo5WGlpMDh4OXBOMmE2TmFEK0RDbStLcEozMm1nK0pva3BHQnZaVUpRSwp6RElHSDdtQWxZU3lNWGVJUWo4cGxwU1JXMWtKTHJpdktBVlAxVHB4NlBQN3JyQUZqL1hGRktZY3pac0FCNWkxCmtlVTZWbnBZa0ltYXdkNVJJaUNWMFN0MkozamUxQUlKeUJKWmhBdC9mSzdkb1Fqd3hKN1o0WUdZdXJsQzEwNlQKbWVxNlVPcXAwZDB2dC9tZ1ppTkVKbFZrRlhVRXVoR2x1WTI0UVFJREFRQUJBb0lCQVFDSjhndjdnQ2pnc2NCbQpzWnVCQVFxV0NzZjdTSGx4MjFpeHRzbWdXblpsU3dqaE5DWlZjZTgyWHo2bjdZcFQrSXZmUW56Vk1WREc1YXlpCisralNxWSt4SzZ6dVF1emlSZDBYc0oyd1BWQVp1azM0Rnc0Smtxdnlmd25oRVkzSk0rUkNGTXhEZ0Y0YlJGQUIKYktQRlRqRy9kTmNmdlNQZ2swMmJFVG1ocjFSUTNFL2F1bURYMS9mVldkVzduMjZndVpkdjlhR0NQMy9QMThtNApFMXlUMVRaYVdUcVVSQ09sU2J5WUh0R2xvdFdtVm14b3A5SktVMVhoVk11bVhBUnZjRUN5blpBY3g0L1cvKzV6Ck81Zld4cEZtZ0QwOEtUUEdoN0R0QjB0eXlkTlZXZXBjM0VZOSs0dFVXamNSTXRKRVpST29mWGNIelV0OFN4c2EKZU1zQmdDUUJBb0dCQU0zRXAzWS9vbi9IU0ovUjlpRUpQNE95SjlnTEw0UXEzR01UcnUzSFBBRDlrY3g3STg5YwpqeU8yMnorNEFJL3RzZFo3TDFubHJrSVlNR3FYYlByQXBTek1oQXBFWUl5cG10aEVneWNTZE9yYmFMaTVOQmpLCnBNMDBhaXgyZW5SeEFwOFEyMkkrTml0Wm1jYmFhTkhERHdIaDdBVnpGVnN0dmJZN0xEcEp2bkRaQW9HQkFQYnEKWHltSVgvV1hNZlpRVjJPUk1zR3kvd2czVUluU05YZ1JWOTU0c2Q4cVRBQnc5N0Z3TFJuRWNxMUMwSTdrS0xoQgpaU2tKUDk3Y0lINXBHNnBIOXBuWHZBQmpCdWc5cW8rclozSEI4Vjd3ZU5waTNzSnBRVFZYUFo4bmVBdStuN1pDCnJsZWJMYUNYUUViN2lRZHA4elRJLzJoa291MXRCM2tXQjBuV3kyR3BBb0dCQU1CdVBaSFhWdmVhZmUrQW9tWW8KeU80M2FRMmhBRkhnNTNQOGoyWXRJWTluazdjZ0hkQXBwbTltN1VsOG9ZSDRiNHkrYlB6c1QvZmR1VUdsMVRQMwpmMEVURGhTdjkzNzBpaXZnZnFyR2x2S2dPQ0l3aVdqNThmODZHbVQwYy9aN1RWRkdxWFFKN0F6RVlZeFc2eG5vCkNodmZsU05QaWRSWVJZZXJkT1FaM1BDWkFvR0FSY3BoTTRBVWYzcEk2UEkwZ1RRZFFKcXpjME1QUktWaDc1b1gKV0E2TldDTEFjSzk5azIyOWtiYnhJdi9ycXpmYU9wcGhXWVAveGFJNm5RQmdqWFRod3dJelpYaVlEelMrN1BUcAp2RUd4VThCc3FHMmh3Um0zRUxpajlrUlZyaHduVUlEd2ZscWlQdTRCZ1E2LzRKU1Y1YW1hWjR0cWNlbUxYekpXCnhRd3RXR0VDZ1lBbWUvM0c1RXN4T0ZkUmltcTRwYjlMMjJzVnFqTFFVcUw1R09QVWNBYlZERDJ3SEU0K0pQdEsKUXdPTVJBZ05oeEJPMmRxWGk0SXgxWmRuRmE5dENFTmNrQzFRL2VmZzUwN2dhdk44ZGJGaGQ1dGpBTU9zVGVDOQp6SnowUFc2VktOMkptd0JWMlhIeWlDTHJ2ZjlnZVVtSHhzR3A5dEhkRlc5S2g1allXNjRRMEE9PQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo="
)

// OverrideInterceptor is controlling access to override values
type OverrideInterceptor interface {
	//String shows the value of the override
	String(value interface{}, key string) string
	//Intercept is executed when the override is retrieved
	Intercept(value interface{}, key string) (interface{}, error)
	//Undefined is executed when the override is not defined
	Undefined(overrides map[string]interface{}, key string) error
}

// DomainNameOverrideInterceptor resolves the domain name for the cluster
type DomainNameOverrideInterceptor struct {
	kubeClient     kubernetes.Interface
	isLocalCluster func() (bool, error) // Returns true if we're on a local cluster like k3s
}

func NewDomainNameOverrideInterceptor(kubeClient kubernetes.Interface) *DomainNameOverrideInterceptor {
	return &DomainNameOverrideInterceptor{
		kubeClient: kubeClient,
		isLocalCluster: func() (bool, error) {
			return IsK3dCluster(kubeClient)
		},
	}
}

func (i *DomainNameOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (i *DomainNameOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	// On gardener, domain provided by user should be ignored
	domainName, err := gardener.Domain(i.kubeClient)
	if err != nil {
		return nil, err
	}

	if domainName != "" {
		return domainName, nil
	}

	// In every other environment, proceed with what was provided by the user.
	return value, nil
}

func (i *DomainNameOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	domain, err := i.getDomainName()
	if err != nil {
		return err
	}

	return NewFallbackOverrideInterceptor(domain).Undefined(overrides, key)
}

func (i *DomainNameOverrideInterceptor) getDomainName() (domainName string, err error) {

	// On gardener always return gardener domain
	domainName, err = gardener.Domain(i.kubeClient)
	if err != nil {
		return "", err
	}
	if domainName != "" {
		return domainName, nil
	}

	// On local k3s cluster return local development domain
	domainName, err = i.findLocalDomain()
	if err != nil {
		return "", err
	}
	if domainName != "" {
		return domainName, nil
	}
	return defaultRemoteKymaDomain, nil
}

func (i *DomainNameOverrideInterceptor) findLocalDomain() (domainName string, err error) {

	isLocalCluster, err := i.isLocalCluster()
	if err != nil {
		return "", err
	}

	if isLocalCluster {
		return localKymaDevDomain, nil
	}

	return "", nil
}

// CertificateOverrideInterceptor handles certificates
type CertificateOverrideInterceptor struct {
	tlsCrtOverrideKey string
	tlsKeyOverrideKey string
	tlsCrtEnc         string
	tlsKeyEnc         string
	isLocalCluster    func() (bool, error)
	isGardenerCluster func() (bool, error)
}

func NewCertificateOverrideInterceptor(tlsCrtOverrideKey, tlsKeyOverrideKey string, kubeClient kubernetes.Interface) *CertificateOverrideInterceptor {
	res := &CertificateOverrideInterceptor{
		tlsCrtOverrideKey: tlsCrtOverrideKey,
		tlsKeyOverrideKey: tlsKeyOverrideKey,
	}

	res.isLocalCluster = func() (bool, error) {
		return IsK3dCluster(kubeClient)
	}

	res.isGardenerCluster = func() (bool, error) {
		gardenerDomain, err := gardener.Domain(kubeClient)
		if err != nil {
			return false, err
		}
		return gardenerDomain != "", nil
	}

	return res
}

func (i *CertificateOverrideInterceptor) String(value interface{}, key string) string {
	return "<masked>"
}

func (i *CertificateOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	isGardener, err := i.isGardenerCluster()
	if err != nil {
		return "", err
	}
	if isGardener {
		return "", nil
	}

	switch key {
	case i.tlsCrtOverrideKey:
		i.tlsCrtEnc = value.(string)
	case i.tlsKeyOverrideKey:
		i.tlsKeyEnc = value.(string)
	}
	if err := i.validate(); err != nil {
		return nil, err
	}
	return value, nil
}

func (i *CertificateOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	isGardener, err := i.isGardenerCluster()
	if err != nil {
		return err
	}
	if isGardener {
		return nil
	}

	isLocalCluster, err := i.isLocalCluster()
	if err != nil {
		return err
	}

	var fbInterc *FallbackOverrideInterceptor
	switch key {
	case i.tlsCrtOverrideKey:
		var val string
		if isLocalCluster {
			val = defaultLocalTLSCrtEnc
		} else {
			val = defaultRemoteTLSCrtEnc
		}
		fbInterc = NewFallbackOverrideInterceptor(val)
		i.tlsCrtEnc = val
	case i.tlsKeyOverrideKey:
		var val string
		if isLocalCluster {
			val = defaultLocalTLSKeyEnc
		} else {
			val = defaultRemoteTLSKeyEnc
		}
		fbInterc = NewFallbackOverrideInterceptor(val)
		i.tlsKeyEnc = val
	default:
		return fmt.Errorf("certificate interceptor can not handle overrides-key '%s'", key)
	}

	if err := fbInterc.Undefined(overrides, key); err != nil {
		return err
	}
	return i.validate()
}

func (i *CertificateOverrideInterceptor) validate() error {
	if i.tlsCrtEnc != "" && i.tlsKeyEnc != "" {
		// Decode tls crt and key
		crt, err := base64.StdEncoding.DecodeString(i.tlsCrtEnc)
		if err != nil {
			return err
		}
		key, err := base64.StdEncoding.DecodeString(i.tlsKeyEnc)
		if err != nil {
			return err
		}
		// Ensure that crt and key are fitting together
		_, err = tls.X509KeyPair(crt, key)
		if err != nil {
			return errors.Wrap(err,
				fmt.Sprintf("Provided TLS certificate (passed in keys '%s' and '%s') is invalid", i.tlsCrtOverrideKey, i.tlsKeyOverrideKey))
		}
	}
	return nil
}

// FallbackOverrideInterceptor sets a default value for an undefined overwrite
type FallbackOverrideInterceptor struct {
	fallback interface{}
}

func (i *FallbackOverrideInterceptor) String(value interface{}, key string) string {
	return fmt.Sprintf("%v", value)
}

func (i *FallbackOverrideInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return value, nil
}

func (i *FallbackOverrideInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	subKeys := strings.Split(key, ".")
	maxDepth := len(subKeys)
	lastProcessedEntry := overrides

	for depth, subKey := range subKeys {
		if _, ok := lastProcessedEntry[subKey]; !ok {
			// Subelement does not exist - add map
			lastProcessedEntry[subKey] = make(map[string]interface{})
		}
		if _, ok := lastProcessedEntry[subKey].(map[string]interface{}); !ok {
			// Ensure existing sub-element is map otherwise fail
			return fmt.Errorf("override '%s' cannot be set with default value as sub-key '%s' is not a map", key, strings.Join(subKeys[:depth+1], "."))
		}

		if depth == (maxDepth - 1) {
			// We are in the last loop, set default value
			lastProcessedEntry[subKey] = i.fallback
		} else {
			// Continue processing the next sub-entry
			lastProcessedEntry = lastProcessedEntry[subKey].(map[string]interface{})
		}
	}

	return nil
}

func (i *FallbackOverrideInterceptor) Fallback() interface{} {
	return i.fallback
}

func NewFallbackOverrideInterceptor(fallback interface{}) *FallbackOverrideInterceptor {
	return &FallbackOverrideInterceptor{
		fallback: fallback,
	}
}

type RegistryDisableInterceptor struct {
	kubeClient kubernetes.Interface
}

func NewRegistryDisableInterceptor(kubeClient kubernetes.Interface) *RegistryDisableInterceptor {
	return &RegistryDisableInterceptor{
		kubeClient: kubeClient,
	}
}
func (i *RegistryDisableInterceptor) String(value interface{}, key string) string {
	newVal, err := i.Intercept(value, key)
	if err != nil {
		return fmt.Sprintf("error during interception: %s", err.Error())
	}
	return fmt.Sprintf("%v", newVal)
}

func (i *RegistryDisableInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	k3dCluster, err := IsK3dCluster(i.kubeClient)
	if err != nil {
		return nil, err
	}
	if k3dCluster {
		return "false", nil
	}
	return value, nil
}

func (i *RegistryDisableInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	k3dCluster, err := IsK3dCluster(i.kubeClient)
	if err != nil {
		return err
	}
	if k3dCluster {
		return NewFallbackOverrideInterceptor(false).Undefined(overrides, key)
	}
	return nil
}

type RegistryInterceptor struct {
	kubeClient kubernetes.Interface
}

func NewRegistryInterceptor(kubeClient kubernetes.Interface) *RegistryInterceptor {
	return &RegistryInterceptor{
		kubeClient: kubeClient,
	}
}

func (i *RegistryInterceptor) String(value interface{}, key string) string {
	newVal, err := i.Intercept(value, key)
	if err != nil {
		return fmt.Sprintf("error during interception: %s", err.Error())
	}
	return fmt.Sprintf("%v", newVal)
}

func (i *RegistryInterceptor) Intercept(value interface{}, key string) (interface{}, error) {
	return value, nil
}

func (i *RegistryInterceptor) Undefined(overrides map[string]interface{}, key string) error {
	k3dCluster, err := IsK3dCluster(i.kubeClient)
	if err != nil {
		return err
	}
	if k3dCluster {
		k3dClusterName, err := K3dClusterName(i.kubeClient)
		if err != nil {
			return err
		}
		return NewFallbackOverrideInterceptor(fmt.Sprintf("k3d-%s-registry:5000", k3dClusterName)).Undefined(overrides, key)
	}
	return nil
}

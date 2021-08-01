// package trust provides trusted certificate management.
package trust

// Certifier defines the contract to manage digital certificates in Kyma CLI.
type Certifier interface {

	// Certificate provides the decoded Kyma root certificate.
	Certificate() ([]byte, error)

	// CertificateAlpha provides the decoded Kyma root certificate for alpha commands.
	CertificateAlpha() ([]byte, error)

	// CertificateKyma2 provides the decoded Kyma root certificate for kyma 2.
	CertificateKyma2() ([]byte, error)

	// StoreCertificate imports the given certificate file into the trusted root certificates manager of the OS.
	StoreCertificate(file string, info Informer) error

	// Instructions provides instructions on how to manually store a certificate.
	// Use in case it can not be stored by calling StoreCertificate.
	Instructions() string

	// InstructionsAlpha provides instructions on how to manually store a certificate for alpha commands.
	// Use in case it can not be stored by calling StoreCertificate.
	InstructionsAlpha() string

	// InstructionsKyma2 provides instructions on how to manually store a certificate for kyma 2.
	// Use in case it can not be stored by calling StoreCertificate.
	InstructionsKyma2() string
}

// informer defines the way certification management informs about its progress.
type Informer interface {
	LogInfo(msg string)
	LogInfof(format string, args ...interface{})
}

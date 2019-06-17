package trust

// Certifier defines the contract to manage digital certificates in Kyma CLI
type Certifier interface {
	// StoreCertificate imports the given certificate file into the trusted root certificates manager of the OS.
	StoreCertificate(file string) error

	// Instructions provides instructions on how to manually store a certificate.
	// Use in case it can not be stored by calling StoreCertificate
	Instructions() string
}

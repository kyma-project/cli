package main

import (
	"encoding/base64"
	"io/ioutil"
	"log"
	"os"

	"github.com/kyma-project/cli/internal/kubectl"
	"github.com/kyma-project/cli/internal/trust"
)

func main() {

	k8s := kubectl.NewWrapper(false)

	// get cert from cluster
	cert, err := k8s.RunCmd("get", "configmap", "net-global-overrides", "-n", "kyma-installer", "-o", "jsonpath='{.data.global\\.ingress\\.tlsCrt}'")
	if err != nil {
		log.Fatal(err)
	}

	decodedCert, err := base64.StdEncoding.DecodeString(cert)
	if err != nil {
		log.Fatal(err)
	}
	tmpFile, err := ioutil.TempFile(os.TempDir(), "kyma-*.crt")
	if err != nil {
		log.Fatal("Cannot create temporary file for Kyma certificate", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err = tmpFile.Write(decodedCert); err != nil {
		log.Fatal("Failed to write the kyma certificate", err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	// store certificate
	if err := trust.NewCertifier().StoreCertificate(tmpFile.Name()); err != nil {
		log.Fatal(err)
	}
}

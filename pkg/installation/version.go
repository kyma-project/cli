package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ArtifactMeta struct {
	Versions []string `xml:"Contents>Key"`
}

func getDataBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("Status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("Read body: %v", err)
	}

	return data, nil
}
func increasePatchVersion(version string) string {
	verArray := strings.Split(version, ".")
	patchVer, _ := strconv.Atoi(verArray[len(verArray)-1])
	return fmt.Sprintf("%s.%s.%d", verArray[0], verArray[1], patchVer+1)
}

func setToLatestPatchVersion(version string) string {
	if xmlBytes, err := getDataBytes("https://storage.googleapis.com/kyma-prow-artifacts/"); err != nil {
		log.Printf("Failed to get XML: %v", err)
	} else {
		v := ArtifactMeta{}
		xml.Unmarshal(xmlBytes, &v)
		newVer := increasePatchVersion(version)
		for _, ver := range v.Versions {
			if strings.Contains(ver, newVer) {
				version = newVer
				newVer = increasePatchVersion(newVer)
			}
		}
	}
	return version
}

func main() {
	version := "1.7.0"
	version = setToLatestPatchVersion(version)
	fmt.Print(version)
}

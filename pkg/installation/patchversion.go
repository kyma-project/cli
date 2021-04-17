package installation

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type ReleaseInfo struct {
	ProwXMLFile string
}

func NewReleaseInfo() ReleaseInfo {
	release := ReleaseInfo{}
	release.ProwXMLFile = "https://storage.googleapis.com/kyma-prow-artifacts/"
	return release
}

func (r ReleaseInfo) GetVersions() []string {
	if xmlBytes, err := getDataBytes(r.ProwXMLFile); err != nil {
		log.Printf("Failed to get XML: %v", err)
	} else {
		v := struct {
			Versions []string `xml:"Contents>Key"`
		}{}
		xml.Unmarshal(xmlBytes, &v)
		return v.Versions
	}
	return make([]string, 0) // skip patch update
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

func findLatestPatchVersion(version string, versions []string) string {
	newVer := increasePatchVersion(version)
	for _, ver := range versions {
		if strings.Contains(ver, newVer) {
			version = newVer
			newVer = increasePatchVersion(newVer)
		}
	}
	return version
}

func setToLatestPatchVersion(version string) string {
	release := NewReleaseInfo()
	versions := release.GetVersions()
	return findLatestPatchVersion(version, versions)
}

package installation

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

func GetReleaseVersions() ([]string, error) {
	const url = "https://storage.googleapis.com/kyma-prow-artifacts/"
	xmlBytes, err := getDataBytes(url)
	if err != nil {
		log.Printf("Failed to get XML: %v", err)
		return make([]string, 0), err // skip patch update
	}
	v := struct {
		Versions []string `xml:"Contents>Key"`
	}{}
	err = xml.Unmarshal(xmlBytes, &v)
	return v.Versions, err
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

func updatePatchVersion(version string, patchVer int) string {
	verArray := strings.Split(version, ".")
	return fmt.Sprintf("%s.%s.%d", verArray[0], verArray[1], patchVer)
}

func getPatchVersion(version string) int {
	verArray := strings.Split(version, ".")
	re := regexp.MustCompile(`[-]?\d[\d,]*[\.]?[\d{2}]*`)
	patchString := re.FindAllString(verArray[2], 1)[0]
	patchVer, _ := strconv.Atoi(patchString)
	return patchVer
}

func getMajorVersion(version string) string {
	verArray := strings.Split(version, ".")
	return fmt.Sprintf("%s.%s", verArray[0], verArray[1])
}

func FindLatestPatchVersion(version string, versions []string) string {
	currPatchVer := getPatchVersion(version)
	majorVer := getMajorVersion(version)
	for _, ver := range versions {
		if strings.Contains(ver, majorVer) {
			loopPatchVer := getPatchVersion(ver)
			if loopPatchVer > currPatchVer {
				currPatchVer = loopPatchVer
			}
		}
	}
	return updatePatchVersion(version, currPatchVer)
}

func SetToLatestPatchVersion(version string) string {
	versions, _ := GetReleaseVersions()
	return FindLatestPatchVersion(version, versions)
}

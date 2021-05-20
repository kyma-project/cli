package installation

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
)

type tagStruct struct {
	Tag        string `json:"tag_name"`
	IsPrelease bool   `json:"prerelease"`
}

func getDataBytes() ([]byte, error) {
	const url = "https://api.github.com/repos/kyma-project/kyma/releases"
	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, fmt.Errorf("GET error: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []byte{}, fmt.Errorf("status error: %v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, fmt.Errorf("read body: %v", err)
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

// get Kyma version that is compatible with CLI
func findKymaPatchVersion(version string, versions []tagStruct) string {
	currPatchVer := 0
	cliPatchVer := getPatchVersion(version)
	majorVer := getMajorVersion(version)
	for _, ver := range versions {
		if strings.Contains(ver.Tag, majorVer) && !ver.IsPrelease {
			loopPatchVer := getPatchVersion(ver.Tag)
			if (loopPatchVer <= cliPatchVer) && (loopPatchVer > currPatchVer) {
				currPatchVer = loopPatchVer
			}
		}
	}
	return updatePatchVersion(version, currPatchVer)
}

func getReleaseTags() ([]tagStruct, error) {
	xmlBytes, err := getDataBytes()
	if err != nil {
		log.Printf("Failed to get XML: %v", err)
		return make([]tagStruct, 0), err // skip patch update
	}
	v := []tagStruct{}
	err = json.Unmarshal(xmlBytes, &v)
	return v, err
}

// Find latest compatible Kyma version to allow CLI patch updates without Kyma release
func SetKymaSemVersion(cliVersion string) string {
	if isSemVer(cliVersion) {
		versions, _ := getReleaseTags()
		return findKymaPatchVersion(cliVersion, versions)
	}
	return cliVersion
}

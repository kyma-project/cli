package module

import (
	"encoding/json"
	"errors"
	"fmt"
	v2 "github.com/gardener/component-spec/bindings-go/apis/v2"
	"github.com/gardener/component-spec/bindings-go/ctf"
	"github.com/mandelsoft/vfs/pkg/vfs"
	"os"
	"path"
	"sigs.k8s.io/yaml"
	"strings"
)

var ErrFailedToParseImageURL = errors.New("error parsing protecode image URL")

const (
	secConfigFileName = "sec-scanners-config.yaml"
	secScanLabelKey   = "scan.security.kyma-project.io"
	secLabelKey       = "security.kyma-project.io"
	secScanEnabled    = "enabled"
)

var labelTemplate = "%s." + secScanLabelKey
var globalLabelTemplate = "%s." + secLabelKey

func AddSecurityScanningMetadata(descriptor *v2.ComponentDescriptor, modDef *Definition, fs vfs.FileSystem,
	securityConfigPath string,
) error {
	//parse security config file
	config, err := parseSecurityScanConfig(securityConfigPath)
	if err != nil {
		return err
	}
	excludedWhitesourcePathPatterns := strings.Join(config.WhiteSource.Exclude, ",")

	// add security scan enabled global label
	err = appendLabelToAccessor(descriptor, "scan", secScanEnabled, globalLabelTemplate)
	if err != nil {
		return err
	}
	if len(descriptor.Sources) == 0 {
		return errors.New("found no sources in component descriptor")
	}
	//add whitesource sec scan labels
	for srcIdx := range descriptor.Sources {
		src := &descriptor.Sources[srcIdx]
		err := appendLabelToAccessor(src, "language", config.WhiteSource.Language, labelTemplate)
		if err != nil {
			return err
		}
		err = appendLabelToAccessor(src, "subprojects", config.WhiteSource.SubProjects, labelTemplate)
		if err != nil {
			return err
		}
		err = appendLabelToAccessor(src, "exclude", excludedWhitesourcePathPatterns, labelTemplate)
		if err != nil {
			return err
		}
	}
	//add protecode sec scan images
	if err := appendProtecodeImagesLayers(descriptor, config); err != nil {
		return err
	}

	return WriteComponentDescriptor(fs, descriptor, modDef.ArchivePath,
		ctf.ComponentDescriptorFileName)
}

func appendProtecodeImagesLayers(descriptor *v2.ComponentDescriptor, config *SecurityScanCfg) error {
	for _, imageURL := range config.Protecode {
		imageName, err := getImageName(imageURL)
		if err != nil {
			return err
		}

		imageTypeLabel, err := generateOCMLabel("type", "third-party-image", labelTemplate)
		if err != nil {
			return err
		}
		imageLayerMetadata := v2.IdentityObjectMeta{
			Name:    imageName,
			Type:    v2.OCIImageType,
			Version: descriptor.Version,
			Labels:  []v2.Label{imageTypeLabel},
		}
		imageLayerAccess, err := v2.NewUnstructured(v2.NewOCIRegistryAccess(imageURL))
		if err != nil {
			return err
		}
		protecodeImageLayerResource := v2.Resource{
			IdentityObjectMeta: imageLayerMetadata,
			Relation:           v2.ExternalRelation,
			Access:             &imageLayerAccess,
		}
		descriptor.Resources = append(descriptor.Resources, protecodeImageLayerResource)
	}
	return nil
}

func generateOCMLabel(prefix, value, tpl string) (v2.Label, error) {
	labelValue, err := json.Marshal(map[string]string{
		prefix: value,
	})
	if err != nil {
		return v2.Label{}, err
	}
	return v2.Label{
		Name:  fmt.Sprintf(tpl, prefix),
		Value: labelValue,
	}, nil
}

func getImageName(imageURL string) (string, error) {
	imageTagSlice := strings.Split(imageURL, ":")
	if len(imageTagSlice) != 2 {
		return "", ErrFailedToParseImageURL
	}
	repoImageSlice := strings.Split(imageTagSlice[0], "/")
	l := len(repoImageSlice)
	if l == 0 {
		return "", ErrFailedToParseImageURL
	}

	return repoImageSlice[l-1], nil
}

type SecurityScanCfg struct {
	ModuleName  string            `json:"module-name"`
	Protecode   []string          `json:"protecode"`
	WhiteSource WhiteSourceSecCfg `json:"whitesource"`
}
type WhiteSourceSecCfg struct {
	Language    string   `json:"language"`
	SubProjects string   `json:"subprojects"`
	Exclude     []string `json:"exclude"`
}

func parseSecurityScanConfig(securityConfigPath string) (*SecurityScanCfg, error) {
	configFilePath := path.Join(securityConfigPath, secConfigFileName)
	fileBytes, err := os.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}
	secCfg := &SecurityScanCfg{}
	if err := yaml.Unmarshal(fileBytes, secCfg); err != nil {
		return nil, err
	}
	return secCfg, nil
}

func appendLabelToAccessor(labeled v2.LabelsAccessor, prefix, value, tpl string) error {
	labels := labeled.GetLabels()
	labelValue, err := generateOCMLabel(prefix, value, tpl)
	if err != nil {
		return err
	}
	labels = append(labels, labelValue)
	labeled.SetLabels(labels)
	return nil
}

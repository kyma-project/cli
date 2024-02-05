package module

import (
	"errors"
	"fmt"
	"os"
	"strings"

	itociartifact "github.com/open-component-model/ocm/cmds/ocm/commands/ocmcmds/common/inputs/types/ociartifact"
	"github.com/open-component-model/ocm/pkg/contexts/ocm/accessmethods/ociartifact"
	ocm "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc"
	ocmv1 "github.com/open-component-model/ocm/pkg/contexts/ocm/compdesc/meta/v1"
	"sigs.k8s.io/yaml"
)

var ErrFailedToParseImageURL = errors.New("error parsing protecode image URL")

const (
	SecScanLabelKey = "scan.security.kyma-project.io"
	secLabelKey     = "security.kyma-project.io"
	secScanEnabled  = "enabled"
)

var labelTemplate = SecScanLabelKey + "/%s"
var globalLabelTemplate = secLabelKey + "/%s"

func AddSecurityScanningMetadata(descriptor *ocm.ComponentDescriptor, securityConfigPath string) error {
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
		err = appendLabelToAccessor(src, "rc-tag", config.RcTag, labelTemplate)
		if err != nil {
			return err
		}
		err = appendLabelToAccessor(src, "language", config.WhiteSource.Language, labelTemplate)
		if err != nil {
			return err
		}
		err = appendLabelToAccessor(src, "dev-branch", config.DevBranch, labelTemplate)
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

	ocm.DefaultResources(descriptor)
	return ocm.Validate(descriptor)
}

func appendProtecodeImagesLayers(descriptor *ocm.ComponentDescriptor, config *SecurityScanCfg) error {
	for _, imageURL := range config.Protecode {
		imageName, imageTag, err := getImageName(imageURL)
		if err != nil {
			return err
		}

		imageTypeLabel, err := generateOCMLabel("type", "third-party-image", labelTemplate)
		if err != nil {
			return err
		}

		access := ociartifact.New(imageURL)
		access.SetType(ociartifact.LegacyType)
		protecodeImageLayerResource := ocm.Resource{
			ResourceMeta: ocm.ResourceMeta{
				ElementMeta: ocm.ElementMeta{
					Name:    imageName,
					Labels:  []ocmv1.Label{*imageTypeLabel},
					Version: imageTag,
				},
				Type:     itociartifact.LEGACY_TYPE,
				Relation: ocmv1.ExternalRelation,
			},
			Access: access,
		}
		descriptor.Resources = append(descriptor.Resources, protecodeImageLayerResource)
	}
	return nil
}

func generateOCMLabel(key, value, tpl string) (*ocmv1.Label, error) {
	return ocmv1.NewLabel(fmt.Sprintf(tpl, key), value, ocmv1.WithVersion("v1"))
}

func getImageName(imageURL string) (string, string, error) {
	imageTagSlice := strings.Split(imageURL, ":")
	if len(imageTagSlice) != 2 {
		return "", "", ErrFailedToParseImageURL
	}
	repoImageSlice := strings.Split(imageTagSlice[0], "/")
	if len(repoImageSlice) == 0 {
		return "", "", ErrFailedToParseImageURL
	}

	return repoImageSlice[len(repoImageSlice)-1], imageTagSlice[len(imageTagSlice)-1], nil
}

type SecurityScanCfg struct {
	ModuleName  string            `json:"module-name" yaml:"module-name" comment:"string, name of your module"`
	Protecode   []string          `json:"protecode" yaml:"protecode" comment:"list, includes the images which must be scanned by the Protecode scanner (aka. Black Duck Binary Analysis)"`
	WhiteSource WhiteSourceSecCfg `json:"whitesource" yaml:"whitesource" comment:"whitesource (aka. Mend) security scanner specific configuration"`
	DevBranch   string            `json:"dev-branch" yaml:"dev-branch" comment:"string, name of the development branch"`
	RcTag       string            `json:"rc-tag" yaml:"rc-tag" comment:"string, release candidate tag"`
}
type WhiteSourceSecCfg struct {
	Language    string   `json:"language" yaml:"language" comment:"string, indicating the programming language the scanner has to analyze"`
	SubProjects string   `json:"subprojects" yaml:"subprojects" comment:"string, specifying any subprojects"`
	Exclude     []string `json:"exclude" yaml:"exclude" comment:"list, directories within the repository which should not be scanned"`
}

func parseSecurityScanConfig(securityConfigPath string) (*SecurityScanCfg, error) {
	fileBytes, err := os.ReadFile(securityConfigPath)
	if err != nil {
		return nil, err
	}
	secCfg := &SecurityScanCfg{}
	if err := yaml.Unmarshal(fileBytes, secCfg); err != nil {
		return nil, err
	}
	return secCfg, nil
}

func appendLabelToAccessor(labeled ocm.LabelsAccessor, key, value, tpl string) error {
	labels := labeled.GetLabels()
	labelValue, err := generateOCMLabel(key, value, tpl)
	if err != nil {
		return err
	}
	labels = append(labels, *labelValue)
	labeled.SetLabels(labels)
	return nil
}

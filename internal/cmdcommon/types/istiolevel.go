package types

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	istioanalysisdiag "istio.io/istio/pkg/config/analysis/diag"
)

const (
	InfoIstioLevel    IstioLevel = "info"
	WarningIstioLevel IstioLevel = "warning"
	ErrorIstioLevel   IstioLevel = "error"
)

var (
	availableIstioLevels = []IstioLevel{
		InfoIstioLevel,
		WarningIstioLevel,
		ErrorIstioLevel,
	}
)

type IstioLevel string

func (f *IstioLevel) String() string {
	return string(*f)
}

func (f *IstioLevel) Set(v string) error {
	for _, format := range availableIstioLevels {
		if v == format.String() {
			*f = IstioLevel(v)
			return nil
		}
	}

	return errors.New(fmt.Sprintf("invalid istio level '%s'", v))
}

func (f *IstioLevel) Type() string {
	return "string"
}

func (f *IstioLevel) ToInternalIstioLevel() istioanalysisdiag.Level {
	levelName := f.String()
	if levelName == "" {
		return istioanalysisdiag.Warning
	}
	return istioanalysisdiag.GetUppercaseStringToLevelMap()[strings.ToUpper(levelName)]
}

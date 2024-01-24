package version

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
)

var (
	currentVersionNoReleaseError = versionCompareError{
		msg: func(e *versionCompareError) string {
			return fmt.Sprintf("Current semanticVersion seems not to be an official Kyma release: %s", e.currentVersion)
		},
	}
	nextVersionNoReleaseError = versionCompareError{
		msg: func(e *versionCompareError) string {
			return fmt.Sprintf("Next semanticVersion seems not to be an official Kyma release: %s", e.nextVersion)
		},
	}
	nextVersionLowerError = versionCompareError{
		msg: func(e *versionCompareError) string {
			return fmt.Sprintf("Next semanticVersion (%s) is lower than current semanticVersion (%s)", e.nextVersion,
				e.currentVersion)
		},
	}
	nextVersionTooGreatError = versionCompareError{
		msg: func(e *versionCompareError) string {
			return fmt.Sprintf("Next semanticVersion (%s) is more than 2 minor versions greater than current semanticVersion (%s)",
				e.nextVersion, e.currentVersion)
		},
	}
)

type versionCompareError struct {
	msg            func(e *versionCompareError) string
	currentVersion string
	nextVersion    string
}

func (e *versionCompareError) Error() string {
	return e.msg(e)
}

func (e *versionCompareError) with(currentVersion, nextVersion string) *versionCompareError {
	e.currentVersion = currentVersion
	e.nextVersion = nextVersion
	return e
}

func checkCompatibility(current string, next string) error {
	curVersion, curVersionErr := semver.StrictNewVersion(current)
	nxtVersion, nxtVersionErr := semver.StrictNewVersion(next)

	if curVersionErr != nil {
		return currentVersionNoReleaseError.with(current, next)
	}
	if nxtVersionErr != nil {
		return nextVersionNoReleaseError.with(current, next)
	}
	if nxtVersion.LessThan(curVersion) {
		return nextVersionLowerError.with(current, next)
	}
	if nxtVersion.Major() > curVersion.Major() || (nxtVersion.Minor()-curVersion.Minor() > 1) { // only upgrade to the next minor version is guaranteed to work
		return nextVersionTooGreatError.with(current, next)
	}
	return nil
}

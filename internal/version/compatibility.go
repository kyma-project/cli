package version

import (
	"fmt"

	"github.com/blang/semver/v4"
)

var (
	currentVersionNoReleaseError = versionCompareError{
		msg: func(e *versionCompareError) string {
			return fmt.Sprintf("Current version seems not to be an official Kyma release: %s", e.currentVersion)
		},
	}
	nextVersionNoReleaseError = versionCompareError{
		msg: func(e *versionCompareError) string {
			return fmt.Sprintf("Next version seems not to be an official Kyma release: %s", e.nextVersion)
		},
	}
	nextVersionLowerError = versionCompareError{
		msg: func(e *versionCompareError) string {
			return fmt.Sprintf("Next version (%s) is lower than current version (%s)", e.nextVersion, e.currentVersion)
		},
	}
	nextVersionTooGreatError = versionCompareError{
		msg: func(e *versionCompareError) string {
			return fmt.Sprintf("Next version (%s) is more than 2 minor versions greater than current version (%s)",
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
	curVersion, curVersionErr := semver.Parse(current)
	nxtVersion, nxtVersionErr := semver.Parse(next)

	if curVersionErr != nil {
		return currentVersionNoReleaseError.with(current, next)
	}
	if nxtVersionErr != nil {
		return nextVersionNoReleaseError.with(current, next)
	}
	if nxtVersion.LT(curVersion) {
		return nextVersionLowerError.with(current, next)
	}
	if nxtVersion.Major > curVersion.Major || (uint64(nxtVersion.Minor)-uint64(curVersion.Minor) > 2) {
		return nextVersionTooGreatError.with(current, next)
	}
	return nil
}

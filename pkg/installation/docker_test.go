package installation

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SplitDockerDomain(t *testing.T) {
	test1 := "localhost:5000/test/testImage:1"
	d1, r1 := splitDockerDomain(test1)
	require.Equal(t, d1, "localhost:5000")
	require.Equal(t, r1, "test/testImage:1")

	test2 := "eu.gcr.io/test/testImage"
	d2, r2 := splitDockerDomain(test2)
	require.Equal(t, d2, "eu.gcr.io")
	require.Equal(t, r2, "test/testImage")

	test3 := "testImage"
	d3, r3 := splitDockerDomain(test3)
	require.Equal(t, d3, "index.docker.io")
	require.Equal(t, r3, "testImage")
}

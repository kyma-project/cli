//package source_test
//
//import (
//	"testing"
//
//	"github.com/open-component-model/ocm/pkg/contexts/ocm"
//	"github.com/stretchr/testify/assert"
//
//	"github.com/kyma-project/cli/pkg/module/source"
//)
//
//func TestGitSource_FetchSource(t *testing.T) {
//	t.Run("Success", func(t *testing.T) {
//		// given
//		ctx := internal.NewContextBuilder().Build()
//
//		path := "foo"
//		repo := "bar"
//		version := "baz"
//
//		expectedOcmD := &ocm.SourceMeta{}
//
//		gitSource := source.NewGitSource()
//
//		gitSource.On("FetchSource", ctx, path, repo, version).Return(expectedSource, nil)
//
//		// when
//		result, err := gitSource.FetchSource(ctx, path, repo, version)
//
//		// then
//		assert.NoError(t, err)
//		assert.Equal(t, expectedSource, result)
//
//	})
//}

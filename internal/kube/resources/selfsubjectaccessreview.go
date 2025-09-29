package resources

import (
	"context"

	"github.com/kyma-project/cli.v3/internal/kube"
	authv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CreateSelfSubjectAccessReview creates a SelfSubjectAccessReview CR to check if the current user can perform the given action.
func CreateSelfSubjectAccessReview(ctx context.Context, client kube.Client, verb, resource, namespace, group string) (*authv1.SelfSubjectAccessReview, error) {
	ssar := buildSelfSubjectAccessReview(verb, resource, namespace, group)
	return client.Static().AuthorizationV1().SelfSubjectAccessReviews().Create(ctx, ssar, metav1.CreateOptions{})
}

func buildSelfSubjectAccessReview(verb, resource, namespace, group string) *authv1.SelfSubjectAccessReview {
	return &authv1.SelfSubjectAccessReview{
		Spec: authv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authv1.ResourceAttributes{
				Verb:      verb,
				Resource:  resource,
				Namespace: namespace,
				Group:     group,
			},
		},
	}
}

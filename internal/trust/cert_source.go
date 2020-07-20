package trust

const (
	CertSourceConfigMap = "configmap"
	CertSourceSecret    = "secret"
)

type Source struct {
	name      string
	namespace string
	resource  string
}

func NewSource(name, namespace, resource string) Source {
	return Source{
		name:      name,
		namespace: namespace,
		resource:  resource,
	}
}

package values

type Sources struct {
	Domain     string
	Values     []string
	ValueFiles []string
	TLSCrtFile string
	TLSKeyFile string
	K3d        bool
}

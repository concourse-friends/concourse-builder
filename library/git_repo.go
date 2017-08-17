package library

type GitRepo struct {
	// URI to the git repo
	URI string `yaml:",omitempty"`

	// Private key the allows access to the repo
	PrivateKey string `yaml:"private_key,omitempty"`
}

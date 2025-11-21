package version_control

type VersionControl interface {
	Init(repoPath string, author string) error
	Commit(message string, author string, files []string) error
	Status() (string, error)
}

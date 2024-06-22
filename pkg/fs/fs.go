package fs

type FS interface {
	ReadFile(name string) ([]byte, error)
}

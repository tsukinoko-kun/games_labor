package scenarios

import "embed"

//go:embed *.txt
var fs embed.FS

func FromID(id string) (string, error) {
	data, err := fs.ReadFile(id + ".txt")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

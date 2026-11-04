package listcmd

import (
	"encoding/json"
	"os"
)

func loadMeta(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var m struct {
		Description string `json:"description"`
	}
	if err := json.Unmarshal(data, &m); err != nil {
		return "", err
	}
	return m.Description, nil
}

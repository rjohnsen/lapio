package kernel

import (
	"errors"
	"io"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Settings struct {
	Host     string
	Username string
	Password string
}

func LoadSettings(settingsPath string) (Settings, error) {
	var settings Settings
	settingsFile, err := os.Open(settingsPath)

	if err != nil {
		return settings, errors.New("Unable to open settings file")
	}

	defer settingsFile.Close()

	bytesContent, err := io.ReadAll(settingsFile)

	if err != nil {
		return settings, errors.New("Unable to read settings file")
	}

	err = toml.Unmarshal(bytesContent, &settings)

	if err != nil {
		return settings, errors.New("Unable to parase TOML settings file")
	}

	return settings, nil
}

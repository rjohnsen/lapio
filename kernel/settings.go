package kernel

import (
	"errors"
	"io"
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Settings struct {
	Host          string
	Username      string
	Password      string
	Workers       int
	Flushbytes    float64
	Flushinterval int
}

func LoadSettings(settingsPath string) (Settings, error) {
	var settings Settings
	settingsFile, err := os.Open(settingsPath)

	if err != nil {
		return settings, errors.New("unable to open settings file")
	}

	defer settingsFile.Close()

	bytesContent, err := io.ReadAll(settingsFile)

	if err != nil {
		return settings, errors.New("unable to read settings file")
	}

	err = toml.Unmarshal(bytesContent, &settings)

	if err != nil {
		return settings, errors.New("unable to parse TOML settings file")
	}

	return settings, nil
}

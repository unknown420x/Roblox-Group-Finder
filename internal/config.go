package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const (
	configDirName  = "GroupFinder"
	configFileName = "config.json"
)

func DefaultConfig() Config {
	return Config{Workers: 4, RPS: 1, BatchSize: 50, MinID: DefaultMinID, MaxID: DefaultMaxID, Timeout: "5s", Unique: true}
}

func ConfigPath() (string, error) {
	var base string
	switch runtime.GOOS {
	case "windows":
		base = os.Getenv("APPDATA")
		if base == "" {
			return "", errors.New("APPDATA is not available")
		}
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, "Library", "Application Support")
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = os.Getenv("XDG_CONFIG_HOME")
		if base == "" {
			base = filepath.Join(home, ".config")
		}
	}
	return filepath.Join(base, configDirName, configFileName), nil
}

func LoadConfig() (Config, error) {
	path, err := ConfigPath()
	if err != nil {
		return Config{}, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return DefaultConfig(), nil
	}
	if err != nil {
		return Config{}, fmt.Errorf("read config: %w", err)
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config: %w", err)
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func SaveConfig(cfg Config) error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	tmp, err := os.CreateTemp(filepath.Dir(path), ".config-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if err := tmp.Chmod(0o600); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return fmt.Errorf("replace config: %w", err)
	}
	return nil
}

func ResetConfig() error {
	path, err := ConfigPath()
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (c Config) TimeoutDuration() (time.Duration, error) { return time.ParseDuration(c.Timeout) }

func (c Config) Validate() error {
	if c.Workers < 1 || c.Workers > 256 {
		return errors.New("workers must be between 1 and 256")
	}
	if c.RPS < 1 {
		return errors.New("requests-per-second must be at least 1")
	}
	if c.BatchSize < 1 || c.BatchSize > 1000 {
		return errors.New("batch-size must be between 1 and 1000")
	}
	if c.MinID < 1 {
		return errors.New("min-group-id must be at least 1")
	}
	if c.MaxID < c.MinID {
		return errors.New("max-group-id must be greater than or equal to min-group-id")
	}
	if _, err := c.TimeoutDuration(); err != nil {
		return fmt.Errorf("invalid timeout: %w", err)
	}
	return nil
}

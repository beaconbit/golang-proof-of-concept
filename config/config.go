package config

import (
    "os"
    "gopkg.in/yaml.v3"
)

type Config struct {
    ScanInterval           int `yaml:"scan_interval"`
    LoginTimeout           int `yaml:"login_timeout"`
    CookieTTLSeconds       int `yaml:"cookie_ttl_seconds"`
    MaxDeviceFailures      int `yaml:"max_device_failures"`
    DeviceCheckInterval    int `yaml:"device_check_interval"`
    InvalidCheckEveryNCycles int `yaml:"invalid_check_every_n_cycles"`

    Credentials []struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
    } `yaml:"credentials"`

    Process struct {
	Mode      string `yaml:"mode"`
	Interface string `yaml:"interface"`
	DBType     string `yaml:"db_type"`
	DBDriver string `yaml:"db_driver"`
	DBMigrationUrl     string `yaml:"db_migration_url"`
	DBPath    string `yaml:"db_path"`
    } `yaml:"process"`
}

func LoadConfigWithDefault() (*Config, error) {
    return LoadConfig("")
}

func LoadConfig(path string) (*Config, error) {
    if path == "" {
	path = "config/config.yml"
    }
    f, err := os.Open(path)
    if err != nil {
	return nil, err
    }
    defer f.Close()

    var cfg Config
    decoder := yaml.NewDecoder(f)
    if err := decoder.Decode(&cfg); err != nil {
	return nil, err
    }
    return &cfg, nil
}


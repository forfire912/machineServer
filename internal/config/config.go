package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Server     ServerConfig     `mapstructure:"server"`
	Log        LogConfig        `mapstructure:"log"`
	Database   DatabaseConfig   `mapstructure:"database"`
	Redis      RedisConfig      `mapstructure:"redis"`
	Auth       AuthConfig       `mapstructure:"auth"`
	Backends   BackendsConfig   `mapstructure:"backends"`
	Resources  ResourcesConfig  `mapstructure:"resources"`
	Monitoring MonitoringConfig `mapstructure:"monitoring"`
	Storage    StorageConfig    `mapstructure:"storage"`
}

type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	File       string `mapstructure:"file"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`
	DSN    string `mapstructure:"dsn"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type AuthConfig struct {
	Enabled     bool          `mapstructure:"enabled"`
	JWTSecret   string        `mapstructure:"jwt_secret"`
	TokenExpire time.Duration `mapstructure:"token_expire"`
	APIKeys     []string      `mapstructure:"api_keys"`
}

type BackendsConfig struct {
	QEMU   BackendConfig `mapstructure:"qemu"`
	Renode BackendConfig `mapstructure:"renode"`
}

type BackendConfig struct {
	Enabled        bool   `mapstructure:"enabled"`
	Binary         string `mapstructure:"binary"`
	DefaultMachine string `mapstructure:"default_machine"`
}

type ResourcesConfig struct {
	MaxSessions    int `mapstructure:"max_sessions"`
	MaxMemoryMB    int `mapstructure:"max_memory_mb"`
	MaxCPUPercent  int `mapstructure:"max_cpu_percent"`
	SessionTimeout int `mapstructure:"session_timeout"`
}

type MonitoringConfig struct {
	Enabled        bool `mapstructure:"enabled"`
	PrometheusPort int  `mapstructure:"prometheus_port"`
}

type StorageConfig struct {
	BasePath           string `mapstructure:"base_path"`
	MaxProgramSizeMB   int    `mapstructure:"max_program_size_mb"`
	MaxSnapshotSizeMB  int    `mapstructure:"max_snapshot_size_mb"`
}

func Load(configPath string) (*Config, error) {
	viper.SetConfigFile(configPath)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

package config

import "fmt"

type Config struct {
	Database DBConfig  `yaml:"db" json:"db"`
	Command  CmdConfig `yaml:"cmd" json:"cmd"`
}

type CmdConfig struct {
	MigrationDir string `yaml:"migration_dir" json:"migration_dir"`
}

func NewCmdConfig(migrationDir string) CmdConfig {
	dirPath := migrationDir
	if dirPath == "" {
		dirPath = "./db/migrations"
	}
	return CmdConfig{
		MigrationDir: dirPath,
	}
}

type DBConfig struct {
	// currently support only "postgres"
	Dialect  string `yaml:"dialect" json:"dialect"`
	Host     string `yaml:"host" json:"host"`
	Port     int    `yaml:"port" json:"port"`
	Username string `yaml:"username" json:"username"`
	Password string `yaml:"password" json:"password"`
	Database string `yaml:"database" json:"database"`
	SSLMode  string `yaml:"sslmode" json:"sslmode"`
}

func (c *DBConfig) DSN() string {
	if c.Dialect == "postgres" {
		return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s", c.Username, c.Password, c.Host, c.Port, c.Database, c.SSLMode)
	}
	return ""
}

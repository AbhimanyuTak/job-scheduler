package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Database  DatabaseConfig  `mapstructure:"database"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Server    ServerConfig    `mapstructure:"server"`
	Scheduler SchedulerConfig `mapstructure:"scheduler"`
	Worker    WorkerConfig    `mapstructure:"worker"`
	Logging   LoggingConfig   `mapstructure:"logging"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Host string `mapstructure:"host"`
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // gin mode: debug, release, test
}

// SchedulerConfig holds scheduler configuration
type SchedulerConfig struct {
	PollInterval time.Duration `mapstructure:"poll_interval"`
	BatchSize    int           `mapstructure:"batch_size"`
	HTTPTimeout  time.Duration `mapstructure:"http_timeout"`
}

// WorkerConfig holds worker configuration
type WorkerConfig struct {
	PoolSize    int           `mapstructure:"pool_size"`
	HTTPTimeout time.Duration `mapstructure:"http_timeout"`
	RetryDelay  time.Duration `mapstructure:"retry_delay"`
	MaxRetries  int           `mapstructure:"max_retries"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `mapstructure:"level"`  // debug, info, warn, error
	Format string `mapstructure:"format"` // json, text
}

// LoadConfig loads configuration from file and environment variables
func LoadConfig(configPath string) (*Config, error) {
	// Set default values
	setDefaults()

	// Set config file path if provided
	if configPath != "" {
		viper.SetConfigFile(configPath)
	} else {
		// Look for config files in current directory
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		viper.AddConfigPath(".")
		viper.AddConfigPath("./config")
		viper.AddConfigPath("/etc/job-scheduler")
	}

	// Enable reading from environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("JOB_SCHEDULER")

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("error reading config file: %w", err)
		}
		// Config file not found is OK, we'll use defaults and env vars
	}

	// Unmarshal into Config struct
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// setDefaults sets default configuration values
func setDefaults() {
	// Database defaults
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.password", "password")
	viper.SetDefault("database.dbname", "job_scheduler")
	viper.SetDefault("database.sslmode", "disable")

	// Redis defaults
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", "6379")
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)

	// Server defaults
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")

	// Scheduler defaults
	viper.SetDefault("scheduler.poll_interval", "5s")
	viper.SetDefault("scheduler.batch_size", 100)
	viper.SetDefault("scheduler.http_timeout", "30s")

	// Worker defaults
	viper.SetDefault("worker.pool_size", 10)
	viper.SetDefault("worker.http_timeout", "90s")
	viper.SetDefault("worker.retry_delay", "10s")
	viper.SetDefault("worker.max_retries", 3)

	// Logging defaults
	viper.SetDefault("logging.level", "info")
	viper.SetDefault("logging.format", "json")
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Database.Port == "" {
		return fmt.Errorf("database port is required")
	}
	if c.Database.User == "" {
		return fmt.Errorf("database user is required")
	}
	if c.Database.DBName == "" {
		return fmt.Errorf("database name is required")
	}
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	if c.Redis.Port == "" {
		return fmt.Errorf("redis port is required")
	}
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}
	if c.Scheduler.PollInterval <= 0 {
		return fmt.Errorf("scheduler poll interval must be positive")
	}
	if c.Worker.PoolSize <= 0 {
		return fmt.Errorf("worker pool size must be positive")
	}
	return nil
}

// GetDSN returns the database connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// GetRedisAddr returns the Redis connection address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

// GetServerAddr returns the server address
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

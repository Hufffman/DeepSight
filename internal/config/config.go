package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/viper"
)

const configEnvKey = "DEEPSIGHT_CONFIG"

type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Database DatabaseConfig `mapstructure:"database"`
	Server   ServerConfig   `mapstructure:"server"`
	Redis    RedisConfig    `mapstructure:"redis"`
	RustFS   RustFSConfig   `mapstructure:"rustfs"`
	RabbitMQ RabbitMQConfig `mapstructure:"rabbitmq"`
	OpenAI   OpenAIConfig   `mapstructure:"openai"`
	Tavily   TavilyConfig   `mapstructure:"tavily"`
	JWT      JWTConfig      `mapstructure:"jwt"`
}

type AppConfig struct {
	Name string `mapstructure:"name"`
	Env  string `mapstructure:"env"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type RustFSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	Region          string `mapstructure:"region"`
	AccessKeyID     string `mapstructure:"accessKeyId"`
	SecretAccessKey string `mapstructure:"secretAccessKey"`
	Bucket          string `mapstructure:"bucket"`
	UseSSL          bool   `mapstructure:"useSSL"`
	UsePathStyle    bool   `mapstructure:"usePathStyle"`
}

type RabbitMQConfig struct {
	Host       string `mapstructure:"host"`
	Port       int    `mapstructure:"port"`
	User       string `mapstructure:"user"`
	Password   string `mapstructure:"password"`
	Exchange   string `mapstructure:"exchange"`
	BindingKey string `mapstructure:"bindingKey"`
	Queue      string `mapstructure:"queue"`
}

type OpenAIConfig struct {
	BaseUrl            string `mapstructure:"baseUrl"`
	ApiKey             string `mapstructure:"apiKey"`
	EmbeddingModel     string `mapstructure:"embeddingModel"`
	ChatModel          string `mapstructure:"chatModel"`
	EmbeddingBatchSize int    `mapstructure:"embeddingBatchSize"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
	Expire string `mapstructure:"expire"`
}

type TavilyConfig struct {
	APIKey string `mapstructure:"apikey"`
}

func (c *JWTConfig) ExpireDuration() time.Duration {
	duration, err := time.ParseDuration(c.Expire)
	if err != nil {
		return 24 * time.Hour
	}
	return duration
}

func Load(path string) (*Config, error) {
	resolvedPath, err := resolveConfigPath(path)
	if err != nil {
		return nil, err
	}

	v := viper.New()
	v.SetConfigFile(resolvedPath)

	ext := strings.TrimPrefix(filepath.Ext(resolvedPath), ".")
	if ext != "" {
		v.SetConfigType(ext)
	}

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

func resolveConfigPath(path string) (string, error) {
	candidates := make([]string, 0, 3)

	if envPath := strings.TrimSpace(os.Getenv(configEnvKey)); envPath != "" {
		candidates = append(candidates, envPath)
	}
	if strings.TrimSpace(path) != "" {
		candidates = append(candidates, path)
		if !filepath.IsAbs(path) {
			if executablePath, err := os.Executable(); err == nil {
				candidates = append(candidates, filepath.Join(filepath.Dir(executablePath), path))
			}
		}
	}

	checked := make([]string, 0, len(candidates))
	seen := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		checked = append(checked, candidate)

		info, err := os.Stat(candidate)
		if err == nil && !info.IsDir() {
			return candidate, nil
		}
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("failed to access config %q: %w", candidate, err)
		}
	}

	return "", fmt.Errorf("config file not found; checked: %s; override with %s", strings.Join(checked, ", "), configEnvKey)
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

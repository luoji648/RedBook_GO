package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	MySQL    MySQLConfig    `mapstructure:"mysql"`
	Redis    RedisConfig    `mapstructure:"redis"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Dify     DifyConfig     `mapstructure:"dify"`
	OSS      OSSConfig      `mapstructure:"oss"`
	RocketMQ RocketMQConfig `mapstructure:"rocketmq"`
}

type ServerConfig struct {
	Port int `mapstructure:"port"`
}

type MySQLConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	Database string `mapstructure:"database"`
}

type RedisConfig struct {
	Addr     string `mapstructure:"addr"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type JWTConfig struct {
	Secret        string `mapstructure:"secret"`
	ExpireSeconds int64  `mapstructure:"expire_seconds"`
}

type DifyConfig struct {
	URL              string `mapstructure:"url"`
	NoteAssistantKey string `mapstructure:"note_assistant_key"`
}

type OSSConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	Bucket          string `mapstructure:"bucket"`
	URLPrefix       string `mapstructure:"url_prefix"`
}

type RocketMQConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	NameServer        string `mapstructure:"name_server"`
	ProducerGroup     string `mapstructure:"producer_group"`
	ConsumerGroup     string `mapstructure:"consumer_group"`
	ClickTopic        string `mapstructure:"click_topic"`
	ClickRecordTopic  string `mapstructure:"click_record_topic"`
	ClickSummaryTopic string `mapstructure:"click_summary_topic"`
}

func setDefaults() {
	viper.SetDefault("server.port", 8081)
	viper.SetDefault("mysql.host", "127.0.0.1")
	viper.SetDefault("mysql.port", 3306)
	viper.SetDefault("mysql.user", "root")
	viper.SetDefault("mysql.password", "123456")
	viper.SetDefault("mysql.database", "redbook")
	viper.SetDefault("redis.addr", "192.168.220.128:6379")
	viper.SetDefault("redis.password", "123321")
	viper.SetDefault("redis.db", 1)
	viper.SetDefault("jwt.secret", "redbook-jwt-secret-key-change-me-in-production-min-32-chars!!")
	viper.SetDefault("jwt.expire_seconds", 36000)
	viper.SetDefault("dify.url", "http://localhost/v1/chat-messages")
	viper.SetDefault("dify.note_assistant_key", "app-placeholder")
	viper.SetDefault("oss.endpoint", "oss-cn-beijing.aliyuncs.com")
	viper.SetDefault("rocketmq.enabled", false)
	viper.SetDefault("rocketmq.name_server", "192.168.220.128:9876")
	viper.SetDefault("rocketmq.producer_group", "redbook-producer-group")
	viper.SetDefault("rocketmq.consumer_group", "redbook-consumer-group")
	viper.SetDefault("rocketmq.click_topic", "click-topic")
	viper.SetDefault("rocketmq.click_record_topic", "click-record-topic")
	viper.SetDefault("rocketmq.click_summary_topic", "click-summary-topic")
}

func Load() (*Config, error) {
	setDefaults()

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("read config: %w", err)
		}
	}

	applyEnv()

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}
	if len(cfg.JWT.Secret) < 32 {
		return nil, fmt.Errorf("jwt.secret must be at least 32 bytes")
	}
	if cfg.Server.Port <= 0 || cfg.Server.Port > 65535 {
		return nil, fmt.Errorf("server.port must be between 1 and 65535")
	}
	if cfg.MySQL.Host == "" || cfg.MySQL.User == "" || cfg.MySQL.Database == "" {
		return nil, fmt.Errorf("mysql.host, mysql.user, and mysql.database are required")
	}
	if cfg.Redis.Addr == "" {
		return nil, fmt.Errorf("redis.addr is required")
	}
	return &cfg, nil
}

func applyEnv() {
	envOverrides := map[string]string{
		"MYSQL_HOST":                   "mysql.host",
		"MYSQL_PORT":                   "mysql.port",
		"MYSQL_USER":                   "mysql.user",
		"MYSQL_PASSWORD":               "mysql.password",
		"MYSQL_DATABASE":               "mysql.database",
		"REDIS_ADDR":                   "redis.addr",
		"REDIS_PASSWORD":               "redis.password",
		"REDIS_DB":                     "redis.db",
		"JWT_SECRET":                   "jwt.secret",
		"JWT_EXPIRE_SECONDS":           "jwt.expire_seconds",
		"DIFY_NOTE_KEY":                "dify.note_assistant_key",
		"DIFY_URL":                     "dify.url",
		"ALIYUN_OSS_ENDPOINT":          "oss.endpoint",
		"ALIYUN_OSS_ACCESS_KEY_ID":     "oss.access_key_id",
		"ALIYUN_OSS_ACCESS_KEY_SECRET": "oss.access_key_secret",
		"ALIYUN_OSS_BUCKET":            "oss.bucket",
		"ALIYUN_OSS_URL_PREFIX":        "oss.url_prefix",
		"ROCKETMQ_ENABLED":             "rocketmq.enabled",
		"ROCKETMQ_NAME_SERVER":         "rocketmq.name_server",
		"ROCKETMQ_PRODUCER_GROUP":      "rocketmq.producer_group",
		"ROCKETMQ_CONSUMER_GROUP":      "rocketmq.consumer_group",
		"ROCKETMQ_CLICK_TOPIC":         "rocketmq.click_topic",
		"ROCKETMQ_CLICK_RECORD_TOPIC":  "rocketmq.click_record_topic",
		"ROCKETMQ_CLICK_SUMMARY_TOPIC": "rocketmq.click_summary_topic",
	}
	for envKey, cfgKey := range envOverrides {
		if value := strings.TrimSpace(os.Getenv(envKey)); value != "" {
			switch envKey {
			case "MYSQL_PORT", "REDIS_DB", "JWT_EXPIRE_SECONDS":
				if n, err := strconv.Atoi(value); err == nil {
					viper.Set(cfgKey, n)
				}
			case "ROCKETMQ_ENABLED":
				if b, err := strconv.ParseBool(value); err == nil {
					viper.Set(cfgKey, b)
				}
			default:
				viper.Set(cfgKey, value)
			}
		}
	}
}

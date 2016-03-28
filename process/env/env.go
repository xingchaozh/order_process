package env

import (
	"os"

	"github.com/Sirupsen/logrus"
	"gopkg.in/gcfg.v1"
)

// The path of configuration files
const (
	REDIS_CFG_FILE = "config/database.gcfg"
	LOG_CFG_FILE   = "config/log.gcfg"
	ServiceName    = "order_process"
)

// The defination of redis configuration
type RedisCfg struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// The definition of log configuration
type LogCfg struct {
	Loglevel string `json:"log_level"`
}

// The definition of service environment
type Env struct {
	RedisConfig RedisCfg
	LogConfig   LogCfg
}

// The constuctor of environment
func New() *Env {
	redisConfig := RedisCfg{}
	return &Env{
		RedisConfig: redisConfig,
	}
}

// Initialize the environment
func (env *Env) InitEnv() *Env {

	// Load redis configuration from file
	type RedisCfgs struct {
		Env map[string]*RedisCfg
	}
	var redisCfgs RedisCfgs
	err := gcfg.ReadFileInto(&redisCfgs, REDIS_CFG_FILE)
	if err != nil {
		logrus.Error(err)
	}

	// Load log configuration from file
	type LogCfgs struct {
		Env map[string]*LogCfg
	}
	var logCfgs LogCfgs
	err = gcfg.ReadFileInto(&logCfgs, LOG_CFG_FILE)
	if err != nil {
		logrus.Error(err)
	}

	// Get the chapter of configurations
	orderProcessEnv := os.Getenv("ORDER_PROCESSING_SERVICE_ENV")
	if orderProcessEnv == "" {
		orderProcessEnv = "dev"
	}

	// Populate the environment
	env.RedisConfig = *redisCfgs.Env[orderProcessEnv]
	env.LogConfig = *logCfgs.Env[orderProcessEnv]

	level, err := logrus.ParseLevel(env.LogConfig.Loglevel)
	if err != nil {
		logrus.Error(err)
	}
	logrus.SetLevel(level)

	logrus.Debugf("Redis configuration loaded: %v", env.RedisConfig)
	logrus.Debugf("Log configuration loaded: %v", env.LogConfig)

	// If no local configuration found, we should qurey the Discovery Service.
	return env
}

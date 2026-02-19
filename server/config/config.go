package config

import (
	"database/sql"
	"log/slog"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/clerk/clerk-sdk-go/v2/user"

	"github.com/luketeo/horizon/config/provider"
)

type Config struct {
	clerk  *user.Client
	db     *sql.DB
	env    *provider.EnvProvider
	logger *slog.Logger
}

func NewConfig() *Config {
	config := Config{}

	config.env = provider.NewEnvProvider()
	config.logger = provider.NewLoggerProvider(config.env)
	config.db = provider.NewDBProvider(config.env)

	// Set clerk key for instantiation of app
	clerk.SetKey(config.env.ClerkSecretKey())

	return &config
}

func (c *Config) Clerk() *user.Client {
	if c.clerk == nil {
		c.clerk = provider.NewClerkProvider(c.env)
	}

	return c.clerk
}

func (c *Config) DB() *sql.DB {
	if c.db == nil {
		c.db = provider.NewDBProvider(c.env)
	}

	return c.db
}

func (c *Config) Env() *provider.EnvProvider {
	return c.env
}

func (c *Config) Logger() *slog.Logger {
	if c.logger == nil {
		c.logger = provider.NewLoggerProvider(c.env)
	}

	return c.logger
}

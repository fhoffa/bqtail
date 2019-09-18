package tail

import (
	"bqtail/base"
	"bqtail/tail/config"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/viant/afs"
	"github.com/viant/afs/matcher"
	"github.com/viant/afs/storage"
	"os"
	"strings"
)

//Config represents a tail config
type Config struct {
	base.Config
	Routes   config.Routes
	BatchURL string
}

func (c *Config) loadRoutes(ctx context.Context) error {
	if c.RoutesBaseURL == "" {
		return nil
	}
	fs := afs.New()
	routesObject, err := fs.List(ctx, c.RoutesBaseURL, matcher.NewBasic("", ".json", ""))
	if err != nil {
		return err
	}
	for _, object := range routesObject {
		if err = c.loadRoute(ctx, fs, object);err != nil {
			return err
		}
	}
	return nil
}

func (c *Config) loadRoute(ctx context.Context, storage afs.Service, object storage.Object) error {
	reader, err := storage.Download(ctx, object)
	if err != nil {
		return err
	}
	defer func() {
		_ = reader.Close()
	}()
	routes := config.Routes{}
	if err = json.NewDecoder(reader).Decode(&routes);err == nil {
		c.Routes = append(c.Routes, routes...)
	}
	return err
}


//Init initializes config
func (c *Config) Init(ctx context.Context) error {
	err := c.Config.Init(ctx)
	if err != nil {
		return err
	}
	if len(c.Routes) == 0 {
		c.Routes  = config.Routes{}
	}

	if err := c.loadRoutes(ctx);err != nil {
		return err
	}
	for _, route := range c.Routes {
		if route.Actions.Async && route.Actions.DeferTaskURL == "" {
			route.Actions.DeferTaskURL = c.DeferTaskURL
		}
		if route.Batch != nil {
			route.Batch.Init()
		}
	}
	return nil
}

//Validate checks if config is valid
func (c *Config) Validate() error {
	err := c.Config.Validate()
	if err != nil {
		return err
	}
	if len(c.Routes) == 0 {
		return fmt.Errorf("routes were empty")
	}
	if c.Routes.UsesBatch() && c.BatchURL == "" {
		return fmt.Errorf("batchURL were empty")
	}
	return c.Routes.Validate()
}

//NewConfigFromEnv creates config from env
func NewConfigFromEnv(ctx context.Context, key string) (*Config, error) {
	if key == "" {
		return nil, errors.New("os env cfg key was empty")
	}
	data := os.Getenv(key)
	if data == "" {
		return nil, fmt.Errorf("env.%v was empty", key)
	}
	cfg := &Config{}
	err := json.NewDecoder(strings.NewReader(data)).Decode(cfg)
	if err == nil {
		if err = cfg.Init(ctx); err != nil {
			return nil, err
		}
		err = cfg.Validate()
	}
	return cfg, err
}

//NewConfigFromURL creates new config from URL
func NewConfigFromURL(ctx context.Context, URL string) (*Config, error) {
	storageService := afs.New()
	reader, err := storageService.DownloadWithURL(ctx, URL)
	if err != nil {
		return nil, err
	}
	cfg := &Config{}
	err = json.NewDecoder(reader).Decode(cfg)
	if err == nil {
		if err = cfg.Init(ctx); err != nil {
			return cfg, err
		}
		err = cfg.Validate()
	}
	return cfg, err
}

//NewConfig creates a new config from env (json or URL)
func NewConfig(ctx context.Context, key string) (*Config, error) {
	if key == "" {
		return nil, fmt.Errorf("config key was empty")
	}
	value := os.Getenv(key)
	if json.Valid([]byte(value)) {
		return NewConfigFromEnv(ctx, key)
	}
	return NewConfigFromURL(ctx, value)
}

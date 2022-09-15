package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/gou/helper"
	"github.com/yaoapp/kun/any"
)

// Connector connector
type Connector struct {
	Name    string
	rdb     *redis.Client
	Options Options `json:"options"`
}

// Options the redis connector option
type Options struct {
	Host    string `json:"host,omitempty"`
	Port    string `json:"port,omitempty"`
	User    string `json:"user,omitempty"`
	Pass    string `json:"pass,omitempty"`
	Timeout int    `json:"timeout,omitempty"`
	DB      string `json:"db,omitempty"`
}

// Register the connections from dsl
func (r *Connector) Register(name string, dsl []byte) error {
	err := jsoniter.Unmarshal(dsl, r)
	if err != nil {
		return err
	}

	err = r.setDefaults()
	if err != nil {
		return err
	}

	r.Name = name
	return r.makeConnection()
}

// Is the connections from dsl
func (r *Connector) Is(typ int) bool {
	return 2 == typ
}

func (r *Connector) setDefaults() error {
	r.Options.Host = helper.EnvString(r.Options.Host)
	r.Options.Pass = helper.EnvString(r.Options.Pass)
	r.Options.User = helper.EnvString(r.Options.User)
	r.Options.Port = helper.EnvString(r.Options.Port)
	r.Options.DB = helper.EnvString(r.Options.DB)
	r.Options.Timeout = helper.EnvInt(r.Options.Timeout, 5)
	if r.Options.Timeout == 0 {
		r.Options.Timeout = 5
	}

	if r.Options.Port == "" {
		r.Options.Port = "6379"
	}

	if r.Options.DB == "" {
		r.Options.DB = "0"
	}

	return nil
}

func (r *Connector) makeConnection() error {
	if r.Options.Host == "" {
		return fmt.Errorf("options.host is required")
	}

	options := &redis.Options{
		Addr: fmt.Sprintf("%s:%s", r.Options.Host, r.Options.Port),
		DB:   any.Of(r.Options.DB).CInt(),
	}

	if r.Options.User != "" {
		options.Username = r.Options.User
	}

	if r.Options.Pass != "" {
		options.Password = r.Options.Pass
	}

	client := redis.NewClient(options).WithTimeout(time.Duration(r.Options.Timeout) * time.Second)
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return err
	}

	r.rdb = client
	return nil
}
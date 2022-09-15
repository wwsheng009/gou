package connector

import (
	"fmt"

	jsoniter "github.com/json-iterator/go"
	"github.com/yaoapp/gou/connector/database"
	"github.com/yaoapp/gou/connector/redis"
)

// Connectors the loaded connectors
var Connectors = map[string]Connector{}

// Load a connector from source
func Load(source string, name string) (Connector, error) {

	dsl := DSL{}
	bytes := []byte(source)
	err := jsoniter.Unmarshal(bytes, &dsl)
	if err != nil {
		return nil, err
	}

	c, err := make(dsl.Type)
	if err != nil {
		return nil, err
	}

	err = c.Register(name, bytes)
	if err != nil {
		return nil, err
	}

	Connectors[name] = c
	return Connectors[name], nil
}

// Select a connector
func Select(name string) (Connector, error) {
	return nil, nil
}

func make(typ string) (Connector, error) {

	t, has := types[typ]
	if !has {
		return nil, fmt.Errorf("%s does not support", typ)
	}

	switch t {
	case DATABASE:
		c := &database.Xun{}
		return c, nil

	case REDIS:
		c := &redis.Connector{}
		return c, nil
	}

	return nil, fmt.Errorf("%s does not support yet", typ)
}
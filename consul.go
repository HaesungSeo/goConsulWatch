package ConsulWatch

import (
	"encoding/json"
	"errors"
	"sync"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
)

var (
	ErrWatch = errors.New("unsupported watch type")
)

type ConsulWatch struct {
	kvMap map[string]string
	mutex *sync.Mutex
	plan  *watch.Plan
}

func (c *ConsulWatch) Stop(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	// flush cache first, then stop it
	c.kvMap = make(map[string]string)
	c.plan.Stop()
}

func (c *ConsulWatch) KV(key string) string {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return c.kvMap[key]
}

func (c *ConsulWatch) KVSet(key, value string) {
	c.mutex.Lock()
	c.kvMap[key] = value
	defer c.mutex.Unlock()
}

func (c *ConsulWatch) KVFlush() {
	c.mutex.Lock()
	// just assing fresh new one
	c.kvMap = make(map[string]string)
	defer c.mutex.Unlock()
}

func (c *ConsulWatch) KVCopy() map[string]string {
	cpy := make(map[string]string)
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k, v := range c.kvMap {
		cpy[k] = v
	}
	return cpy
}

// See TestKeyWatch @ github.com/hashicorp/consul/api@v1.13.0/watch/funcs_test.go
func mustParse(q string) (*watch.Plan, error) {
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(q), &params); err != nil {
		return nil, err
	}
	plan, err := watch.Parse(params)
	if err != nil {
		return nil, err
	}
	return plan, nil
}

type ParseError struct {
	Query string
	Err   error
}

func (e *ParseError) Error() string {
	return e.Err.Error() + ": Qeury" + e.Query
}

func (e *ParseError) Unwrap() error { return e.Err }

type KeytypeError struct {
	keytype string
	Err     error
}

func (e *KeytypeError) Error() string {
	return e.Err.Error() + ": " + e.keytype
}

func (e *KeytypeError) Unwrap() error { return e.Err }

// Watch example
func New(addr, keyType, key string) (*ConsulWatch, error) {
	cfgMap := &ConsulWatch{
		kvMap: make(map[string]string),
		mutex: &sync.Mutex{},
	}
	query := ""

	switch keyType {
	case "key":
		query = `{"type":"` + keyType + `", "key":"` + key + `"}`
	case "keyprefix":
		query = `{"type":"` + keyType + `", "prefix":"` + key + `"}`
	default:
		return nil, &KeytypeError{keytype: keyType, Err: ErrWatch}
	}
	plan, err := mustParse(query)
	if err != nil {
		return nil, &ParseError{Query: query, Err: err}
	}
	plan.Handler = func(idx uint64, raw interface{}) {
		if raw == nil { // nil is a valid return value
			// fmt.Printf("key=[%s] not found\n", key)
			// flush all configurations
			cfgMap.KVFlush()
			return
		}

		switch raw.(type) {
		case *api.KVPair:
			v, _ := raw.(*api.KVPair)
			cfgMap.KVSet(v.Key, string(v.Value))
		case api.KVPairs:
			pairs, _ := raw.(api.KVPairs)
			for _, v := range pairs {
				cfgMap.KVSet(v.Key, string(v.Value))
			}
		}
	}
	cfgMap.plan = plan

	go func() {
		plan.Run(addr)
	}()

	return cfgMap, nil
}

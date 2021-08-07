package jotto

// Configuration is an interface that application's custom settings struct must conform to.
type Configuration interface {
	Motto() *Settings
	Load() error
}

type Settings struct {
	Protocol string `json:"protocol" xml:"Protocol"`
	Address  string `json:"address" xml:"Address"`

	Cache []*CacheSettings `json:"cache,omitempty" xml:"Cache>Instance,omitempty"`
	Queue []*QueueSettings `json:"queue,omitempty" xml:"Queue>Instance,omitempty"`
}

type CacheSettings struct {
	Name      string             `json:"name" xml:"Name"`
	Driver    string             `json:"driver" xml:"Driver"`
	Redis     *RedisSettings     `json:"redis,omitempty" xml:"Redis,omitempty"`
	Memcached *MemcachedSettings `json:"memcached,omitempty" xml:"Memcached,omitempty"`
}

type QueueSettings struct {
	Name   string         `json:"name" xml:"Name"`
	Driver string         `json:"driver" xml:"Driver"`
	Queues []string       `json:"queues" xml:"Queues>Name,omitempty"`
	Redis  *RedisSettings `json:"redis,omitempty" xml:"Redis,omitempty"`
}

type RedisSettings struct {
	Address      string `json:"address" xml:"Address"`
	Database     int    `json:"database" xml:"database"`
	Password     string `json:"password,omitempty" xml:"Password,omitempty"`
	DialTimeout  int    `json:"dial-timeout,omitempty" xml:"DialTimeout,omitempty"`
	ReadTimeout  int    `json:"read-timeout,omitempty" xml:"ReadTimeout,omitempty"`
	WriteTimeout int    `json:"write-timeout,omitempty" xml:"WriteTimeout,omitempty"`
}

type MemcachedSettings struct {
	Address []string `json:"address" xml:"Address"`
}

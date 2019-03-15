package conf

import (
	"flag"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/BurntSushi/toml"
)

const (
	defaultRedisAddress = "127.0.0.1:6379"
	defaultEtcdAddress  = "127.0.0.1:2379"
	defaultZKAddress    = "127.0.0.1:2181"
	defaultStep         = 1000

	defaultStoreMinSecond = 300
	defaultStoreMaxSecond = 1800
)

type Config struct {
	Server struct {
		Address string `toml:"addr"`
	} `toml:"server"`
	Store struct {
		Type      string `toml:"type"`
		URI       string `toml:"uri"`
		MinSecond int    `toml:"min_second"`
		MaxSecond int    `toml:"max_second"`
		Min       time.Duration
		Max       time.Duration
	} `toml:"store"`
	Generate struct {
		DataCenter uint8 `toml:"dataCenter"`
		Step       int64 `toml:"step"`
	} `toml:"generate"`
}

func Init() Config {
	var config Config

	if _, err := toml.DecodeFile("etc/rabbitid.toml", &config); err != nil {
		log.Fatalln("decode toml err", err.Error())
	}

	if config.Generate.Step == 0 {
		config.Generate.Step = defaultStep
	}
	var (
		httpAddr   = flag.String("http.addr", envString("ADDRESS", config.Server.Address), "HTTP listen address")
		dataCenter = flag.Uint64("dataCenter", envUint64("DATA_CENTER", uint64(config.Generate.DataCenter)), "DataCenter ID: {M5: 0, LG: 1, SJQ: 2}")
		step       = flag.Int64("step", envInt64("DATA_CENTER", config.Generate.Step), "Step")
		storeType  = flag.String("store", envString("STORE", config.Store.Type), "Store type：redis etcd zk")
		storeURI   = flag.String("store.uri", envString("URI", config.Store.URI), "Store URI")
	)
	flag.Parse()

	config.Server.Address = *httpAddr
	config.Store.Type = *storeType
	config.Generate.DataCenter = uint8(*dataCenter)
	config.Generate.Step = *step
	if config.Store.MaxSecond == 0 {
		config.Store.MaxSecond = defaultStoreMaxSecond
	}
	if config.Store.MinSecond == 0 {
		config.Store.MinSecond = defaultStoreMinSecond
	}
	switch *storeType {
	default:
		log.Fatalln("store type Error", *storeType)
	case "redis":
		config.Store.URI = *storeURI
		if config.Store.URI == "" {
			config.Store.URI = defaultRedisAddress
		}
	case "etcd":
		config.Store.URI = *storeURI
		if config.Store.URI == "" {
			config.Store.URI = defaultEtcdAddress
		}
	case "zk":
		config.Store.URI = *storeURI
		if config.Store.URI == "" {
			config.Store.URI = defaultZKAddress
		}
	}

	config.Store.Min = time.Duration(config.Store.MinSecond) * time.Second
	config.Store.Max = time.Duration(config.Store.MaxSecond) * time.Second
	return config
}

func envInt64(env string, fallback int64) int64 {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	i, err := strconv.ParseInt(e, 10, 64)
	if err != nil {
		return fallback
	}
	return i
}

func envUint64(env string, fallback uint64) uint64 {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	i, err := strconv.ParseUint(e, 10, 64)
	if err != nil {
		return fallback
	}
	return i
}

func envString(env, fallback string) string {
	e := os.Getenv(env)
	if e == "" {
		return fallback
	}
	return e
}

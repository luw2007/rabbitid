package conf

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/lestrrat/go-file-rotatelogs"
	"github.com/pkg/errors"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
)

const (
	AppName             = "rabbitid"
	defaultConfFile     = "etc/rabbitid.toml"
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
		Debug   bool   `toml:"debug"`
	} `toml:"server"`
	Log struct {
		Level string `toml:"level"`
		Path  string `toml:"path"`
	} `toml:"log"`
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
	Logger *logrus.Logger `toml:"-"`
}

func Init() Config {
	var (
		confFile = flag.String("c", defaultConfFile, "config file (default: etc/shrike.toml)")
	)

	// 系统目录
	confPath, _ := filepath.Abs(*confFile)
	APPPath := strings.SplitAfterN(confPath, "/"+AppName, 2)[0]
	ConfFile := filepath.Base(*confFile)
	file := filepath.Join(APPPath, "etc", ConfFile)

	var config Config
	flag.Parse()
	if _, err := toml.DecodeFile(file, &config); err != nil {
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

	var (
		logLevel = flag.String("log", envString("LOG_LEVEL", config.Log.Level), "log level")
	)
	flag.Parse()

	config.Log.Level = *logLevel

	// 设置日志级别
	baseLogPath := path.Join(config.Log.Path, AppName+".log")
	writer, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(baseLogPath),   // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),  // 文件最大保存时间
		rotatelogs.WithRotationTime(time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		logrus.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}
	baseLogPath = path.Join(config.Log.Path, AppName+"_error.log")
	errWriter, err := rotatelogs.New(
		baseLogPath+".%Y%m%d%H%M",
		rotatelogs.WithLinkName(baseLogPath),   // 生成软链，指向最新日志文件
		rotatelogs.WithMaxAge(7*24*time.Hour),  // 文件最大保存时间
		rotatelogs.WithRotationTime(time.Hour), // 日志切割时间间隔
	)
	if err != nil {
		logrus.Errorf("config local file system logger error. %v", errors.WithStack(err))
	}

	log := logrus.New()

	lvl, err := logrus.ParseLevel(config.Log.Level)
	if err != nil {
		log.WithError(err).Error("conf logLevel err")
	}
	log.SetLevel(lvl)
	if !config.Server.Debug {
		log.Out = ioutil.Discard
	}
	config.Logger = log

	customFormatter := new(logrus.JSONFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05.999999"
	log.Formatter = customFormatter
	// 为不同级别设置不同的输出目的
	log.AddHook(lfshook.NewHook(
		lfshook.WriterMap{
			logrus.DebugLevel: writer,
			logrus.InfoLevel:  writer,
			logrus.WarnLevel:  errWriter,
			logrus.ErrorLevel: errWriter,
			logrus.FatalLevel: errWriter,
			logrus.PanicLevel: errWriter,
		},
		customFormatter))
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

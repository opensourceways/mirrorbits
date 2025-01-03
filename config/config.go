// Copyright (c) 2014-2019 Ludovic Fauvet
// Licensed under the MIT license

package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"

	"github.com/op/go-logging"
	"github.com/opensourceways/mirrorbits/core"
	"gopkg.in/yaml.v2"
)

var (
	// TEMPLATES_PATH is set at compile time
	TEMPLATES_PATH = ""
)

var (
	log         = logging.MustGetLogger("main")
	config      *Configuration
	configMutex sync.RWMutex

	subscribers     []chan bool
	subscribersLock sync.RWMutex
)

func defaultConfig() Configuration {
	return Configuration{
		Repository:             "",
		Templates:              TEMPLATES_PATH,
		LocalJSPath:            "",
		OutputMode:             "auto",
		ListenAddress:          ":8080",
		Gzip:                   false,
		RedisAddress:           "127.0.0.1:6379",
		RedisPassword:          "",
		RedisDB:                0,
		LogDir:                 "",
		TraceFileLocation:      "",
		GeoipDatabasePath:      "/usr/share/GeoIP/",
		ConcurrentSync:         50,
		ScanInterval:           60,
		CheckInterval:          30,
		RepositoryScanInterval: 50,
		MaxLinkHeaders:         10,
		FixTimezoneOffsets:     false,
		Hashes: hashing{
			SHA1:   false,
			SHA256: true,
			MD5:    false,
		},
		DisallowRedirects:       false,
		WeightDistributionRange: 1.5,
		DisableOnMissingFile:    false,
		RPCListenAddress:        "localhost:3390",
		RPCPassword:             "",
		SchemaStrictMatch:       true,
	}
}

// Configuration contains all the option available in the yaml file
type Configuration struct {
	Repository                string     `yaml:"Repository"`
	RepositoryFileListText    string     `yaml:"RepositoryFileListText"`
	RepositorySourcesLockFile string     `yaml:"RepositorySourcesLockFile"`
	Templates                 string     `yaml:"Templates"`
	LocalJSPath               string     `yaml:"LocalJSPath"`
	OutputMode                string     `yaml:"OutputMode"`
	ListenAddress             string     `yaml:"ListenAddress"`
	Gzip                      bool       `yaml:"Gzip"`
	RedisAddress              string     `yaml:"RedisAddress"`
	RedisPassword             string     `yaml:"RedisPassword"`
	RedisDB                   int        `yaml:"RedisDB"`
	LogDir                    string     `yaml:"LogDir"`
	TraceFileLocation         string     `yaml:"TraceFileLocation"`
	GeoipDatabasePath         string     `yaml:"GeoipDatabasePath"`
	ConcurrentSync            int        `yaml:"ConcurrentSync"`
	ScanInterval              int        `yaml:"ScanInterval"`
	CheckInterval             int        `yaml:"CheckInterval"`
	RepositoryScanInterval    int        `yaml:"RepositoryScanInterval"`
	MaxLinkHeaders            int        `yaml:"MaxLinkHeaders"`
	FixTimezoneOffsets        bool       `yaml:"FixTimezoneOffsets"`
	Hashes                    hashing    `yaml:"Hashes"`
	DisallowRedirects         bool       `yaml:"DisallowRedirects"`
	WeightDistributionRange   float32    `yaml:"WeightDistributionRange"`
	DisableOnMissingFile      bool       `yaml:"DisableOnMissingFile"`
	Fallbacks                 []Fallback `yaml:"Fallbacks"`
	SchemaStrictMatch         bool       `yaml:"SchemaStrictMatch"`

	RedisSentinelMasterName string      `yaml:"RedisSentinelMasterName"`
	RedisSentinels          []sentinels `yaml:"RedisSentinels"`

	RPCListenAddress string `yaml:"RPCListenAddress"`
	RPCPassword      string `yaml:"RPCPassword"`

	PreReleaseVersion   string           `yaml:"PreReleaseVersion"`
	RepositoryFilter    DirFilter        `yaml:"RepositoryFilter"`
	RepoFileIntoVersion []FileVersionMap `yaml:"RepoFileIntoVersion"`
}

type ParticularFileMapping struct {
	VersionName  string   `yaml:"VersionName"`
	ScenarioName string   `yaml:"ScenarioName"`
	ArchName     string   `yaml:"ArchName"`
	SourcePath   []string `yaml:"SourcePath"`
	SHA256List   []string `yaml:"SHA256List"`
}

type DirFilter struct {
	SecondDir      []string                `yaml:"SecondDir"`
	ThirdDir       []string                `yaml:"ThirdDir"`
	ParticularFile []ParticularFileMapping `yaml:"ParticularFileMapping"`
}

type FileVersionMap struct {
	FilePath string `yaml:"FilePath"`
	Version  string `yaml:"Version"`
}

type Fallback struct {
	URL              string `yaml:"URL"`
	CountryCode      string `yaml:"CountryCode"`
	ContinentCode    string `yaml:"ContinentCode"`
	Name             string `yaml:"Name"`
	NetworkBandwidth int32  `yaml:"NetworkBandwidth"`
}

type sentinels struct {
	Host string `yaml:"Host"`
}

type hashing struct {
	SHA1   bool `yaml:"SHA1"`
	SHA256 bool `yaml:"SHA256"`
	MD5    bool `yaml:"MD5"`
}

// LoadConfig loads the configuration file if it has not yet been loaded
func LoadConfig() {
	if config != nil {
		return
	}
	err := ReloadConfig()
	if err != nil {
		log.Fatal(err)
	}
}

// ReloadConfig reloads the configuration file and update it globally
func ReloadConfig() error {
	if core.ConfigFile == "" {
		if fileExists("/etc/mirrorbits.conf") {
			core.ConfigFile = "/etc/mirrorbits.conf"
		}
	}

	content, err := ioutil.ReadFile(core.ConfigFile)
	if err != nil {
		fmt.Println("Configuration could not be found.\n\tUse -config <path>")
		os.Exit(1)
	}

	if os.Getenv("DEBUG") != "" {
		fmt.Println("Reading configuration from", core.ConfigFile)
	}

	c := defaultConfig()

	// Overload the default configuration with the user's one
	err = yaml.Unmarshal(content, &c)
	if err != nil {
		return fmt.Errorf("%s in %s", err, core.ConfigFile)
	}

	// Sanitize
	if c.WeightDistributionRange <= 0 {
		return fmt.Errorf("WeightDistributionRange must be > 0")
	}
	if !isInSlice(c.OutputMode, []string{"auto", "json", "redirect"}) {
		return fmt.Errorf("Config: outputMode can only be set to 'auto', 'json' or 'redirect'")
	}
	if c.Repository == "" {
		return fmt.Errorf("Path to local repository not configured (see mirrorbits.conf)")
	}
	c.Repository, err = filepath.Abs(c.Repository)
	if err != nil {
		return fmt.Errorf("Invalid local repository path: %s", err)
	}
	if c.RepositoryScanInterval < 0 {
		c.RepositoryScanInterval = 0
	}

	if config != nil &&
		(c.RedisAddress != config.RedisAddress ||
			c.RedisPassword != config.RedisPassword ||
			!testSentinelsEq(c.RedisSentinels, config.RedisSentinels)) {
		// TODO reload redis connections
		// Currently established connections will be updated only in case of disconnection
	}
	//if config != nil &&
	//	(c.RedisAddress != GetRedisAddress() ||
	//		c.RedisPassword != GetRedisPwd() ||
	//		!testSentinelsEq(c.RedisSentinels, config.RedisSentinels)) {
	//	// TODO reload redis connections
	//	// Currently established connections will be updated only in case of disconnection
	//}

	// Lock the pointer during the swap
	configMutex.Lock()
	config = &c
	configMutex.Unlock()

	// Notify all subscribers that the configuration has been reloaded
	notifySubscribers()

	return nil
}

// GetConfig returns a pointer to a configuration object
// FIXME reading from the pointer could cause a race!
func GetConfig() *Configuration {
	configMutex.RLock()
	defer configMutex.RUnlock()

	if config == nil {
		panic("Configuration not loaded")
	}

	return config
}

// SetConfiguration is only used for testing purpose
func SetConfiguration(c *Configuration) {
	config = c
}

// SubscribeConfig allows subscribers to get notified when
// the configuration is updated.
func SubscribeConfig(subscriber chan bool) {
	subscribersLock.Lock()
	defer subscribersLock.Unlock()

	subscribers = append(subscribers, subscriber)
}

func notifySubscribers() {
	subscribersLock.RLock()
	defer subscribersLock.RUnlock()

	for _, subscriber := range subscribers {
		select {
		case subscriber <- true:
		default:
			// Don't block if the subscriber is unavailable
			// and discard the message.
		}
	}
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func testSentinelsEq(a, b []sentinels) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i].Host != b[i].Host {
			return false
		}
	}

	return true
}

// DUPLICATE
func isInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func GetRedisAddress() string {
	return os.Getenv("REDIS_ADDRESS_PORT")
}

func GetRedisPwd() string {
	return os.Getenv("REDIS_PWD")
}

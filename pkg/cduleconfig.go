package pkg

import (
	"gorm.io/gorm/logger"
)

// CduleConfig cdule configuration
type CduleConfig struct {
	// Whether to run scheduler immediately at startup; by default only run
	// after each tick (e.g if ticker is 60 seconds then first run has to wait
	// 60 seconds).
	RunImmediately   bool            `yaml:"runimmediately"`
	// The tick/refresh rate of workers, as a string acceptable by time.ParseDuration()
	TickDuration     string          `yaml:"tickduration"`
	Cduletype        string          `yaml:"cduletype"`
	Dburl            string          `yaml:"dburl"` // underscore creates the problem for e.f. db_url, so should be avoided
	Cduleconsistency string          `yaml:"cduleconsistency"`
	Loglevel         logger.LogLevel `yaml:"loglevel"` // gorm log level
	WatchPast        bool            `yaml:"watchpast"`
	TablePrefix      string          `yaml:"tableprefix"`
}

func NewDefaultConfig() *CduleConfig {
	return &CduleConfig{
		RunImmediately:   false,
		TickDuration:     "60s",
		Cduletype:        string(DATABASE),
		Dburl:            "postgres://cduleuser:cdulepassword@localhost:5432/cdule?sslmode=disable",
		Cduleconsistency: "AT_MOST_ONCE",
		Loglevel:         logger.Error,
		WatchPast:        false,
		TablePrefix:      "",
	}
}

func ResolveConfig(config ...*CduleConfig) *CduleConfig {
	var cfg *CduleConfig
	if nil == config || len(config) == 0 || config[0] == nil {
		cfg = NewDefaultConfig()
	} else {
		cfg = config[0]
	}

	// Set defaults
	if cfg.Loglevel == 0 {
		cfg.Loglevel = logger.Error
	}
	if cfg.TickDuration == "" {
		cfg.TickDuration = "60s"
	}

	return cfg
}

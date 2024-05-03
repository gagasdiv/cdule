package pkg

import "gorm.io/gorm/logger"

// CduleConfig cdule configuration
type CduleConfig struct {
	Cduletype        string          `yaml:"cduletype"`
	Dburl            string          `yaml:"dburl"` // underscore creates the problem for e.f. db_url, so should be avoided
	Cduleconsistency string          `yaml:"cduleconsistency"`
	Loglevel         logger.LogLevel `yaml:"loglevel"` // gorm log level
	WatchPast        bool            `yaml:"watchpast"`
	TablePrefix      string          `yaml:"tableprefix"`
}

func NewDefaultConfig() *CduleConfig {
	return &CduleConfig{
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

	return cfg
}

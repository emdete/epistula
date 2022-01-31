package main

import (
	"log"
	"os"
	"path"
	"github.com/robfig/config"
	)

type Config struct {
	database_path, user_name, user_primary_email string
}

func NewConfig() *Config {
	this := Config{}
	configfilename := os.Getenv("NOTMUCH_CONFIG")
	if configfilename == "" {
		configfilename = path.Join(os.Getenv("HOME"), ".notmuch-config")
	}
	if cfg, err := config.ReadDefault(configfilename); err != nil {
		panic(err)
	} else {
		log.Printf("%#v", cfg)
		this.database_path,_ = cfg.String("database", "path")
		this.user_name,_ = cfg.String("user", "name")
		this.user_primary_email,_ = cfg.String("user", "primary_email")
	}
	return &this
}


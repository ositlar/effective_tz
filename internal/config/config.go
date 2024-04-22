package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `yaml:"env"`
	StoragePath string `yaml:"storage_path"`
	Host        string `yaml:"host"`
	Port        string `yaml:"port"`
}

func MustConfig() *Config {
	err := godotenv.Load("C:\\Users\\Максим\\GolandProjects\\tz\\internal\\config\\configs\\local.env")
	if err != nil {
		log.Fatal("Can not load local.env file", err)
		return nil
	}

	// if _, err := os.Stat("C:\\Users\\Максим\\GolandProjects\\tz\\internal\\config\\configs\\config.yaml"); os.IsNotExist(err) {
	// 	panic("config file does not exists: config.yaml")
	// }
	var cfg Config
	// if err := cleanenv.ReadConfig("C:\\Users\\Максим\\GolandProjects\\tz\\internal\\config\\configs\\config.yaml", &cfg); err != nil {
	// 	fmt.Println(err)
	// 	panic("cannot read config.yaml")
	// }
	cfg.Env = os.Getenv("env")
	cfg.Host = os.Getenv("host")
	cfg.Port = os.Getenv("port")
	cfg.StoragePath = os.Getenv("storage_path")
	return &cfg
}

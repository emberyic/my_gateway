package config

import(
	"os"
	"gopkg.in/yaml.v3"
)

type Route struct{
	Path string `yaml:"path"`
	Backend string `yaml:"backend"`
}

type Config struct{
	Routes []Route `yaml:"routes"`
}

func LoadConfig(filename string) (*Config, error){
	data,err := os.ReadFile(filename)
	if err != nil{
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil{
		return nil, err
	}
	return &config, nil
}
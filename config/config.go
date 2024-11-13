package config

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// 配置文件结构
type SiteConfig struct {
	Name        string            `yaml:"name"`
	BaseURL     string            `yaml:"base_url"`
	ParseRules  map[string]string `yaml:"parse_rules"`
	DateFormats []string          `yaml:"date_formats"`
}

type Config struct {
	Sites []SiteConfig `yaml:"sites"`
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	file, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("无法读取配置文件: %v", err)
	}

	var config Config
	err = yaml.Unmarshal(file, &config)
	if err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}
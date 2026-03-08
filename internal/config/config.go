package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type TaskSets struct {
	Name  string   `yaml:"-"`
	Tasks []string `yaml:"tasks"`
}

type Project struct {
	Name      string     `yaml:"name"`
	Pinned    bool       `yaml:"pinned,omitempty"`
	Path      string     `yaml:"path"`
	Execution *Execution `yaml:"execution"`
	Tasks     []string   `yaml:"tasks,omitempty"`
	TasksFrom string     `yaml:"tasks_from,omitempty"`
	Tags      []string   `yaml:"tags,omitempty"`

	Exists bool `yaml:"-"`
}

type Scan struct {
	Paths []string
}

type Execution struct {
	Type      string `yaml:"type"`
	Container string `yaml:"container,omitempty"`
}

type Interface struct {
	Layout string `yaml:"layout,omitempty"` // card, card-compact, table, table-compact
}

type Config struct {
	Execution *Execution          `yaml:"execution"`
	Theme     string              `yaml:"theme"`
	Interface Interface           `yaml:"interface,omitempty"`
	Scan      Scan                `yaml:"scan"`
	Projects  []Project           `yaml:"projects"`
	TaskSets  map[string][]string `yaml:"task_sets,omitempty"`
}

// ConfigPath returns the absolute path to the charon config file.
func ConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get home directory: %w", err)
	}
	return filepath.Join(homeDir, ".config", "charon", ".charon.yaml"), nil
}

func Load() (*Config, error) {
	configPath, err := ConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{Projects: []Project{}}, nil
		}
		return nil, fmt.Errorf("failed to read config file %s: %w", configPath, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	for i := range cfg.Projects {
		cfg.Projects[i].Path = expandPath(cfg.Projects[i].Path)
	}

	for i := range cfg.Scan.Paths {
		cfg.Scan.Paths[i] = expandPath(cfg.Scan.Paths[i])
	}

	return &cfg, nil
}

func Save(cfg *Config) error {
	configPath, err := ConfigPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Contract paths back to ~ for saving.
	saveCfg := *cfg
	saveCfg.Projects = make([]Project, len(cfg.Projects))
	for i, p := range cfg.Projects {
		saveCfg.Projects[i] = p
		saveCfg.Projects[i].Path = contractPath(p.Path)
	}

	data, err := yaml.Marshal(&saveCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

func expandPath(path string) string {
	if len(path) > 0 && path[0] == '~' {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		return filepath.Join(homeDir, path[1:])
	}
	return path
}

func contractPath(path string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	if strings.HasPrefix(path, homeDir) {
		return "~" + path[len(homeDir):]
	}
	return path
}

func (c *Config) FindProject(name string) (*Project, error) {
	for i := range c.Projects {
		if c.Projects[i].Name == name {
			return &c.Projects[i], nil
		}
	}
	return nil, fmt.Errorf("project %q not found", name)
}

func (c *Config) HasProjectPath(path string) bool {
	for _, p := range c.Projects {
		if p.Path == path {
			return true
		}
	}
	return false
}

func (c *Config) AddProject(p Project) {
	c.Projects = append(c.Projects, p)
}

func ProjectFromPath(path string) Project {
	return Project{
		Name: filepath.Base(path),
		Path: path,
	}
}

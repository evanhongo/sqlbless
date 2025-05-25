package sqlbless

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"
)

type SourceConfig struct {
	ActiveSource string   `yaml:"active_source"`
	Sources      []Source `yaml:"sources"`
}

type Source struct {
	Name string `yaml:"name"`
	DB   string `yaml:"db"`
	DSN  string `yaml:"dsn"`
}

func expandHome(path string) (string, error) {
	if path[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		return filepath.Join(usr.HomeDir, path[2:]), nil
	}
	return path, nil
}

func loadConfig(path string) (*SourceConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg SourceConfig
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func saveConfig(path string, cfg *SourceConfig) error {
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func addSource(cfg *SourceConfig, newSource Source) {
	for _, s := range cfg.Sources {
		if s.Name == newSource.Name {
			fmt.Println("Source already exists, skipping.")
			return
		}
	}
	cfg.Sources = append(cfg.Sources, newSource)
}

func setActiveSource(cfg *SourceConfig, name string) error {
	for _, s := range cfg.Sources {
		if s.Name == name {
			cfg.ActiveSource = name
			return nil
		}
	}
	return fmt.Errorf("source '%s' not found", name)
}

func listSources(cfg *SourceConfig) {
	table := tablewriter.NewWriter(os.Stdout)
	table.Header([]string{"NAME", "DB", "DSN"})

	for _, src := range cfg.Sources {
		prefix := ""
		if src.Name == cfg.ActiveSource {
			prefix = "* "
		}

		dsn := src.DSN
		if len(dsn) > 60 {
			dsn = dsn[:60] + "..."
		}

		table.Append([]string{prefix + src.Name, src.DB, dsn})
	}

	table.Render()
}

func getActiveSourceDetails() ([]string, error) {
	configPath, err := expandHome("~/.config/sqlbless/config.yml")
	if err != nil {
		panic(err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		panic(err)
	}

	activeName := cfg.ActiveSource
	for _, src := range cfg.Sources {
		if src.Name == activeName {
			return []string{src.DB, src.DSN}, nil
		}
	}

	return nil, fmt.Errorf("active source '%s' not found in sources", activeName)
}

func handleCommand(command string, args []string) {
	configPath, err := expandHome("~/.config/sqlbless/config.yml")
	if err != nil {
		panic(err)
	}

	cfg, err := loadConfig(configPath)
	if err != nil {
		panic(err)
	}

	if command == "ls" {
		listSources(cfg)
		return
	}

	if command == "add" {
		name := args[1]
		db := args[2]
		dsn := args[3]
		addSource(cfg, Source{Name: name, DB: db, DSN: dsn})

		if err := saveConfig(configPath, cfg); err != nil {
			panic(err)
		}

		fmt.Println("Source added successfully!")

		return
	}

	if command == "src" {
		name := args[1]
		if err := setActiveSource(cfg, name); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		if err := saveConfig(configPath, cfg); err != nil {
			panic(err)
		}
		fmt.Printf("Active source set to '%s'\n", name)

		return
	}
}

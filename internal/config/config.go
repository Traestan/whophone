package config

// DB structure for storing a link to a database file
type DB struct {
	Path   string
	Bucket string
}

// FileNumCap structure for storing links to files
type Filesnumcap struct {
	Pathfilestore string `toml:"pathfilestore,omitempty"`
	Filename      []string
}

// TomlConfig common configuration file
type TomlConfig struct {
	Port   string
	Db     DB
	Numcap Filesnumcap
}

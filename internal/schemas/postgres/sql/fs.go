package sql

import "embed"

//go:embed migrator/*.tmpl
var Migrator embed.FS

//go:embed client/*.tmpl
var Client embed.FS

//go:embed migrations/*.tmpl
var Migrations embed.FS

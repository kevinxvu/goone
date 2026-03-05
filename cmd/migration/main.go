package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/pressly/goose/v3"
	"github.com/vuduongtp/go-core/config"
	"github.com/vuduongtp/go-core/pkg/database"
	cfgutil "github.com/vuduongtp/go-core/pkg/util/config"

	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/mattn/go-sqlite3"
)

const (
	defaultMigrationsDir = "internal/migrations"
)

var (
	flags        = flag.NewFlagSet("goose", flag.ExitOnError)
	dir          = flags.String("dir", defaultMigrationsDir, "directory with migration files")
	table        = flags.String("table", "goose_db_version", "migrations table name")
	verbose      = flags.Bool("v", false, "enable verbose mode")
	help         = flags.Bool("h", false, "print help")
	version      = flags.Bool("version", false, "print version")
	certfile     = flags.String("certfile", "", "file path to root CA's certificates in pem format (only support on mysql)")
	sequential   = flags.Bool("s", false, "use sequential numbering for new migrations")
	allowMissing = flags.Bool("allow-missing", false, "applies missing (out-of-order) migrations")
	sslcert      = flags.String("ssl-cert", "", "file path to SSL certificates in pem format (only support on mysql)")
	sslkey       = flags.String("ssl-key", "", "file path to SSL key in pem format (only support on mysql)")
	noVersioning = flags.Bool("no-versioning", false, "apply migration commands with no versioning, in file order, from directory pointed to")
)

func main() {
	flags.Usage = usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		log.Fatalf("Error parsing flags: %v", err)
	}

	if *version {
		fmt.Printf("goose version: %s\n", goose.VERSION)
		return
	}

	if *help {
		flags.Usage()
		return
	}

	args := flags.Args()
	if len(args) == 0 {
		flags.Usage()
		return
	}

	command := args[0]

	// Load configuration
	cfg := new(config.Configuration)
	if err := cfgutil.LoadConfig(cfg, "gocore", cfg.Stage); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get database connection
	db, err := openDB(cfg)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}
	defer db.Close()

	// Set goose options
	if *table != "" {
		goose.SetTableName(*table)
	}
	if *verbose {
		goose.SetVerbose(true)
	}
	if *sequential {
		goose.SetSequential(true)
	}
	if err := goose.SetDialect(getDialect(cfg.DbType)); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	// Special handling for create command
	if command == "create" {
		if len(args) < 2 {
			flags.Usage()
			return
		}

		if err := goose.Create(db, *dir, args[1], "sql"); err != nil {
			log.Fatalf("goose create: %v", err)
		}
		return
	}

	// Run the migration command
	if err := runCommand(db, command, args[1:]); err != nil {
		log.Fatalf("goose %s: %v", command, err)
	}
}

func openDB(cfg *config.Configuration) (*sql.DB, error) {
	driver, dsn := database.GetDriverAndDSN(cfg.DbType, cfg.DbLog, cfg.DbDsn)

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

func getDialect(dbType string) string {
	switch dbType {
	case "mysql":
		return "mysql"
	case "postgres":
		return "postgres"
	case "sqlite":
		return "sqlite3"
	default:
		return dbType
	}
}

func runCommand(db *sql.DB, command string, args []string) error {
	switch command {
	case "up":
		if *noVersioning {
			return goose.UpByOne(db, *dir)
		}
		if *allowMissing {
			return goose.UpByOne(db, *dir)
		}
		return goose.Up(db, *dir)

	case "up-by-one":
		return goose.UpByOne(db, *dir)

	case "up-to":
		if len(args) == 0 {
			return fmt.Errorf("up-to requires a version argument")
		}
		return goose.UpTo(db, *dir, parseVersion(args[0]))

	case "down":
		if *noVersioning {
			return goose.DownTo(db, *dir, 0)
		}
		return goose.Down(db, *dir)

	case "down-to":
		if len(args) == 0 {
			return fmt.Errorf("down-to requires a version argument")
		}
		return goose.DownTo(db, *dir, parseVersion(args[0]))

	case "redo":
		return goose.Redo(db, *dir)

	case "reset":
		return goose.Reset(db, *dir)

	case "status":
		return goose.Status(db, *dir)

	case "version":
		return goose.Version(db, *dir)

	case "fix":
		return goose.Fix(*dir)

	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func parseVersion(s string) int64 {
	var version int64
	if _, err := fmt.Sscanf(s, "%d", &version); err != nil {
		log.Fatalf("Invalid version: %s", s)
	}
	return version
}

func usage() {
	fmt.Println(`
goose - database migration tool

Usage:
    go run cmd/migration/main.go [options] <command> [arguments...]

Commands:
    up                   Migrate the DB to the most recent version available
    up-by-one            Migrate the DB up by 1
    up-to VERSION        Migrate the DB to a specific VERSION
    down                 Roll back the version by 1
    down-to VERSION      Roll back to a specific VERSION
    redo                 Re-run the latest migration
    reset                Roll back all migrations
    status               Dump the migration status for the current DB
    version              Print the current version of the database
    create NAME [sql]    Creates new migration file with the current timestamp
    fix                  Apply sequential ordering to migrations

Options:
    -dir string          Directory with migration files (default "internal/migrations")
    -table string        Migrations table name (default "goose_db_version")
    -v                   Enable verbose mode
    -h                   Print this help
    -version             Print goose version
    -s                   Use sequential numbering for new migrations
    -allow-missing       Allow missing (out-of-order) migrations
    -no-versioning       Apply migrations with no versioning

Examples:
    go run cmd/migration/main.go status
    go run cmd/migration/main.go up
    go run cmd/migration/main.go down
    go run cmd/migration/main.go create add_users_table sql
    go run cmd/migration/main.go -dir=pkg/migrations up
`)
}

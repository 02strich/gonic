package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/peterbourgon/ff"

	"senan.xyz/g/gonic/db"
	"senan.xyz/g/gonic/scanner"
	"senan.xyz/g/gonic/version"
)

func main() {
	set := flag.NewFlagSet(version.NAME_SCAN, flag.ExitOnError)
	musicPath := set.String(
		"music-path", "",
		"path to music")
	sqlitePath := set.String("db-path", "gonic.db", "path to database (optional, default: gonic.db)")
	postgresHost := set.String("postgres-host", "", "name of the PostgreSQL server (optional)")
	postgresPort := set.Int("postgres-port", 5432, "port to use for PostgreSQL connection (optional, default: 5432)")
	postgresName := set.String("postgres-db", "gonic", "name of the PostgreSQL database (optional, default: gonic)")
	postgresUser := set.String("postgres-user", "gonic", "name of the PostgreSQL user (optional, default: gonic)")
	_ = set.String("config-path", "", "path to config (optional)")
	showVersion := set.Bool("version", false, "show gonic version")
	if err := ff.Parse(set, os.Args[1:],
		ff.WithConfigFileFlag("config-path"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix(version.NAME_UPPER),
	); err != nil {
		log.Fatalf("error parsing args: %v\n", err)
	}
	if _, err := os.Stat(*musicPath); os.IsNotExist(err) {
		log.Fatal("please provide a valid music directory")
    }

	if *showVersion {
		fmt.Println(version.VERSION)
		os.Exit(0)
	}
	var database *db.DB
	if len(*postgresHost) > 0 {
		database, err = db.NewPostgres(*postgresHost, *postgresPort, *postgresName, *postgresUser, os.Getenv("GONIC_POSTGRES_PW"))
	} else {
		database, err = db.NewSqlite3(*sqlitePath)
	}
	if err != nil {
		log.Fatalf("error opening database: %v\n", err)
	}
	defer database.Close()

	s := scanner.New(
		*musicPath,
		database,
	)
	if err := s.Start(); err != nil {
		log.Fatalf("error starting scanner: %v\n", err)
	}
}

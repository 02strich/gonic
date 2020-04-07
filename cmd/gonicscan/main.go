package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/peterbourgon/ff"

	"senan.xyz/g/gonic/db"
	"senan.xyz/g/gonic/dir"
	"senan.xyz/g/gonic/scanner"
	"senan.xyz/g/gonic/version"
)

func main() {
	set := flag.NewFlagSet(version.NAME_SCAN, flag.ExitOnError)
	localMusicPath := set.String("music-path", "", "path to music (optional)")
	remoteMusicS3Region := set.String("remote-music-s3-region", "us-west-2", "region of the S3 bucket to read music from (optional, default us-west-2)")
	remoteMusicS3Bucket := set.String("remote-music-s3-bucket", "", "name of the S3 bucket to read music from (optional)")
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

	if *showVersion {
		fmt.Println(version.VERSION)
		os.Exit(0)
	}

	var musicDir dir.Dir
	var err error
	if len(*remoteMusicS3Bucket) > 0 {
		musicDir, err = dir.NewS3Dir(*remoteMusicS3Region, *remoteMusicS3Bucket)
	} else {
		musicDir, err = dir.NewLocalDir(*localMusicPath)
	}
	if err != nil {
		log.Fatalf("please provide a valid music directory: %v\n", err)
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
		database,
		musicDir,
	)
	if err := s.Start(); err != nil {
		log.Fatalf("error starting scanner: %v\n", err)
	}
}

package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"time"

	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/peterbourgon/ff"

	"senan.xyz/g/gonic/db"
	"senan.xyz/g/gonic/dir"
	"senan.xyz/g/gonic/server"
	"senan.xyz/g/gonic/version"
)

func main() {
	set := flag.NewFlagSet(version.NAME, flag.ExitOnError)
	listenAddr := set.String("listen-addr", "0.0.0.0:4747", "listen address (optional)")
	localMusicPath := set.String("music-path", "", "path to music (optional)")
	cachePath := set.String("cache-path", "/tmp/gonic_cache", "path to cache (optional, default: /tmp/gonic_cache)")
	sqlitePath := set.String("db-path", "gonic.db", "path to database (optional, default: gonic.db)")
	postgresHost := set.String("postgres-host", "", "name of the PostgreSQL server (optional)")
	postgresPort := set.Int("postgres-port", 5432, "port to use for PostgreSQL connection (optional, default: 5432)")
	postgresName := set.String("postgres-db", "gonic", "name of the PostgreSQL database (optional, default: gonic)")
	postgresUser := set.String("postgres-user", "gonic", "name of the PostgreSQL user (optional, default: gonic)")
	scanInterval := set.Int("scan-interval", 0, "interval (in minutes) to automatically scan music (optional)")
	proxyPrefix := set.String("proxy-prefix", "", "url path prefix to use if behind proxy. eg '/gonic' (optional)")
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

	musicDir, err := dir.NewLocalDir(*localMusicPath)
	if err != nil {
		log.Fatalf("please provide a valid music directory: %v\n", err)
	}

	if _, err := os.Stat(*cachePath); os.IsNotExist(err) {
		if err := os.MkdirAll(*cachePath, os.ModePerm); err != nil {
			log.Fatalf("couldn't create cache path: %v\n", err)
		}
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

	proxyPrefixExpr := regexp.MustCompile(`^\/*(.*?)\/*$`)
	*proxyPrefix = proxyPrefixExpr.ReplaceAllString(*proxyPrefix, `/$1`)
	serverOptions := server.Options{
		DB:           database,
		MusicDir:     musicDir,
		CachePath:    *cachePath,
		ListenAddr:   *listenAddr,
		ScanInterval: time.Duration(*scanInterval) * time.Minute,
		ProxyPrefix:  *proxyPrefix,
	}

	log.Printf("using opts %+v\n", serverOptions)
	s := server.New(serverOptions)

	log.Printf("starting server at %s", *listenAddr)
	if err := s.Start(); err != nil {
		log.Fatalf("error starting server: %v\n", err)
	}
}

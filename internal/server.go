package tapedeck

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	// avoid clash with local var 'db'
	dbpkg "tapedeck/internal/db"

	// avoid clash with local var 'user'
	userpkg "tapedeck/internal/db/user"
)

// checkDir will join the parentDir to dirName and check that the new dir exists.
func checkDir(parentDir string, dirName string) (fullDir string, err error) {
	log.Println("enter checkDir", parentDir, dirName)
	defer log.Println("exit checkDir")

	fullDir = filepath.Join(parentDir, dirName)
	_, err = os.Stat(fullDir)
	if err != nil {
		err = fmt.Errorf("failed to verify %s in %s: %w", dirName, parentDir, err)
	}
	return
}

type ServerConfig struct {
	ServerDir        string `json:"serverDir"`
	UserDir          string `json:"userDir"`
	ServerListenAddr string `json:"serverListenAddr"`
	DbFile           string `json:"dbFile"`
	ProductionMode   bool   `jons:"productionMode"`
}

func readConfig(jsonFilePath string) (ServerConfig, error) {
	log.Println("enter readConfig", jsonFilePath)
	defer log.Println("exit readConfig")

	var config ServerConfig

	bytes, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return config, err
	}

	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return config, err
	}

	config.ServerDir = filepath.Join(config.ServerDir)
	config.UserDir = filepath.Join(config.UserDir)

	return config, nil
}

func RunServer(jsonConfigPath string) (rc int, err error) {
	log.Println("enter RunServer", jsonConfigPath)
	defer log.Println("exit RunServer")

	config, configErr := readConfig(jsonConfigPath)
	if configErr != nil {
		return 209, configErr
	}

	// config validation
	log.Println("validate server directory")
	_, dirErr := os.Stat(config.ServerDir)
	if dirErr != nil {
		return 148, dirErr
	}

	log.Println("validate user directory")
	_, dirErr2 := os.Stat(config.UserDir)
	if dirErr2 != nil {
		return 95, dirErr2
	}

	log.Println("validate server listen address")
	if strings.Count(config.ServerListenAddr, ":") != 1 {
		return 92, fmt.Errorf("invalid server listen address: %v", config.ServerListenAddr)
	}

	log.Println("validate sqlite db file path")
	_, dbErr := os.Stat(config.DbFile)
	if dbErr != nil {
		return 155, dbErr
	}

	log.Println("using sqlite file", config.DbFile)
	db := &dbpkg.Database{
		FilePath: config.DbFile,
	}

	log.Println("upgrade check")
	upgrade, checkErr := db.UpgradeCheck(userpkg.SchemaVersion)
	if checkErr != nil {
		return 200, fmt.Errorf("db upgrade check failed: %w", checkErr)
	}

	if upgrade {
		log.Println("upgrade required")
		db.Upgrade()
	} else {
		log.Println("upgrade not required")
	}

	templateDir, tmplErr := checkDir(config.ServerDir, "templates")
	if tmplErr != nil {
		return 210, tmplErr
	}

	cache := config.ProductionMode
	log.Println("using template directory", templateDir, "and cache", cache)
	tmplEngine := NewTemplateEngine(templateDir, cache)
	initErr := tmplEngine.Init()
	if initErr != nil {
		return 217, fmt.Errorf("template engine init failed: %w", initErr)
	}

	staticDir, staticErr := checkDir(config.ServerDir, "static")
	if staticErr != nil {
		return 222, staticErr
	}
	log.Println("using static directory", staticDir)

	log.Println("server verification complete")

	// Open routes
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", MakeRootHandler(tmplEngine))

	// Secure routes
	http.HandleFunc("/s/list", MakeListHandler(db, tmplEngine))
	http.HandleFunc("/s/playback", MakePlaybackHandler(db, tmplEngine))
	http.HandleFunc("/s/record", MakeRecordHandler(db, tmplEngine))

	//TODO: define middelware to dump headers based on log level.
	//TODO: extract `X-Email` header value and maybe `X-User`.
	//X-Email:[rocketnewton@gmail.com]
	//X-User:[112165920196384629909]

	log.Println("server starting on", config.ServerListenAddr)
	err = http.ListenAndServe(config.ServerListenAddr, nil)
	if err != nil {
		return 286, err
	} else {
		return 0, nil
	}
}

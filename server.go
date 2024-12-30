package tapedeck

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func checkDir(serverDir string, dirName string) (fullDir string, err error) {
	log.Println("enter checkDir", serverDir, dirName)
	defer log.Println("exit checkDir")

	fullDir = filepath.Join(serverDir, dirName)
	_, err = os.Stat(fullDir)
	if err != nil {
		err = fmt.Errorf("failed to verify %s in %s: %w", dirName, serverDir, err)
	}
	return
}

type ServerConfig struct {
	ServerDir        string `json:"serverDir"`
	ServerListenAddr string `json:"serverListenAddr"`
	DbFile           string `json:"dbFile"`
	productionMode   bool
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

	return config, nil
}

func RunDevServer(serverDir string, listenAddr string) (rc int, err error) {
	log.Println("enter RunDevServer", serverDir, listenAddr)
	defer log.Println("exit RunDevServer")

	config := ServerConfig{
		ServerDir:        serverDir,
		ServerListenAddr: listenAddr,
		DbFile:           filepath.Join(serverDir, "tapedeck.db"),
		productionMode:   false,
	}

	return start(config)
}

func RunProdServer(jsonConfigPath string) (rc int, err error) {
	log.Println("enter RunProdServer", jsonConfigPath)
	defer log.Println("exit RunProdServer")

	config, configErr := readConfig(jsonConfigPath)
	if configErr != nil {
		return 209, configErr
	}

	// populate private fields
	config.productionMode = true

	return start(config)
}

func start(config ServerConfig) (rc int, err error) {
	log.Println("enter start", config)
	defer log.Println("exit start")

	// config validation
	log.Println("validate server directory")
	_, dirErr := os.Stat(config.ServerDir)
	if dirErr != nil {
		return 148, dirErr
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
	db := &Database{config.DbFile}
	dbConnErr := db.TestConnection()
	if dbConnErr != nil {
		return 200, fmt.Errorf("db connection test failed: %w", dbConnErr)
	}

	templateDir, tmplErr := checkDir(config.ServerDir, "templates")
	if tmplErr != nil {
		return 210, tmplErr
	}

	cache := config.productionMode
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

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", MakeRootHandler(tmplEngine))
	http.HandleFunc("/list", MakeListHandler(db, tmplEngine))
	http.HandleFunc("/playback", MakePlaybackHandler(db, tmplEngine))
	http.HandleFunc("/record", MakeRecordHandler(db, tmplEngine))

	log.Println("server starting on", config.ServerListenAddr)
	err = http.ListenAndServe(config.ServerListenAddr, nil)
	if err != nil {
		return 286, err
	} else {
		return 0, nil
	}
}

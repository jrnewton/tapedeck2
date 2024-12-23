package tapedeck

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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
	ServerListenPort int    `json:"serverListenPort"`
	DbFile           string `json:"dbFile"`
	CertFile         string `json:"certFile"`
	KeyFile          string `json:"keyFile"`
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

func RunDevServer(serverDir string, serverPort int) (rc int, err error) {
	log.Println("enter RunDevServer", serverDir, serverPort)
	defer log.Println("exit RunDevServer")

	config := ServerConfig{
		ServerDir:        serverDir,
		ServerListenAddr: "",
		ServerListenPort: serverPort,
		DbFile:           filepath.Join(serverDir, "tapedeck.db"),
		CertFile:         "",
		KeyFile:          "",
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

	log.Println("validate sqlite db file path")
	_, dbErr := os.Stat(config.DbFile)
	if dbErr != nil {
		return 155, dbErr
	}

	if config.CertFile != "" {
		log.Println("validate certificate file path")
		_, certErr := os.Stat(config.CertFile)
		if certErr != nil {
			return 163, certErr
		}
	}

	if config.KeyFile != "" {
		// TODO: check permissions on this file and panic if not strict enough.
		log.Println("validate key file path")
		_, keyErr := os.Stat(config.KeyFile)
		if keyErr != nil {
			return 176, keyErr
		}
	}

	if config.productionMode {
		if config.CertFile == "" || config.KeyFile == "" {
			return 213, fmt.Errorf("production server requires TLS but CertFile or KeyFile empty")
		}
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

	//TODO: implement ACME support to allow for certbot autorenew.
	//Always run a server on HTTP, port 80 and serve ./.well-known/acme-challenge/
	//https://eff-certbot.readthedocs.io/en/stable/using.html#webroot

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))
	http.HandleFunc("/", MakeRootHandler(tmplEngine))
	http.HandleFunc("/list", MakeListHandler(db, tmplEngine))
	http.HandleFunc("/playback", MakePlaybackHandler(db, tmplEngine))
	http.HandleFunc("/record", MakeRecordHandler(db, tmplEngine))

	listenAddress := config.ServerListenAddr + ":" + strconv.Itoa(config.ServerListenPort)

	if config.productionMode {
		log.Println("production server starting in HTTPS mode. listen address is", listenAddress)
		listener, listenErr := net.Listen("tcp4", listenAddress)
		if listenErr != nil {
			return 292, listenErr
		}

		err = http.ServeTLS(listener, nil, config.CertFile, config.KeyFile)
		if err != nil {
			return 297, err
		} else {
			return 0, nil
		}
	} else {
		log.Println("dev server starting in HTTP mode. listen address is", listenAddress)
		err = http.ListenAndServe(listenAddress, nil)
		if err != nil {
			return 286, err
		} else {
			return 0, nil
		}
	}
}

package main

import(
	"os"
	"log"
	"time"
	"runtime"
	"math/rand"
	"encoding/json"
	"path/filepath"

	"github.com/GoTyro/pool-proxy/server"
)

var cfg server.Config

func readConfig(cfg *server.Config) {
	configFileName := "config.json"
	if len(os.Args) > 1 {
		configFileName = os.Args[1]
	}
	configFileName, _ = filepath.Abs(configFileName)
	log.Printf("加载配置文件: %v", configFileName)

	configFile, err := os.Open(configFileName)
	if err != nil {
		log.Fatal("配置文件加载失败,错误: %v", err.Error())
	}
	defer configFile.Close()
	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		log.Fatal("配置文件有错误,详情: %v", err.Error())
	}
}

func startProxy() {
	//随机种子
	rand.Seed(time.Now().UnixNano())

	if cfg.Threads > 0 {
		runtime.GOMAXPROCS(cfg.Threads)
		log.Printf("启动程序,线程数量: %v", cfg.Threads)
	} else {
		n := runtime.NumCPU()
		runtime.GOMAXPROCS(n)
		log.Printf("启动程序,使用CPU数量设置线程: %v", n)
	}

	s := server.NewProxy(&cfg)
	s.Start()
}

func main() {
	//读取配置文件
	readConfig(&cfg)

	//启动主程序
	go startProxy()

	quit := make(chan bool)
	<-quit
}

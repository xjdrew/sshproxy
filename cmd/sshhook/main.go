package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/xjdrew/sshproxy/pkg/webhook"
)

func main() {
	configFile := flag.String("config", "webhook.yaml", "path to webhook config file")
	flag.Parse()

	// 加载配置
	config, err := webhook.LoadConfig(*configFile)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 创建 webhook 服务
	server, err := webhook.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create webhook server: %v", err)
	}

	// 启动服务
	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start webhook server: %v", err)
	}

	log.Printf("Webhook server started on %s", config.Listen)

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down webhook server...")
	if err := server.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}
}

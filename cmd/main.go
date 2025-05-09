package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/shester1kov/myproxy/internal/config"
	"github.com/shester1kov/myproxy/internal/lb"
	"github.com/shester1kov/myproxy/internal/middleware"
	"github.com/shester1kov/myproxy/internal/proxy"
)

func main() {
	// получаем конфигурацию из файла
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v\n", err)
	}

	// создаем балансировщик нагрузки
	lb := lb.NewRoundRobin(cfg.Backends)
	proxy := proxy.NewProxy(lb)

	// создаем структуру для сервера
	server := http.Server{
		Addr:    ":" + strconv.Itoa(cfg.Port),
		Handler: middleware.LogMiddleware(proxy.Handler()),
	}

	// канал, в который будут записываться значения при остановке программы
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// запускаем сервер
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Ошибка сервера: %v", err)
		}
	}()

	log.Printf("Запуск сервера на %s", server.Addr)

	// при остановке сервера значение пишется в канал и программа продолжает выполнение
	<-done
	log.Println("Сервер останавливается...")

	// создаем контекст для установки таймаута
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// останавливаем сервер, ожидаем пока процессы завершатся
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Ошибка остановки сервера: %v", err)
	}

	log.Println("Сервер остановлен")
}

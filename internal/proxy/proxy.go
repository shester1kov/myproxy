package proxy

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/shester1kov/myproxy/internal/lb"
)

// структура для прокси
type Proxy struct {
	lb *lb.RoundRobin
}

// конструктор, возвращает структуру прокси
func NewProxy(lb *lb.RoundRobin) *Proxy {
	return &Proxy{
		lb: lb,
	}
}

// обработчик запросов, перенаправляет входящий запрос на другой сервер
func (p *Proxy) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//получаем адрес сервера из балансировщика нагрузки
		targetAddr, err := p.lb.GetNext()
		if err != nil {
			log.Printf("Ошибка получения сервера: %v\n", err)
			http.Error(w, "Сервис недоступен", http.StatusServiceUnavailable)
			return
		}

		log.Printf("[%s] %s -> %s\n", r.Method, r.URL.Path, targetAddr)

		targetURL, err := url.Parse(targetAddr)
		if err != nil {
			log.Printf("Ошибка парсинга URL: %v\n", err)
			http.Error(w, "Внутренняя ошибка сервера", http.StatusInternalServerError)
			return
		}

		// создаем новый прокси с полученным адресом
		proxy := httputil.NewSingleHostReverseProxy(targetURL)
		proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
			log.Printf("Ошибка проксирования на %s: %v\n", targetAddr, err)
			http.Error(w, "ошибка прокси: "+err.Error(), http.StatusBadGateway)
		}

		proxy.ServeHTTP(w, r)

	}
}

package lb

import (
	"errors"
	"net/http"
	"sync"
	"time"
)

// структура для сервера
type BackendStatus struct {
	URL   string
	Alive bool
	mutex sync.RWMutex
}

// структура для балансировщика нагрузки
type RoundRobin struct {
	backends []*BackendStatus
	current  int
	mu       sync.Mutex
}

// конструктор, возвращает балансировщик нагрузки
func NewRoundRobin(backends []string) *RoundRobin {
	rr := &RoundRobin{
		backends: make([]*BackendStatus, 0, len(backends)),
		current:  0,
	}

	// заполняем список серверов
	for _, b := range backends {
		rr.backends = append(
			rr.backends,
			&BackendStatus{
				URL:   b,
				Alive: true,
			},
		)
	}

	go rr.HealthCheck()
	return rr
}

// проверка работоспособности серверов
func (r *RoundRobin) HealthCheck() {
	ticker := time.NewTicker(10 * time.Second)
	client := http.Client{Timeout: 2 * time.Second}

	// запускается раз в 10 секунд
	for range ticker.C {
		for _, backend := range r.backends {
			go func(b *BackendStatus) {
				// отправка запроса на эндпоинт health
				resp, err := client.Get(b.URL + "/health")
				b.mutex.Lock()
				defer b.mutex.Unlock()

				// при ошибке переводим сервер в статус неактивного
				if err != nil || resp.StatusCode != http.StatusOK {
					b.Alive = false
					//log.Printf("Сервер %s не работает", b.URL)
				} else {
					b.Alive = true
					//log.Printf("Сервер %s работает", b.URL)
				}
			}(backend)
		}
	}
}

// метод для получения адреса сервера
func (r *RoundRobin) GetNext() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.backends) == 0 {
		return "", errors.New("нет доступных серверов")
	}

	start := r.current

	for {
		backend := r.backends[r.current]
		backend.mutex.RLock()
		alive := backend.Alive
		backend.mutex.RUnlock()

		r.current = (r.current + 1) % len(r.backends)

		// если сервер активен, возвращаем его адрес
		if alive {
			return backend.URL, nil
		}

		if r.current == start {
			break
		}
	}

	// если ни один из серверов не активен, возвращаем ошибку
	return "", errors.New("ни один сервер не доступен")

}

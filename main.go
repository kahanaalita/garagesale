package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func main() {
	log.Println("main : Started")
	defer log.Println("main : Completed")
	// Открываем базу данных
	db, err := openDB()
	if err != nil {
		log.Fatal(err)
		defer db.Close()
	}
	// Настраиваем сервер API
	api := http.Server{
		Addr:         "localhost:8000",
		Handler:      http.HandlerFunc(ListProducts),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	// Канал для ошибок сервера
	serverErrors := make(chan error, 1)

	// Запуск сервера в отдельной горутине
	go func() {
		log.Printf("main : API listening on %s", api.Addr)
		serverErrors <- api.ListenAndServe()
	}()

	// Канал для обработки сигналов завершения
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Ожидание завершения или ошибки сервера
	select {
	case err := <-serverErrors:
		log.Fatalf("error: listening and serving: %s", err)

	case <-shutdown:
		log.Println("main : Start shutdown")

		// Устанавливаем тайм-аут для завершения
		const timeout = 5 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		// Останавливаем сервер
		err := api.Shutdown(ctx)
		if err != nil {
			log.Printf("main : Graceful shutdown did not complete in %v : %v", timeout, err)
			err = api.Close()
			if err != nil {
				log.Fatalf("main : could not stop server gracefully : %v", err)
			}
		}
	}
}

func openDB() (*sqlx.DB, error) {
	q := url.Values{}
	q.Set("sslmode", "disable")
	q.Set("timezone", "utc")

	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("postgres", "sakhalin"),
		Host:     "localhost:5433",
		Path:     "postgres",
		RawQuery: q.Encode(),
	}

	return sqlx.Open("postgres", u.String())
}

// Product это то что мы продаем
type Product struct {
	Name     string `json:"name"`
	Cost     int    `json:"cost"`
	Quantity int    `json:"quantity"`
}

// ListProducts - базовый HTTP обработчик.
func ListProducts(w http.ResponseWriter, r *http.Request) {
	list := []Product{}
	if false {
		list = append(list, Product{Name: "Comic Book", Cost: 75, Quantity: 50})
		list = append(list, Product{Name: "Mc'Donalds toys", Cost: 25, Quantity: 120})
	}
	data, err := json.Marshal(list)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("error marshaling : %v", err)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if _, err := w.Write(data); err != nil {
		log.Printf("error writing : %v", err)
	}
}

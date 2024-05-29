package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"runtime"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CotacaoDolar struct {
	Usdbrl struct {
		Code       string `gorm:"code"`
		Codein     string `gorm:"codein"`
		Name       string `gorm:"name"`
		High       string `gorm:"high"`
		Low        string `gorm:"low"`
		VarBid     string `gorm:"varBid"`
		PctChange  string `gorm:"pctChange"`
		Bid        string `gorm:"bid"`
		Ask        string `gorm:"ask"`
		Timestamp  string `gorm:"timestamp"`
		CreateDate string `gorm:"create_date"`
	} `gorm:"USDBRL"`
}

type BidResponse struct {
	Bid string `json:"bid"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		structCotacaoDolar, err := getContacaoDolar(r.Context())

		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("erro"))
		}

		w.Header().Set("Content-Type:", "application/json")
		w.WriteHeader(http.StatusOK)

		bidResponse := BidResponse{
			Bid: structCotacaoDolar.Usdbrl.Bid,
		}

		json.NewEncoder(w).Encode(bidResponse)

	})

	http.ListenAndServe(":8080", mux)

}

func getContacaoDolar(ctx context.Context) (*CotacaoDolar, error) {

	client := &http.Client{
		Timeout: 200 * time.Millisecond,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)

	if err != nil {
		logErrorWithLine(err)
		return nil, err

	}

	res, err := client.Do(req)

	if err != nil {
		logErrorWithLine(err)
		return nil, err

	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		err := fmt.Errorf("failed to get cotacao dolar: %s", res.Status)
		logErrorWithLine(err)
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		logErrorWithLine(err)
		return nil, err

	}

	var contatoDolar CotacaoDolar
	err = json.Unmarshal(body, &contatoDolar)
	if err != nil {
		logErrorWithLine(err)
		return nil, err
	}

	err = storeContacaoDB(ctx, contatoDolar)
	if err != nil {
		logErrorWithLine(err)
		return nil, err
	}

	return &contatoDolar, nil

}

func logErrorWithLine(err error) {
	if err != nil {
		_, file, line, ok := runtime.Caller(1)
		if ok {
			log.Fatalf("%s:%d: %v", file, line, err)
		} else {
			log.Fatal(err)
		}
	}
}

func storeContacaoDB(ctx context.Context, c CotacaoDolar) error {

	db, err := connectionDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS contacaos (
		code TEXT,
		codein TEXT,
		name TEXT,
		high TEXT,
		low TEXT,
		varBid TEXT,
		pctChange TEXT,
		bid TEXT,
		ask TEXT,
		timestamp TEXT,
		create_date TEXT
	);`)

	logErrorWithLine(err)

	// Executar a instrução de inserção
	_, err = db.ExecContext(ctx, "INSERT INTO contacaos (code, codein, name, high, low, varBid, pctChange, bid, ask, timestamp, create_date) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		c.Usdbrl.Code,
		c.Usdbrl.Codein,
		c.Usdbrl.Name,
		c.Usdbrl.High,
		c.Usdbrl.Low,
		c.Usdbrl.VarBid,
		c.Usdbrl.PctChange,
		c.Usdbrl.Bid,
		c.Usdbrl.Ask,
		c.Usdbrl.Timestamp,
		c.Usdbrl.CreateDate,
	)

	logErrorWithLine(err)

	log.Println("Inserido com sucesso!")

	return err

}

func connectionDB() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "./db/database.db")
	if err != nil {
		logErrorWithLine(err)
		return nil, err
	}

	// Verificar se o banco de dados é válido
	err = db.Ping()
	if err != nil {
		logErrorWithLine(err)
		db.Close()
		return nil, err
	}

	return db, nil
}

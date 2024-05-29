package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type CotacaoDolar struct {
	Bid string `json:"bid"`
}

func main() {

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		fmt.Printf("Failed to make request: %s\n", err)
		return
	}

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		fmt.Printf("Failed to read response body: %s\n", err)
		return
	}

	defer resp.Body.Close()

	f, err := os.Create("cotacao.txt")

	if err != nil {
		fmt.Printf("Failed to encode : %s\n", err)
		return
	}

	r, err := io.ReadAll(resp.Body)

	if err != nil {
		fmt.Printf("Failed to encode : %s\n", err)
		return
	}

	var c CotacaoDolar

	json.Unmarshal(r, &c)

	tamanho, err := f.WriteString("Dolar: " + c.Bid)

	if err != nil {
		fmt.Printf("Failed to create file cotacao.txt: %s\n", err)
		return
	}
	fmt.Printf("Arquivo com tamanho: %d criado com sucesso!\n", tamanho)
	io.Copy(os.Stdout, resp.Body)

}

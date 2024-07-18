package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Cotacao struct {
	Bid float64 `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatal("Erro ao criar a requisição: ", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Erro ao efetuar a requisição", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Ocorreu algum erro no servidor! status: %v", resp.Status)
	}

	var result Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Fatal("Não foi possivel decodificar a resposta:", err)
	}

	bid := result.Bid
	fmt.Printf("Dólar: %f\n", bid)

	if err := ioutil.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar: %f", bid)), 0644); err != nil {
		log.Fatal("Hum! Parece que não conseguir gravar no arquivo! :", err)
	}
}

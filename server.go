package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Cotacao struct {
	ID  string  `gorm:"primaryKey"`
	Bid float64 `json:"bid"`
}

var db *gorm.DB
var sqlDB *sql.DB

func main() {
	var err error
	db, err = gorm.Open(sqlite.Open("cotacoes.db"), &gorm.Config{})
	if err != nil {
		log.Fatalf("Falha na abertura da base de dados, erro: %v", err)
	}

	sqlDB, err = db.DB()
	if err != nil {
		log.Fatalf("Falha ao carregar o objeto database, erro:  %v", err)
	}

	err = db.AutoMigrate(&Cotacao{})
	if err != nil {
		log.Fatal(err)
	}

	r := mux.NewRouter()
	r.HandleFunc("/cotacao", CotacaoHandler).Methods("GET")

	http.Handle("/", r)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func CotacaoHandler(w http.ResponseWriter, r *http.Request) {

	ctx, cancel := context.WithTimeout(r.Context(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		log.Println("Erro ao criar uma requisição ", err)
		http.Error(w, "Aconteceu algum erro interno verificar ", http.StatusInternalServerError)
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println("Erro na criação da requisição: ", err)
		http.Error(w, "Tempo excedido ", http.StatusGatewayTimeout)
		return
	}
	defer resp.Body.Close()

	var result map[string]struct {
		Bid string `json:"bid"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Println("Erro ao receber a resposta decodificada:", err)
		http.Error(w, "Aconteceu algum erro interno verificar  ", http.StatusInternalServerError)
		return
	}

	bidStr := result["USDBRL"].Bid
	bid, err := strconv.ParseFloat(bidStr, 64)
	if err != nil {
		log.Println("Erro ao converter o valor do dolar:", err)
		http.Error(w, "Aconteceu algum erro interno verificar  ", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]float64{"bid": bid})

	go saveCotacao(bid)
}

func saveCotacao(bid float64) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	id := uuid.New().String()

	stmt, err := sqlDB.PrepareContext(ctx, "INSERT INTO cotacaos (id, bid) VALUES (?, ?)")
	if err != nil {
		log.Println("Houve falha nos parametros:", err)
		return
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, id, bid)
	if err != nil {
		log.Println("Ops! Parece que não consegui salvar na base de dados:", err)
		return
	}

	log.Printf("Cotacao salva com sucesso : %s - %f", id, bid)
}

package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "github.com/mattn/go-sqlite3"
)

type Cotacao struct {
	Usdbrl struct {
		Code       string `json:"code"`
		Codein     string `json:"codein"`
		Name       string `json:"name"`
		High       string `json:"high"`
		Low        string `json:"low"`
		VarBid     string `json:"varBid"`
		PctChange  string `json:"pctChange"`
		Bid        string `json:"bid"`
		Ask        string `json:"ask"`
		Timestamp  string `json:"timestamp"`
		CreateDate string `json:"create_date"`
	} `json:"USDBRL"`
}

type Dolar struct {
	Id          int64   `gorm:"primaryKey" json:"-"`
	ValorCambio float64 `json:"valor"`
}

type ErrorResp struct {
	Mensagem string
}

func main() {

	http.HandleFunc("/cotacao", BuscaCotacao)
	http.ListenAndServe(":8080", nil)

}

func BuscaCotacao(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if r.URL.Path != "/cotacao" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	cot, err := ExecutaChamadaRest()
	if err != nil {
		log.Println("erro api", err)
		retorno := ErrorResp{Mensagem: err.Error()}
		json.NewEncoder(w).Encode(retorno)
		return
	}

	//converte para float e prepara json retorno
	var retorno Dolar
	retorno.ValorCambio, err = strconv.ParseFloat(cot.Usdbrl.Bid, 64)
	if err != nil {
		log.Println("erro conversao", err)
		retorno := ErrorResp{Mensagem: err.Error()}
		json.NewEncoder(w).Encode(retorno)
		return
	}

	//persiste a cotacao no banco de dados
	err = PersisteCotacao(&retorno)
	if err != nil {
		log.Println("erro Persiste", err)
		retorno := ErrorResp{Mensagem: err.Error()}
		json.NewEncoder(w).Encode(retorno)
		return
	}

	//seta o status e retorno, e envia o response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(retorno)
}

func PersisteCotacao(cotacao *Dolar) error {

	db, err := ConexaoDb()
	if err != nil {
		return err
	}

	err = AddDolar(db, cotacao)
	if err != nil {
		return err
	}

	return nil
}

func AddDolar(db *gorm.DB, cotacao *Dolar) error {

	log.Println(cotacao)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	db.WithContext(ctx).Create(&cotacao)

	return nil
}

func ConexaoDb() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("./db/cotacao.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&Dolar{})

	return db, nil

}

func ExecutaChamadaRest() (*Cotacao, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var cot Cotacao
	err = json.Unmarshal(body, &cot)
	if err != nil {
		return nil, err
	}

	log.Println(cot)

	return &cot, nil
}

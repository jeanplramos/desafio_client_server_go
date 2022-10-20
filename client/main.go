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
	Valor float64 `json:"valor"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	body, error := io.ReadAll(res.Body)
	if error != nil {
		panic(error)
	}
	var cot CotacaoDolar
	error = json.Unmarshal(body, &cot)
	if error != nil {
		panic(error)
	}
	file, err := os.OpenFile("cotacao.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Erro ao criar arquivo: %v\n", err)
	}
	defer file.Close()
	_, err = file.WriteString(fmt.Sprintf("Dólar: %v\n", cot.Valor))
	if err != nil {
		panic(err)
	}
	fmt.Println("Arquivo criado com sucesso!")

}

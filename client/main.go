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

	if res.StatusCode != http.StatusOK {
		fmt.Printf("Erro ao chamar o serviço: %s\n", res.Status)
		return
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Erro ao processar retorno: %s\n", err)
		return
	}

	var cot CotacaoDolar
	err = json.Unmarshal(body, &cot)
	if err != nil {
		fmt.Printf("Erro ao converter Body: %s\n", err)
		return
	}
	file, err := os.OpenFile("cotacao.txt", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
	if err != nil {
		fmt.Printf("Erro ao criar arquivo: %v\n", err)
		return
	}
	defer file.Close()

	_, err = file.WriteString(fmt.Sprintf("Dólar: %v\n", cot.Valor))
	if err != nil {
		fmt.Printf("Erro ao escrever no arquivo: %v\n", err)
		return
	}

	fmt.Println("Arquivo criado com sucesso!")
}

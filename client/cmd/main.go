package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/LucianoGiope/posgolangdesafio1/pkg/httpResponseErr"
)

type DollarQuote struct {
	ID           string `json:"id"`
	Value        string `json:"value"`
	TypeCurrency string `json:"typecurrency"`
	CreatedAt    string `json:"created_at"`
}

func main() {

	fmt.Println("Iniciando a busca da cotação da moeda [USD-BRL]")

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*300)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("__Falha na requisição da cotação")
	}

	res, err := http.DefaultClient.Do(req)
	if ctx.Err() != nil {
		log.Fatal("__SERVER demorou para responder a Cotação. Tente novamente!! \n")
	}
	if err != nil {
		log.Fatalf("Erro ao chamar o SERVER\n__%v\n", err.Error())
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		var quoteErrorType httpResponseErr.QuoteError
		jsonBody, err := io.ReadAll(res.Body)
		if err != nil {
			log.Fatalf("\nErro ao ler corpo da mensagem de erro\n__%v\n", err.Error())
		}

		msgErro, err := quoteErrorType.DisplayMessage(jsonBody)
		if err != nil {
			log.Fatalf("Erro ao converter resposta.\n__[MESSAGE]%v\n", err.Error())
		}
		log.Fatalf("Falha durante a cotação\n%v\n", msgErro)

	} else {
		var dataResult DollarQuote
		jsonBody, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nErro ao ler a resposta Body:%v\n", err.Error())
		}

		err = json.Unmarshal(jsonBody, &dataResult)
		if err != nil {
			fmt.Fprintf(os.Stderr, "\nErro ao converter a resposta Body:%v\n", err.Error())
		}
		if err != nil {
			fmt.Println("Erro ao converter a data da cotação")
		}

		pathDocs := "../docs/"
		nameFile := "cotacao.txt"

		fmt.Printf("Atualizando dados no arquivo %s com valor de R$ %s \n", nameFile, dataResult.Value)

		file, err := os.Create(pathDocs + nameFile)
		if err != nil {
			fmt.Println("Falha ao criar o arquivo ", err)
		}
		defer file.Close()
		file.WriteString("Dólar: R$ " + dataResult.Value)

		fmt.Printf("Cotação do dolar atualizado em %s com valor de R$ %s \n", dataResult.CreatedAt, dataResult.Value)
	}
}

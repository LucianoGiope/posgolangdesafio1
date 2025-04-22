package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/LucianoGiope/posgolangdesafio1/pkg/httpResponseErr"
	"github.com/LucianoGiope/posgolangdesafio1/server/internal/entity"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func main() {
	println("\nIniciando o servidor na porta 8080")
	database, err := sql.Open("sqlite3", "../internal/sqlitebd/desafio1.db")
	if err != nil {
		panic(err)
	}
	db = database
	defer db.Close()

	routers := http.NewServeMux()

	routers.HandleFunc("/cotacao", searchDollarQuoteHandler)
	routers.HandleFunc("/listLast", listLastDollarQuoteHandler)
	routers.HandleFunc("/listAll", listAllDollarQuoteHandler)
	err = http.ListenAndServe(":8080", routers)
	if err != nil {
		log.Fatal(err)
	}

}

func searchDollarQuoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/cotacao" {
		println("The access must by in  of the endpoint http://localhost:8080/cotacao")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("The access must by in  of the endpoint http://localhost:8080/cotacao\n"))
		return
	}
	ctxClient := r.Context()
	timeAtual := time.Now()
	fmt.Printf("\n-> Search for dollar quote in %v.\n", timeAtual.Format("02/01/2006 15:04:05 ")+timeAtual.String()[20:29]+" ms")
	ctxSearch, cancelSearch := context.WithTimeout(ctxClient, time.Millisecond*200)
	defer cancelSearch()

	newDollar, err := GetQuotation(ctxSearch)
	if err != nil {
		msgErrFix := "__Error searching for dollar quote."
		println(msgErrFix, "\n____[MESSAGE]", err.Error())
		errCode := 0
		errText := ""
		if ctxSearch.Err() != nil {
			errCode = http.StatusRequestTimeout
			errText = msgErrFix + "\n____[MESSAGE] Tempo de pesquisa excedido"
		} else {
			errCode = http.StatusBadRequest
			errText = msgErrFix + "\n____[MESSAGE] Falha na requisição."
		}
		w.WriteHeader(errCode)
		msgErro := httpResponseErr.NewQuoteError(errText, errCode)
		json.NewEncoder(w).Encode(msgErro)
		return
	}

	ctxSave, cancelSave := context.WithTimeout(ctxSearch, time.Millisecond*15)
	defer cancelSave()

	fmt.Printf("\n-> Saving for dollar quote in %v.\n", time.Now().Format("02/01/2006 15:04:05"))
	err = saveDataDollar(ctxSave, newDollar)
	if err != nil {

		msgErrFix := "__Error saving for dollar quote in datatable."
		println(msgErrFix, "\n____[MESSAGE]", err.Error())

		errCode := 0
		errText := ""
		if ctxSave.Err() != nil {
			errCode = http.StatusRequestTimeout
			errText = msgErrFix + "\n____[MESSAGE] Excedido tempo de registro no banco"
		} else {
			errCode = http.StatusBadRequest
			errText = msgErrFix + "\n____[MESSAGE] Falha ao registrar dados no banco."
		}
		w.WriteHeader(errCode)
		msgErro := httpResponseErr.NewQuoteError(errText, errCode)
		json.NewEncoder(w).Encode(msgErro)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	res, err := json.Marshal(newDollar)
	if err != nil {
		println("__Erro ao converter o struct DollarQuote.", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Erro ao converter a estrutura de cotação. %v", err)))
		return
	}
	w.Write(res)
	fmt.Printf("\n-> Dollar quote type USD-BRL registred in database. Date %v.\n", time.Now().Format("02/01/2006 15:04:05"))

	fmt.Printf("\n-> Time total in milliseconds traveled %v.\n", time.Since(timeAtual))

	json.NewEncoder(w)
}

func GetQuotation(ctxSearch context.Context) (*entity.DollarQuote, error) {

	response, err := http.Get("https://economia.awesomeapi.com.br/json/last/USD-BRL")
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	bodyResp, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	var dataResult map[string]interface{}
	err = json.Unmarshal(bodyResp, &dataResult)
	if err != nil {
		return nil, err
	}
	var valueCurrency = dataResult["USDBRL"].(map[string]interface{})["bid"]

	newDollar, err := entity.NewDollarQuote(fmt.Sprint(valueCurrency))
	if err != nil {
		return nil, err
	}
	select {
	case <-ctxSearch.Done():
		return nil, ctxSearch.Err()
	default:
		return newDollar, nil
	}
}
func saveDataDollar(ctxSave context.Context, d *entity.DollarQuote) error {
	stmt, err := db.PrepareContext(ctxSave, "INSERT INTO dollarquote(id, value, createdat) VALUES (?,?,?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctxSave, d.ID, d.Value, d.CreatedAt)
	if err != nil {
		return err
	}

	select {
	case <-ctxSave.Done():
		err = rollBackQuote(d.ID)
		if err != nil {
			fmt.Printf("___ Erro no delete cotação ID:%s erro:%s \n", d.ID, err.Error())
			return err
		}
		fmt.Printf("___ Rollback da Cotação ID:%s\n", d.ID)
		return ctxSave.Err()
	default:
		return nil
	}
}

func rollBackQuote(id string) error {
	stmt, err := db.Prepare("DELETE FROM dollarquote where id=?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(id)
	if err != nil {
		return err
	}
	return nil
}

func listLastDollarQuoteHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select * from dollarquote order by createdat desc limit 1")
	if err != nil {
		log.Fatalf("Erro no prapare. Erro:%v\n", err.Error())
	}
	defer rows.Close()
	var dolar entity.DollarQuote
	rows.Next()
	err = rows.Scan(&dolar.ID, &dolar.Value, &dolar.CreatedAt)
	if err != nil {
		log.Fatalf("Erro no Scan. Erro:%v\n", err.Error())
	}
	fmt.Printf("\nÚltima cotação em %s R$ %v USD-BRL \n", dolar.CreatedAt, dolar.Value)
	_, err2 := w.Write([]byte(fmt.Sprintf("Última cotação em %s R$ %v USD-BRL \n", dolar.CreatedAt, dolar.Value)))
	if err2 != nil {
		fmt.Printf("Error printing return message:%s \n", err2)
	}
}
func listAllDollarQuoteHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("select * from dollarquote order by createdat asc")
	if err != nil {
		log.Fatalf("Erro no prapare. Erro:%v\n", err.Error())
	}
	defer rows.Close()

	fmt.Printf("Cotações realizadas até o momento\n")
	w.Write([]byte(fmt.Sprintf("Cotações realizadas até o momento\n")))
	var dolar entity.DollarQuote
	qtde := 1
	for rows.Next() {
		err = rows.Scan(&dolar.ID, &dolar.Value, &dolar.CreatedAt)
		if err != nil {
			log.Fatalf("Erro no Scan. Errp:%v\n", err.Error())
		}
		fmt.Printf("%v__Data:%s R$ %v USD-BRL\n", qtde, dolar.CreatedAt, dolar.Value)
		_, err2 := w.Write([]byte(fmt.Sprintf("%v__Data:%s R$ %v USD-BRL\n", qtde, dolar.CreatedAt, dolar.Value)))
		if err2 != nil {
			fmt.Printf("Error printing return message:%s \n", err2)
		}
		qtde++
	}
}

package main

import (
	"fmt"
	"encoding/json"
	"net/http"
	"net/url"
	"io"
	"os"
	"strings"
)

func Getauth() map[string]interface{} {
        tokenURL := "http://oauth.battle.net/token"
        clientID := os.Getenv("CLIENTID")
        clientSECRET := os.Getenv("CLIENTSECRET")

	if clientID == "" || clientSECRET == "" {
    		fmt.Println("Variáveis de ambiente CLIENTID ou CLIENTSECRET não estão definidas")
    		return nil
	}

        data := url.Values{}
        data.Set("grant_type","client_credentials")

        req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
        if err != nil {
                fmt.Println("Erro ao criar a requisição:",err)
                return nil
        }

        req.SetBasicAuth(clientID,clientSECRET)
        req.Header.Set("Content-Type","application/x-www-form-urlencoded")

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
                fmt.Println("Erro ao ler a resposta:",err)
                return nil
        }

		defer resp.Body.Close()
		
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Println("Erro ao ler a resposta:",err)
		}

        if resp.StatusCode == http.StatusOK {
				var result map[string]interface{}
                json.Unmarshal(body, &result)
				return result
        } else{
            return nil
        }
}

func main() {
	authTOK := Getauth()
	fmt.Println("Token: ",authTOK)
	return
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Auction struct {
	ID        int    `json:"id"`
	Item      Item   `json:"item"`
	Quantity  int    `json:"quantity"`
	UnitPrice int    `json:"unit_price"`
	TimeLeft  string `json:"time_left"`
}

type Item struct {
	ID int `json:"id"`
}

type AuctionsData struct {
	Auctions []Auction `json:"auctions"`
}

var (
	authToken   string    	
  tokenExpiry time.Time
)

func getAuth() (string, time.Time) {
	clientID := os.Getenv("CLIENTID")
	clientSecret := os.Getenv("CLIENTSECRET")
	postURL := "https://oauth.battle.net/token"

	if clientID == "" || clientSecret == "" {
		fmt.Println("CLIENTID ou CLIENTSECRET não configurados.")
		os.Exit(1)
	}

	data := bytes.NewBufferString("grant_type=client_credentials")
	request, err := http.NewRequest("POST", postURL, data)
	if err != nil {
		panic(err)
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.SetBasicAuth(clientID, clientSecret)

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var respMap map[string]interface{}
	json.Unmarshal(bodyResp, &respMap)

	token := respMap["access_token"].(string)
	expiresIn := respMap["expires_in"].(float64)
	expiryTime := time.Now().Add(time.Duration(expiresIn) * time.Second)

	return token, expiryTime
}

func ensureAuthToken() string {
	if authToken == "" || time.Now().After(tokenExpiry) {
		authToken, tokenExpiry = getAuth()
		fmt.Println("Token renovado.")
	}
	return authToken
}

func sendRequest(authToken string) AuctionsData {
	getURL := "https://us.api.blizzard.com/data/wow/auctions/commodities?namespace=dynamic-us&locale=en_US"
	request, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		panic(err)
	}

	authHeader := fmt.Sprintf("Bearer %s", authToken)
	request.Header.Set("Authorization", authHeader)

	client := http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	bodyResp, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data AuctionsData
	json.Unmarshal(bodyResp, &data)
	return data
}

func main() {
	botToken := os.Getenv("BOTTOK")
	groupID := os.Getenv("GROUPID")

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		panic("Erro ao obter bot: " + err.Error())
	}

	groupIDInt, err := strconv.Atoi(groupID)
	if err != nil {
		panic("GROUPID deve ser um número inteiro.")
	}

	for {
		authToken := ensureAuthToken()

		auctionData := sendRequest(authToken)

		minPrice := math.MaxInt
		for _, auction := range auctionData.Auctions {
			if auction.Item.ID == 210810 && auction.UnitPrice < minPrice {
				minPrice = auction.UnitPrice
			}
		}

		if minPrice > 910000 {
			notification := fmt.Sprintf("O preço de venda da lança de Arathor 3 estrelas é %d", minPrice)
			msg := tgbotapi.NewMessage(int64(groupIDInt), notification)
			_, err = bot.Send(msg)
			if err != nil {
				fmt.Println("Erro no envio da mensagem:", err)
			}
		}

		time.Sleep(time.Hour)
	}
}


package test

import (
	"fmt"
	"math/rand"
	"time"
	"encoding/json"
)

type HistoryResponse struct {
	Wallet  string
	Amount  float32
	Message string
	Time    string
}

func main() {
	response := []HistoryResponse{}
	for i := 0; i < 10; i++ {
		response = append(response, MockHistoryRow())
	}
	jsonStr, err := json.Marshal(response)
	if err != nil {
		fmt.Print("Error marshaling")
	}
	fmt.Printf("Value %s", jsonStr)
}
var letterRunes = []rune("0123456789ABCDEF")

func randStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func randomTimestamp() time.Time {
	randomTime := rand.Int63n(time.Now().Unix()-94608000) + 94608000
	randomNow := time.Unix(randomTime, 0)
	return randomNow
}

func MockHistoryRow() HistoryResponse {
	wallet := randStringRunes(16)
	amount := rand.Float32()*200 - 100
	message := "INCOME"
	if amount < 0 {
		message = "SPENT"
	}
	timeStamp := randomTimestamp().Format("15.01.2006 15:04:05")

	return HistoryResponse{
		Wallet:  wallet,
		Amount:  amount,
		Message: message,
		Time:    timeStamp,
	}
}

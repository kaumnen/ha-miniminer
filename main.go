package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type ChallengeData struct {
	Difficulty int   `json:"difficulty"`
	Block      Block `json:"block"`
}

type Block struct {
	Nonce int             `json:"nonce"`
	Data  [][]interface{} `json:"data"`
}

func main() {
	cd := getChallengeData()

	solutionNonce := cd.getNonce()

	submitNonce(solutionNonce)
}

func getChallengeData() ChallengeData {
	response, err := http.Get(constructEndpoint("problem"))

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	var challengeData ChallengeData

	json_err := json.Unmarshal(responseData, &challengeData)

	if json_err != nil {
		fmt.Print(json_err.Error())
		os.Exit(1)
	}

	fmt.Printf("Difficulty: %d\n", challengeData.Difficulty)
	fmt.Printf("Block Data: %v\n", challengeData.Block.Data)

	return challengeData
}

func (cd ChallengeData) getNonce() int {
	nonceIterator := 0

	for {
		fmt.Printf("-------------------------\nTrying nonce: %d\n", nonceIterator)

		payload := constructPayload(cd.Block.Data, nonceIterator)
		fmt.Println("Payload: ", payload)

		shaBytes, iterationHash := hashPayload(payload)

		fmt.Printf("Iteration Hash: %s\n", iterationHash)

		nonceOK := CheckBitHash(shaBytes, cd.Difficulty)

		if nonceOK {
			fmt.Printf("Nonce found! %d\n", nonceIterator)
			break
		} else {
			fmt.Println("Nonce not found! Continuing...")
			nonceIterator++
		}
	}

	return nonceIterator
}

func submitNonce(nonce int) {
	solutionBody := fmt.Sprintf(`{"nonce":%d}`, nonce)
	response, err := http.Post(constructEndpoint("solve"), "application/json", bytes.NewBuffer([]byte(solutionBody)))

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	defer response.Body.Close()

	responseData, err := io.ReadAll(response.Body)

	if err != nil {
		fmt.Print(err.Error())
		os.Exit(1)
	}

	fmt.Printf("Response body: %s\n", responseData)

	if response.StatusCode == 200 {
		fmt.Println("Solved!")
	} else {
		fmt.Println("There was a problem with submission. Try again!")
	}
}

func constructEndpoint(phase string) string {
	ha_domain := os.Getenv("HA_DOMAIN")
	ha_token := os.Getenv("HA_TOKEN")

	return fmt.Sprintf("%s/challenges/mini_miner/%s?access_token=%s", ha_domain, phase, ha_token)
}

func constructPayload(data [][]interface{}, nonce int) string {
	var dataString string
	var payload string = `{"data":[`

	for i, d := range data {
		if i == len(data)-1 {
			dataString = fmt.Sprintf(`["%s",%v]`, d[0], d[1])
		} else {
			dataString = fmt.Sprintf(`["%s",%v],`, d[0], d[1])
		}
		payload += dataString
	}

	payload += fmt.Sprintf(`],"nonce":%d}`, nonce)

	return payload
}

func hashPayload(payload string) ([32]byte, string) {
	shaBytes := sha256.Sum256([]byte(payload))
	hash := hex.EncodeToString(shaBytes[:])

	return shaBytes, hash
}

func CheckBitHash(hash [32]byte, difficulty int) bool {
	bits := ""
	for _, b := range hash {
		bits += fmt.Sprintf("%08b", b)
	}

	for i := 0; i < difficulty; i++ {
		fmt.Printf("Bit %d: %s\n", i, string(bits[i]))
		if bits[i] != '0' {
			return false
		}
	}
	return true
}

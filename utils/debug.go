package utils

import (
	"encoding/json"
	"fmt"
	"log"
)

func PrintJsonLog(v interface{}, label string) {
	fmt.Printf("============= %s ==================", label)
	prevTXsJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(prevTXsJSON))
}

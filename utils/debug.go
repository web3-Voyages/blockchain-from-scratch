package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

func PrintJsonLog(v interface{}, label string) {
	fmt.Printf("%s: ==> %s\n", time.Now().Format("2006-01-02 15:04:05.000"), label)
	prevTXsJSON, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(prevTXsJSON))
}

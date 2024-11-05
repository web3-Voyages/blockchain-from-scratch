package utils

import (
	"bytes"
	"encoding/gob"
	"log"
)

func Serialize(v interface{}) []byte {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(v)
	if err != nil {
		log.Panic(err)
	}
	return result.Bytes()
}

func Deserialize(data []byte, v interface{}) {
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(v)
	if err != nil {
		log.Panic(err)
	}
}
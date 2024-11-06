package utils

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"log"
)

const commandLength = 12

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

func CommandToBytes(command string) []byte {
	var bytes [commandLength]byte

	for i, c := range command {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

func BytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

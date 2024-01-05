package utils

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"io"
	"os"
)

type Encoder[T any] func(T) ([]byte, error)

type Decoder[T any] func([]byte) (T, error)

func B64JsonEncoder[T any](value T) (data []byte, err error) {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return data, err
	}
	encoder := base64.RawStdEncoding
	data = make([]byte, encoder.EncodedLen(len(jsonData)))
	encoder.Encode(data, jsonData)

	return data, err
}

func B64JsonDecoder[T any](data []byte) (value T, err error) {
	encoding := base64.RawStdEncoding
	jsonValue := make([]byte, encoding.DecodedLen(len(data)))
	_, err = encoding.Decode(jsonValue, data)
	if err != nil {
		return value, err
	}
	err = json.Unmarshal(jsonValue, &value)
	if err != nil {
		return value, err
	}
	return value, err
}

func Append[T any](file *os.File, value T, encoder Encoder[T]) error {
	encoded, err := encoder(value)
	if err != nil {
		return err
	}

	buffLenBytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(buffLenBytes, uint32(len(encoded)))
	if _, err := file.Write(buffLenBytes); err != nil {
		return err
	}
	_, err = file.Write(encoded)
	return err
}

// liest itmes aus einer Datei LEN:ITEM
func LoadFromFile(path string, consumer func(value []byte) error) (count int, err error) {
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	for {
		lenBytes := make([]byte, 4)
		if _, err := io.ReadFull(file, lenBytes); err != nil {
			if err == io.EOF {
				return count, nil
			}
			return count, err
		}
		dataLen := binary.LittleEndian.Uint32(lenBytes)
		dataBuffer := make([]byte, int(dataLen))
		if _, err := io.ReadFull(file, dataBuffer); err != nil {
			return count, err
		}
		if err = consumer(dataBuffer); err != nil {
			return count, err
		} else {
			count = count + 1
		}
	}
}

func CreateFileIfNotExists(path string) error {
	if !FileExist(path) {
		_, err := os.Create(path)
		return err
	}
	return nil
}

func FileExist(path string) bool {
	_, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return false
	} else {
		return true
	}
}

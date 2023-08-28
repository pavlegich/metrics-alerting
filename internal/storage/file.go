package storage

import (
	"encoding/json"
	"os"

	"github.com/pavlegich/metrics-alerting/internal/interfaces"
)

type FileWriter struct {
	file    *os.File
	encoder *json.Encoder
}

func NewFileWriter(path string) (*FileWriter, error) {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		return nil, err
	}

	return &FileWriter{file: file, encoder: json.NewEncoder(file)}, nil
}

func (fw *FileWriter) Close() error {
	return fw.file.Close()
}

func (fw *FileWriter) WriteMemStorage(m *interfaces.MetricStorage) error {
	return fw.encoder.Encode(&m)
}

type FileReader struct {
	file    *os.File
	decoder *json.Decoder
}

func NewFileReader(path string) (*FileReader, error) {
	file, err := os.OpenFile(path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	return &FileReader{
		file:    file,
		decoder: json.NewDecoder(file),
	}, nil
}

func (fr *FileReader) Close() error {
	return fr.file.Close()
}

func (fr *FileReader) ReadMemStorage() (*MemStorage, error) {
	storage := &MemStorage{}
	if err := fr.decoder.Decode(&storage); err != nil {
		return nil, err
	}

	return storage, nil
}

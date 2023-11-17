// Пакет compress содержит объекты и методы, реализующие сжатие данных.
package compress

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
)

// compressWriter реализует интерфейс http.ResponseWriter и позволяет прозрачно для сервера
// сжимать передаваемые данные и выставлять правильные HTTP-заголовки.
type compressWriter struct {
	w  http.ResponseWriter
	zw *gzip.Writer
}

// NewCompressWriter создаёт новый compressWriter.
func NewCompressWriter(w http.ResponseWriter) *compressWriter {
	return &compressWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}

// Header реализует получение заголовка ответа.
func (c *compressWriter) Header() http.Header {
	return c.w.Header()
}

// Write реализует запись данных в ответ.
func (c *compressWriter) Write(p []byte) (int, error) {
	return c.zw.Write(p)
}

// Write формирует заголовок ответа.
func (c *compressWriter) WriteHeader(statusCode int) {
	c.w.Header().Set("Content-Encoding", "gzip")
	c.w.WriteHeader(statusCode)
}

// Close закрывает gzip.Writer и досылает все данные из буфера.
func (c *compressWriter) Close() error {
	return c.zw.Close()
}

// compressReader реализует интерфейс io.ReadCloser и позволяет прозрачно для сервера
// декомпрессировать получаемые от клиента данные.
type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

// NewCompressReader создаёт новый compressReader.
func NewCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("NewCompressReader: gzip new reader %w", err)
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

// Read реализует io.Reader для чтения несжатых данных из gzip.Reader.
func (c compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

// Close закрывает gzip.Reader.
func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return fmt.Errorf("Close: reader close error %w", err)
	}
	return c.zr.Close()
}

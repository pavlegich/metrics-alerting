package models

import (
	"errors"
	"strconv"
	"strings"
)

type (
	Address struct {
		Host string
		Port int
	}
)

func (a *Address) Set(value string) error {
	values := strings.Split(value, ":")
	if len(values) != 2 || values[0] == "" {
		return errors.New("need address in a form host:port")
	}
	port, err := strconv.Atoi(values[1])
	if err != nil {
		return err
	}
	a.Host = values[0]
	a.Port = port
	return nil
}

func (a *Address) String() string {
	return a.Host + ":" + strconv.Itoa(a.Port)
}

func NewAddress() *Address {
	return &Address{Host: "localhost", Port: 8080}
}

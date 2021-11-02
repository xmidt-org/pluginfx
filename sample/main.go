package main

import "errors"

var Value int = 12

func New() (float64, error) {
	return 67.5, nil
}

func AlwaysErrors() (string, error) {
	return "", errors.New("expected sample plugin constructor error")
}

func Initialize() {
}

func Shutdown() {
}

func main() {
}

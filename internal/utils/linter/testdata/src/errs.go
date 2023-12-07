package main

import "os"

func main() {
	os.Exit(1) // want "os.Exit calls are prohibited in main()"
}

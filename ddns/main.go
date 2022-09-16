package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println(os.Getenv("ENV"))
	fmt.Println("It works")
	fmt.Println(config.Hosts)
}

package main

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	fmt.Println("\nPrint text: ")
	reader := bufio.NewReader(os.Stdin)
	text, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error")
	}
	for i := 0; i <= len(text); i++ {
		fmt.Println(text[i])
	}
}


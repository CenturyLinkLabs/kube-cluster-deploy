package utils

import "fmt"

func LogInfo(message string) {
	fmt.Println(message)
}

func LogErrorMessage(message string) {
	fmt.Println(message)
}

func LogError(err error) {
	fmt.Println(err)
}

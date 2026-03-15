package main

import (
	"fmt"
	"os"
	"proxy/internal/application"
	"proxy/internal/storage/users"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: ./create-user <username> <password>")
		fmt.Println("Example: ./create-user alice alice123")
		os.Exit(1)
	}

	username := os.Args[1]
	password := os.Args[2]

	app := application.NewApp()

	// Проверяем, не существует ли уже пользователь
	foundedUser, _ := app.UserStorage.FindByUsername(username)
	if foundedUser != nil {
		fmt.Printf("Error: User %s already exists", username)
		os.Exit(1)
	}

	// Создаём пользователя
	newUser := users.User{
		Username: username,
		Password: password,
	}

	// Добавляем нового пользователя
	err := app.UserStorage.Add(newUser)
	if err != nil {
		fmt.Printf("Failed to add user: %v", err)
		os.Exit(1)
	}

	fmt.Printf("User %s created", username)
}

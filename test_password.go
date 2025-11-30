package main

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	dbHash := "$2a$10$Zy7cS3ECLenBsqkYqYFJAuPpYV9zp1KimGJhIVLG0zGjAIK5tD9i."
	password := "password123"

	err := bcrypt.CompareHashAndPassword([]byte(dbHash), []byte(password))

	if err == nil {
		fmt.Println("✅ Password matches!")
	} else {
		fmt.Println("❌ Password does NOT match!")
		correctHash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		fmt.Printf("\nUpdate SQL:\n")
		fmt.Printf("UPDATE users SET password_hash = '%s' WHERE email = 'smith.robert@example.com';\n", string(correctHash))
	}
}

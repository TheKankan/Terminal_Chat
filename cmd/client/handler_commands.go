package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/TheKankan/TerminalSecuredChat/internal/validator"

	_ "github.com/lib/pq"
)

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (cfg *config) handlerLogin() *User {
	reader := bufio.NewReader(os.Stdin)
	url := "http://" + cfg.addr + "/login"

	// Get the username and check if it is valid
	var username string
	for {
		fmt.Print("Username: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if validator.IsValidUsername(input) {
			username = input
			break
		}

		fmt.Print("❌ Invalid username (no spaces, max 20 chars)\n\n")
	}

	// Get the password and check if it is valid
	var password string
	for {
		fmt.Print("Password: ")
		input, _ := reader.ReadString('\n')

		if validator.IsValidPassword(input) {
			password = input
			break
		}

		fmt.Print("❌ Invalid password (min 4 chars, max 50 chars)\n\n")
	}

	credentials := Credentials{
		Username: username,
		Password: password,
	}
	resp, err := sendJSON(url, credentials)
	if err != nil {
		fmt.Printf("Failed to send credentials: %s \n\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			fmt.Printf("Server Error: %s\n\n", errResp.Error)
		}
		return nil
	}

	var userResp struct {
		User  User   `json:"user"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		fmt.Printf("Failed to decode user: %s\n", err)
		return nil
	}
	cfg.token = userResp.Token

	fmt.Printf("Logged in as: %s\n", userResp.User.Username)
	fmt.Printf("Your token: %s\n\n", userResp.Token)
	return &userResp.User

}

func (cfg *config) handlerRegister() *User {
	url := "http://" + cfg.addr + "/register"

	reader := bufio.NewReader(os.Stdin)

	// Get the username and check if it is valid
	var username string
	for {
		fmt.Print("Username: ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if validator.IsValidUsername(input) {
			username = input
			break
		}

		fmt.Print("❌ Invalid username (no spaces, max 20 chars)\n\n")
	}

	// Get the password and check if it is valid
	var password string
	for {
		fmt.Print("Password: ")
		input, _ := reader.ReadString('\n')

		if validator.IsValidPassword(input) {
			password = input
			break
		}

		fmt.Print("❌ Invalid password (min 4 chars, max 50 chars)\n\n")
	}

	fmt.Printf("username : %s, password : %s\n", username, password)

	credentials := Credentials{
		Username: username,
		Password: password,
	}

	resp, err := sendJSON(url, credentials)
	if err != nil {
		fmt.Printf("Failed to send credentials: %s \n\n", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		var errResp ErrorResponse
		if json.Unmarshal(body, &errResp) == nil && errResp.Error != "" {
			fmt.Printf("Server Error: %s\n\n", errResp.Error)
		}
		return nil
	}

	var userResp struct {
		User  User   `json:"user"`
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		fmt.Printf("Failed to decode user: %s\n", err)
		return nil
	}
	cfg.token = userResp.Token

	fmt.Println("✅ Registration successful!")
	return &userResp.User
}

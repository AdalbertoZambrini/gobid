package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	migrationsPath := filepath.Join(wd, "internal", "store", "pgstore", "migrations")
	configPath := filepath.Join(wd, "internal", "store", "pgstore", "migrations", "tern.conf")

	fmt.Println("Running migrations from:", migrationsPath)

	cmd := exec.Command(
		"tern",
		"migrate",
		"--migrations", migrationsPath,
		"--config", configPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("Command failed. Working Dir: %s\n", wd)
		fmt.Println("Output:", string(output))
		panic(err)
	}
	fmt.Println("Success:", string(output))
}

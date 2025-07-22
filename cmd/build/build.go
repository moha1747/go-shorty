package main

import (
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "build", "-o", "bin/go-shorty", ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		panic(err)
	}
}

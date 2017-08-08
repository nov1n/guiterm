package main

import (
	"os"
	"os/exec"

	gh "github.com/nov1n/guitarhero/game"
)

var (
	game   *gh.Game
	points = 0
)

func main() {
	game = gh.NewGame()
	game.Initialize()
	go captureInput()
	game.Loop()
}

func captureInput() {
	// Disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()

	// Disable echo'ing
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	// Read byte
	var b []byte = make([]byte, 1)
	for {
		n, err := os.Stdin.Read(b)
		if err != nil {
			panic(err)
		}
		if n > 0 {
			game.KeyPressed(string(b))
		}
	}

	// Eventually reenable echoing
	defer exec.Command("stty", "-F", "/dev/tty", "echo").Run()
}

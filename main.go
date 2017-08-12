package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/nov1n/guitarhero/colors"
	gh "github.com/nov1n/guitarhero/game"
)

var game *gh.Game

func main() {
	// Cleanup when quit
	sigc := make(chan os.Signal)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		sig := <-sigc
		fmt.Printf("Caught signal %s: shutting down.\n", sig)

		cleanup()
		os.Exit(0)
	}()
	// Get name
	fmt.Print("Enter your name: ")
	r := bufio.NewReader(os.Stdin)
	name, err := r.ReadString('\n')
	if err != nil {
		panic(err)
	}
	name = strings.TrimSpace(name)

	// Get terminal dimensions
	w, h, err := getTerminalDims()
	if err != nil {
		panic(err)
	}

	// Create and start the game
	game = gh.New(name, w, h)
	game.Initialize()
	go captureInput()
	game.Loop()

	// Cleanup upon exit
	cleanup()
}

func captureInput() {
	// Disable input buffering
	exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()

	// Disable echo'ing
	exec.Command("stty", "-F", "/dev/tty", "-echo").Run()

	// Disable blinking
	fmt.Print("\033[?25l")

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
}

func getTerminalDims() (h int, w int, err error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin // stty uses ioctl on stdin filedescriptor to ask kernel for terminal size, supply parent's stdin to get correct size
	b, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	_, err = fmt.Sscanf(string(b), "%d %d\n", &h, &w)
	if err != nil {
		return 0, 0, err
	}
	return w, h, err
}

func cleanup() {
	// Eventually reenable echoing and blinking
	gh.Clear()

	// Reset terminfo
	fmt.Print(colors.Color("", colors.Normal))

	fmt.Print("\033[?25h")
	exec.Command("stty", "-F", "/dev/tty", "+echo").Run()
}

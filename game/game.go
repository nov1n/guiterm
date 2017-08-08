package game

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	defaultFps = 6
	barIndex   = 20
	halfSymbol = "/"
	fullSymbol = "*"
)

type Game struct {
	fps    int
	width  int
	height int
	screen []string
	score  int
	keys   []string
}

func NewGame() *Game {
	tw, th, err := getTerminalDims()
	if err != nil {
		panic(err)
	}

	sh := th - 1 // allow 2 lines for user input

	return &Game{
		fps:    defaultFps,
		width:  tw,
		height: sh,
		screen: []string{},
		score:  0,
		keys:   []string{"a", "s", "d", "f"},
	}
}

func (g *Game) clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (g *Game) addScore(n int) {
	g.score += n
}

func (g *Game) KeyPressed(k string) {
	g.updateScore(k)
}

func (g *Game) updateScore(k string) {
	full := g.screen[barIndex]
	belowHalf := g.screen[barIndex-1]
	aboveHalf := g.screen[barIndex+1]

	hit := false

	// Full
	if strings.Contains(full, k) {
		hit = true
		// Mark the note as hit
		g.changeLine(barIndex, strings.Replace(g.screen[barIndex], k, fullSymbol, 1))

		g.addScore(100)

		// Remove k from the half point strings to prevent double count
		belowHalf = strings.Replace(belowHalf, k, " ", 1)
		aboveHalf = strings.Replace(aboveHalf, k, " ", 1)
	}

	// Below
	if strings.Contains(belowHalf, k) {
		hit = true
		// Mark the note as hit
		g.changeLine(barIndex-1, strings.Replace(g.screen[barIndex-1], k, halfSymbol, 1))

		// Remove k from next half point string to prevent double count
		aboveHalf = strings.Replace(aboveHalf, k, " ", 1)

		g.addScore(50)
	}

	// Above
	if strings.Contains(aboveHalf, k) {
		hit = true
		// Mark the note as hit
		g.changeLine(barIndex+1, strings.Replace(g.screen[barIndex+1], k, halfSymbol, 1))

		g.addScore(50)
	}

	// Miss
	if !hit {
		g.changeLine(barIndex, strings.Replace(g.screen[barIndex], " ", "-", -1))

		g.addScore(-50)
	}
}

func (g *Game) Initialize() {
	g.clear()
	for i := 0; i < g.height; i++ {
		g.appendFret()
	}
}

func (g *Game) appendFret() {
	line := ""
	for i := 0; i < len(g.keys); i++ {
		line += fmt.Sprint("|   ")
	}
	line += fmt.Sprint("|")
	g.appendLine(line)
}

func (g *Game) appendRandomNote() {
	line := ""

	keyIdx := rand.Intn(len(g.keys))
	key := g.keys[keyIdx]
	for j := 0; j < len(g.keys); j++ {
		curKey := " "
		if j == keyIdx {
			curKey = key
		}
		line += fmt.Sprintf("| %s ", curKey)
	}
	line += fmt.Sprint("|")

	g.appendLine(line)
}

func (g *Game) changeLine(n int, l string) {
	g.screen[n] = l
}

func (g *Game) appendLine(l string) {
	g.screen = append(g.screen, l)
}

func (g *Game) render() {
	g.trim()

	for i := g.height - 1; i >= 0; i-- {
		line := g.screen[i]
		if i == barIndex+1 || i == barIndex-1 { // Draw the horizontal lines
			line = strings.Replace(g.screen[i], " ", "-", -1)
		}

		var score string
		if i == barIndex {
			score = fmt.Sprintf("score: %d", g.score)
		}
		fmt.Printf("%s % 10d   %s\n", line, i, score)
	}
	fmt.Println("") // Gives one extra whitespace at the bottom
}

func (g *Game) trim() {
	nScreen := len(g.screen)
	if nScreen > g.height {
		g.screen = g.screen[len(g.screen)-g.height:]
	}
}

func getTerminalDims() (int, int, error) {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin // stty uses ioctl on stdin filedescriptor to ask kernel for terminal size, supply parent's stdin to get correct size
	b, err := cmd.Output()
	if err != nil {
		return 0, 0, err
	}
	res := strings.Split(strings.TrimSpace(string(b)), " ")

	w, err := strconv.Atoi(res[1])
	if err != nil {
		return 0, 0, err
	}

	h, err := strconv.Atoi(res[0])
	if err != nil {
		return 0, 0, err
	}

	return w, h, err
}

func (g *Game) Loop() {
	t := time.Tick(time.Duration(1000/g.fps) * time.Millisecond)
	r := time.After(30 * time.Second)

	for {
		select {
		case <-t:
			if rand.Intn(2) == 0 {
				g.appendRandomNote()
			} else {
				g.appendFret()
			}
			g.render()
			break
		case <-r:
			fmt.Printf("Congratulations, your score was%d\n", g.score)
			os.Exit(0)
		}
	}

}

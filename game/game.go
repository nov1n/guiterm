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
	horizontalBarIndex = 4
)

type Game struct {
	width  int
	height int
	screen []string
	points int
	keys   []string
}

func NewGame() *Game {
	tw, th, err := getTerminalDims()
	if err != nil {
		panic(err)
	}

	sh := th - 1 // allow 2 lines for user input

	return &Game{
		width:  tw,
		height: sh,
		screen: []string{},
		points: 0,
		keys:   []string{"a", "s", "d", "f"},
	}
}

func (g *Game) clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (g *Game) addPoints(n int) {
	g.points += n
}

func (g *Game) KeyPressed(k string) {
	line := g.screen[horizontalBarIndex]
	if strings.Contains(line, k) {
		g.addPoints(100)
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
		if i == 3 || i == 5 { // Draw the horizontal lines
			line = strings.Replace(g.screen[i], " ", "-", -1)
		}

		var points string
		if i == horizontalBarIndex {
			points = fmt.Sprintf("points: %d", g.points)
		}
		fmt.Printf("%s % 10d   %s\n", line, i, points)
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
	t := time.Tick(250 * time.Millisecond)

	for {
		select {
		case <-t:
			if rand.Intn(2) == 0 {
				g.appendRandomNote()
			} else {
				g.appendFret()
			}
			g.render()
		}
	}

}

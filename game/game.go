package game

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	highscores "github.com/nov1n/guitarhero/highscores"
	stats "github.com/nov1n/guitarhero/stats"
)

var (
	defaultFps     = 6
	debugIndex     = 0
	missIndex      = barIndex - 2
	timeIndex      = 2
	barIndex       = 3
	scoreIndex     = 3
	streakIndex    = 4
	flameIndex     = 5
	shortcutsIndex = 38
	roundLength    = 30 * time.Second
	debugString    = ""
	defaultKeys    = []string{"j", "k", "l", ";"}
	halfSymbol     = ">"
	fullSymbol     = "v"
	flame          = `
  )
 ) \
/ ) (
\(_)/
`
)

type Game struct {
	name             string
	fps              int
	timeLeft         time.Duration
	width            int
	height           int
	screen           []string
	keys             []string
	highscores       *highscores.Highscores
	stats            *stats.Stats
	stopChan         chan int
	restartChan      chan int
	pauseUnpauseChan chan int
	paused           bool
}

func New(n string) *Game {
	rand.Seed(time.Now().UTC().UnixNano())
	tw, th, err := getTerminalDims()
	if err != nil {
		panic(err)
	}

	sh := th - 1 // allow 2 lines for user input

	return &Game{
		name:             n,
		fps:              defaultFps,
		timeLeft:         roundLength,
		width:            tw,
		height:           sh,
		screen:           []string{},
		keys:             defaultKeys,
		stats:            stats.New(),
		highscores:       highscores.New(),
		stopChan:         make(chan int, 1),
		restartChan:      make(chan int, 1),
		pauseUnpauseChan: make(chan int, 1),
		paused:           false,
	}
}

func (g *Game) clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (g *Game) KeyPressed(k string) {
	switch k {
	case "q":
		g.Stop()
		return
	case "r":
		g.Restart()
		return
	case "p":
		g.PauseUnpause()
		return
	}

	if g.paused {
		return
	}

	g.updateScore(k)

	// Rerender to skip frame logic
	g.rerenderFrame()
}

func (g *Game) updateScore(k string) {
	full := g.screen[barIndex]
	belowHalf := g.screen[barIndex-1]
	aboveHalf := g.screen[barIndex+1]

	// Check wrong key and invalid key
	wrongKey := !strings.Contains((full + belowHalf + aboveHalf), k)
	invalidKey := !strings.Contains(strings.Join(g.keys, ""), k)
	if invalidKey || wrongKey {
		g.stats.Incorrect()
		return
	}

	// From here on the note was correct
	g.stats.Correct()

	// Check half below
	if strings.Contains(belowHalf, k) {
		// Mark the note as hit
		g.changeLine(barIndex-1, strings.Replace(g.screen[barIndex-1], k, halfSymbol, 1))

		// Remove k from next half point string to prevent double count
		aboveHalf = strings.Replace(aboveHalf, k, "", 1)
		full = strings.Replace(full, k, "", 1)

		g.stats.Add(stats.Half)
	}

	// Check full
	if strings.Contains(full, k) {
		// Mark the note as hit
		g.changeLine(barIndex, strings.Replace(g.screen[barIndex], k, fullSymbol, 1))

		g.stats.Add(stats.Full)

		// Remove k from the half point strings to prevent double count
		aboveHalf = strings.Replace(aboveHalf, k, "", 1)
	}

	// Check half above
	if strings.Contains(aboveHalf, k) {
		// Mark the note as hit
		g.changeLine(barIndex+1, strings.Replace(g.screen[barIndex+1], k, halfSymbol, 1))

		g.stats.Add(stats.Half)
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

// nextFrame advances the frame which means all of the frame logic is done
// and the new frame is drawn based on the internal representation
func (g *Game) advanceFrame() {
	g.trim()
	g.frameLogic()
	g.render()
}

// rerender the same frame for instant update, may be called more frequent
// than fps. Skips frame logic to prevent duplicate results.
func (g *Game) rerenderFrame() {
	g.trim()
	g.render()
}

func (g *Game) render() {
	for i := g.height - 1; i >= 0; i-- {
		line := g.screen[i]
		if i == barIndex+1 || i == barIndex-1 { // Draw the horizontal lines
			line = strings.Replace(g.screen[i], " ", "-", -1)
		}

		var sidebar string
		if i == scoreIndex {
			sidebar = fmt.Sprintf("score: %d (%d%%) %d", g.stats.Score, g.stats.Accuracy(), g.stats.LastNote)
		}
		if i == timeIndex {
			sidebar = fmt.Sprintf("time: %d", int(g.timeLeft.Seconds()))
		}
		if i == streakIndex {
			sidebar = fmt.Sprintf("%d (%dx)", g.stats.Streak, g.stats.Multiplier())
		}
		if i == shortcutsIndex {
			sidebar = "Shortcuts:"
		}
		if i == shortcutsIndex-1 {
			sidebar = fmt.Sprintf("(r) restart")
		}
		if i == shortcutsIndex-2 {
			sidebar = fmt.Sprintf("(q) quit")
		}
		if i == shortcutsIndex-3 {
			sidebar = fmt.Sprintf("(p) pause/resume")
		}
		if i == debugIndex {
			sidebar = fmt.Sprintf("%s", debugString)
		}

		// Draw flame
		f := strings.Split(flame, "\n")
		for j := 0; j < len(f) && g.stats.Multiplier() > 1; j++ {
			if i == (j + flameIndex) {
				sidebar = f[len(f)-j-1]
			}
		}

		//fmt.Printf("%s % 10d   %s\n", line, i, sidebar)
		fmt.Printf("%s  %s\n", line, sidebar)
	}
	fmt.Println("") // Gives one extra whitespace at the bottom
}

// Called once before every frame is rendered
func (g *Game) frameLogic() {
	// Check misses
	missLine := g.screen[missIndex]
	miss := strings.ContainsAny(missLine, strings.Join(g.keys, ""))
	if miss {
		g.stats.Incorrect()
	}

	// Count total notes
	if strings.ContainsAny(missLine, strings.Join(g.keys, "")) {
		g.stats.TotalNotesAdd(1)
	}
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

func (g *Game) Stop() {
	g.stopChan <- 1
}

func (g *Game) Restart() {
	g.restartChan <- 1
}

func (g *Game) PauseUnpause() {
	if !g.Finished() { // Game is not finished (which also uses pause = true for now)
		g.paused = !g.paused
		g.pauseUnpauseChan <- 1
	}
}

func (g *Game) Loop() {
	frameLength := time.Duration(1000/g.fps) * time.Millisecond
	t := time.Tick(frameLength)

	for {
		select {
		case <-t:
			if rand.Intn(2) == 0 {
				g.appendRandomNote()
			} else {
				g.appendFret()
			}

			g.timeLeft -= frameLength

			g.advanceFrame()

			if g.Finshed() {
				t = nil // Stop ticking
				g.paused = true
				g.showFinalScreen()
			}
			break
		case <-g.pauseUnpauseChan:
			if t == nil {
				t = time.Tick(frameLength)
			} else {
				t = nil
			}
		case <-g.stopChan:
			return
		case <-g.restartChan:
			*g = *New(g.name) // Reassign main's reference to game to a new one
			g.Initialize()
			g.Loop()
			return
		}
	}
}

func (g *Game) Finished() {
	return g.timeLeft <= 0
}

func (g *Game) showFinalScreen() {
	g.clear()

	fmt.Printf("Congratulations, your score was %d (%d%%)!\n", g.stats.Score, g.stats.Accuracy())
	fmt.Printf("Correct: %d, Mistakes: %d, Total: %d\n", g.stats.CorrectNotes, g.stats.MistakenNotes, g.stats.TotalNotes)
	fmt.Println()

	g.highscores.Add(g.name, g.stats.Score, g.stats.CorrectNotes, g.stats.TotalNotes)
	fmt.Println(g.highscores.String())

	fmt.Println()
	fmt.Println("Press 'r' to restart or 'q' to quit.")
}

func debug(s interface{}) {
	debugString = fmt.Sprintf("%v", s)
}

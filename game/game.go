package game

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/nov1n/guitarhero/colors"
	highscores "github.com/nov1n/guitarhero/highscores"
	stats "github.com/nov1n/guitarhero/stats"
)

var (
	debugString = ""
	keys        = "jkl;"
	keyColors   = []string{colors.GreenBackground, colors.RedBackground,
		colors.YellowBackground, colors.BlueBackground}
	flame = `  )
 ) \
/ ) (
\(_)/`

	// Indexes
	debugIndex     = 0
	barIndex       = 3
	missIndex      = barIndex - 2
	timeIndex      = barIndex - 1
	scoreIndex     = barIndex
	streakIndex    = barIndex + 1
	flameIndex     = barIndex + 2
	shortcutsIndex = barIndex + 10

	// Defaults
	roundLength = 30 * time.Second
	fps         = 7 // speed
	difficulty  = 6 // [0,10)
)

type Game struct {
	name             string
	fps              int
	difficulty       int
	timeLeft         time.Duration
	width            int
	height           int
	screen           []string
	keys             string
	highscores       *highscores.Highscores
	stats            *stats.Stats
	quitChan         chan int
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

	sh := th - 2 // allow 2 lines for user input

	return &Game{
		name:             n,
		fps:              fps,
		difficulty:       difficulty,
		timeLeft:         roundLength,
		width:            tw,
		height:           sh,
		screen:           []string{},
		keys:             keys,
		stats:            stats.New(),
		highscores:       highscores.New(),
		quitChan:         make(chan int, 1),
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
		if !g.Finished() {
			g.PauseUnpause()
		}
		return
	}

	if g.paused {
		return
	}

	g.updateScore(k)

	// Rerender to skip frame logic, but provide instant update
	g.rerenderFrame()
}

func (g *Game) updateScore(k string) {

	full := g.screen[barIndex]
	belowHalf := g.screen[barIndex-1]
	aboveHalf := g.screen[barIndex+1]

	// Check wrong key and invalid key
	colorIndex := strings.Index(keys, k)
	invalidKey := false
	wrongKey := false
	c := ""
	if colorIndex == -1 {
		invalidKey = true
	} else {
		c = keyColors[colorIndex]
		wrongKey = !strings.Contains((full + belowHalf + aboveHalf), c)
	}

	if invalidKey || wrongKey {
		g.stats.Incorrect()
		return
	}

	// From here on the note was correct
	g.stats.Correct()

	// Remove the note
	cString := fmt.Sprintf("%s %s", c, colors.Normal)

	// Check half below
	if strings.Contains(belowHalf, cString) {
		g.changeLine(barIndex-1, strings.Replace(g.screen[barIndex-1], cString, " ", 1))
		g.stats.Add(stats.Half)
		return
	}

	// Check full
	if strings.Contains(full, cString) {
		g.changeLine(barIndex, strings.Replace(g.screen[barIndex], cString, " ", 1))
		g.stats.Add(stats.Full)
		return
	}

	// Check half above
	if strings.Contains(aboveHalf, cString) {
		g.changeLine(barIndex+1, strings.Replace(g.screen[barIndex+1], cString, " ", 1))
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
	//key := g.keys[keyIdx]
	for i := 0; i < len(g.keys); i++ {
		curKey := " "
		if i == keyIdx {
			curKey = colors.Color(curKey, keyColors[i])
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
		if i == barIndex {
			for j := 0; j < len(keys); j++ {
				line = strings.Replace(line, "   ", fmt.Sprintf(" %s ", string(keys[j])), 1)
			}
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
			sidebar = "(r) restart"
		}
		if i == shortcutsIndex-2 {
			sidebar = "(q) quit"
		}
		if i == shortcutsIndex-3 {
			sidebar = "(p) pause/resume"
		}
		if i == debugIndex {
			sidebar = fmt.Sprintf("%s", debugString)
		}

		// Draw flame
		mul := g.stats.Multiplier()
		if mul > 1 {
			f := strings.Split(flame, "\n")
			flameColors := []string{colors.ThinGrey, colors.ThinWhite, colors.White}
			primaryFlameColor := flameColors[(mul-2)/(len(f))]
			secondaryFlameColor := flameColors[(mul-2)/len(f)+1]
			divIndex := (mul - 2) % len(f)

			for j := 0; j < len(f); j++ {
				if i == (j + flameIndex) {
					if j > divIndex {
						sidebar = colors.Color(f[len(f)-j-1], primaryFlameColor)
					} else {
						sidebar = colors.Color(f[len(f)-j-1], secondaryFlameColor)
					}
				}
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
	miss := strings.Contains(missLine, "\x1b")
	if miss {
		g.stats.Incorrect()
	}
}

func (g *Game) trim() {
	nScreen := len(g.screen)
	if nScreen > g.height {
		g.screen = g.screen[len(g.screen)-g.height:]
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

func (g *Game) Stop() {
	g.quitChan <- 1
}

func (g *Game) Restart() {
	g.restartChan <- 1
}

func (g *Game) PauseUnpause() {
	g.pauseUnpauseChan <- 1
}

func (g *Game) Loop() {
	frameLength := time.Duration(1000/g.fps) * time.Millisecond
	t := time.Tick(frameLength)

	for {
		select {
		case <-t:
			if rand.Intn(10) > (10 - difficulty) {
				g.appendRandomNote()
			} else {
				g.appendFret()
			}

			g.timeLeft -= frameLength

			g.advanceFrame()

			if g.Finished() {
				g.PauseUnpause()
				g.showFinalScreen()
			}
			break
		case <-g.pauseUnpauseChan:
			g.paused = !g.paused
			if t == nil {
				t = time.Tick(frameLength)
			} else {
				t = nil
			}
			break
		case <-g.quitChan:
			// Reset terminfo
			fmt.Print(colors.Color("", colors.Normal))
			return
		case <-g.restartChan:
			*g = *New(g.name) // Reassign main's reference to game to a new one
			g.Initialize()
			g.Loop()
			return
		}
	}
}

func (g *Game) Finished() bool {
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

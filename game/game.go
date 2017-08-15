package game

import (
	"fmt"
	"math"
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
	speedIndex     = barIndex - 2
	missIndex      = barIndex - 2
	timeIndex      = barIndex - 1
	scoreIndex     = barIndex
	streakIndex    = barIndex + 1
	flameIndex     = barIndex + 2
	shortcutsIndex = barIndex + 15
	shortcuts      = []string{
		"Shortcuts:",
		"(r) restart",
		"(q) quit",
		"(p) pause/resume",
		"(a) increase speed",
		"(z) decrease speed ",
	}

	// Defaults
	roundLength  = 30 * time.Second
	defaultSpeed = 7 // speed
	difficulty   = 6 // [0,10)
)

type Game struct {
	name             string
	speed            int
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
	frameTicker      <-chan time.Time
	paused           bool
}

func New(n string, w, h int) *Game {
	// Seed the random package to prevent default seed of 1
	rand.Seed(time.Now().UTC().UnixNano())

	return &Game{
		name:             n,
		speed:            defaultSpeed,
		difficulty:       difficulty,
		timeLeft:         roundLength,
		width:            w,
		height:           h,
		screen:           []string{},
		keys:             keys,
		highscores:       highscores.New(),
		quitChan:         make(chan int, 1),
		restartChan:      make(chan int, 1),
		pauseUnpauseChan: make(chan int, 1),
		paused:           true,
	}
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
	case "a":
		g.speed = int(math.Min(float64(g.speed+1), 12))
		g.Restart()
		return
	case "z":
		g.speed = int(math.Max(float64(g.speed-1), 1))
		g.Restart()
		return
	}

	if g.paused {
		return
	}

	g.updateScore(k)

	// Rerender to skip frame logic, but provide instant update
	g.rerenderFrame()
}

func (g *Game) updateScore(key string) {

	full := g.screen[barIndex]
	belowHalf := g.screen[barIndex-1]
	aboveHalf := g.screen[barIndex+1]

	// Check wrong key and invalid key
	colorIndex := strings.Index(keys, key)
	invalidKey := false
	wrongKey := false
	keyColor := ""
	if colorIndex == -1 {
		invalidKey = true
	} else {
		keyColor = keyColors[colorIndex]
		wrongKey = !strings.Contains(fmt.Sprintf("%s%s%s", full, belowHalf, aboveHalf), keyColor)
	}

	if invalidKey || wrongKey {
		g.stats.Incorrect()
		return
	}

	// From here on the note was correct
	g.stats.Correct()

	// Remove the note
	colorString := fmt.Sprintf("%s %s", keyColor, colors.Normal)

	// Check half below
	if strings.Contains(belowHalf, colorString) {
		g.changeLine(barIndex-1, strings.Replace(g.screen[barIndex-1], colorString, " ", 1))
		g.stats.Add(stats.Half)
		return
	}

	// Check full
	if strings.Contains(full, colorString) {
		g.changeLine(barIndex, strings.Replace(g.screen[barIndex], colorString, " ", 1))
		g.stats.Add(stats.Full)
		return
	}

	// Check half above
	if strings.Contains(aboveHalf, colorString) {
		g.changeLine(barIndex+1, strings.Replace(g.screen[barIndex+1], colorString, " ", 1))
		g.stats.Add(stats.Half)
	}
}

func (g *Game) Initialize() {
	Clear()

	if g.paused {
		g.PauseUnpause()
	}

	g.frameTicker = time.Tick(g.FrameLength()) // TODO: figure out why this doesn't work
	g.timeLeft = roundLength
	g.stats = stats.New()

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
	fmt.Println()
	for i := g.height - 1; i >= 0; i-- {
		line := g.screen[i]
		if i == barIndex+1 || i == barIndex-1 { // Draw the horizontal lines
			line = strings.Replace(line, " ", "-", -1)
		}

		if i == barIndex {
			// Draw the keys (in such a way that they may be coloured if a note passes)
			for j := 0; j < len(keys); j++ {
				line = replaceNth(line, " ", string(keys[j]), 2+2*j)
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
		if i == speedIndex {
			sidebar = fmt.Sprintf("speed: %d", g.speed)
		}
		for j := 0; j < len(shortcuts); j++ {
			if i == shortcutsIndex-j {
				sidebar = shortcuts[j]
			}
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

		fmt.Printf("  %s  %s", line, sidebar)

		if i != 0 { // Skip the last newline to allow for 'full screen'
			fmt.Println()
		}
	}
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

func (g *Game) Stop() {
	g.quitChan <- 1
}

func (g *Game) Restart() {
	g.restartChan <- 1
}

func (g *Game) PauseUnpause() {
	g.pauseUnpauseChan <- 1
}

func (g *Game) FrameLength() time.Duration {
	return time.Duration(1000/g.speed) * time.Millisecond
}

func (g *Game) Loop() {
	for {
		select {
		case <-g.frameTicker:
			if rand.Intn(10) > (10 - difficulty) {
				g.appendRandomNote()
			} else {
				g.appendFret()
			}
			g.timeLeft -= g.FrameLength()
			g.advanceFrame()
			if g.Finished() {
				g.PauseUnpause()
				g.showFinalScreen()
			}
			break
		case <-g.pauseUnpauseChan:
			g.paused = !g.paused
			if g.frameTicker == nil {
				g.frameTicker = time.Tick(g.FrameLength())
			} else {
				g.frameTicker = nil
			}
			break
		case <-g.quitChan:
			return
		case <-g.restartChan:
			g.Initialize()
			break
		}
	}
}

func (g *Game) Finished() bool {
	return g.timeLeft <= 0
}

func (g *Game) showFinalScreen() {
	Clear()

	fmt.Printf("Congratulations, your score was %d (%d%%)!\n", g.stats.Score, g.stats.Accuracy())
	fmt.Printf("Correct: %d, Mistakes: %d, Total: %d\n\n", g.stats.CorrectNotes, g.stats.MistakenNotes, g.stats.TotalNotes)

	// Add current score to highscores
	g.highscores.Add(g.name, g.speed, g.stats.Score, g.stats.CorrectNotes, g.stats.TotalNotes)

	// Show highscores
	fmt.Printf("%s\n\n\nPress 'r' to restart or 'q' to quit.", g.highscores.String(g.speed))
}

func debug(s interface{}) {
	debugString = fmt.Sprintf("%v", s)
}

func Clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func replaceNth(s, old, new string, n int) string {
	i := 0
	for m := 1; m <= n; m++ {
		x := strings.Index(s[i:], old)
		if x < 0 {
			break
		}
		i += x
		if m == n {
			return s[:i] + new + s[i+len(old):]
		}
		i += len(old)
	}
	return s
}

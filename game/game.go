package game

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	defaultFps    = 7
	maxStreak     = 9
	debugIndex    = 0
	barIndex      = 3
	missIndex     = barIndex - 2
	scoreIndex    = 3
	timeIndex     = 2
	streakIndex   = 4
	flameIndex    = 5
	multiplierInc = 10
	roundLength   = 30 * time.Second
	debugString   = ""
	defaultKeys   = []string{"u", "i", "o", "p"}
	halfSymbol    = "+"
	fullSymbol    = "-"
	flame         = `
  )
 ) \
/ ) (
\(_)/
`
)

type Game struct {
	name          string
	fps           int
	streak        int
	lastScore     int
	timeLeft      time.Duration
	totalNotes    int
	correctNotes  int
	mistakenNotes int
	width         int
	height        int
	screen        []string
	score         int
	keys          []string
	highscores    *Highscores
}

func NewGame(n string) *Game {
	rand.Seed(time.Now().UTC().UnixNano())
	tw, th, err := getTerminalDims()
	if err != nil {
		panic(err)
	}

	sh := th - 1 // allow 2 lines for user input

	return &Game{
		name:          n,
		fps:           defaultFps,
		streak:        0,
		lastScore:     0,
		totalNotes:    0,
		correctNotes:  0,
		mistakenNotes: 0,
		timeLeft:      roundLength,
		width:         tw,
		height:        sh,
		screen:        []string{},
		score:         0,
		keys:          defaultKeys,
		highscores:    NewHighscores(),
	}
}

func (g *Game) clear() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func (g *Game) addScore(n int) {
	g.lastScore = n
	g.score += n

	if g.score < 0 {
		g.score = 0
	}
}

func (g *Game) KeyPressed(k string) {
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
		g.mistakenNotes += 1
		g.streak = 0

		g.addScore(-g.halfScore())
		return
	}

	// From here on the note was correct

	g.correctNotes += 1
	g.totalNotes += 1

	// Check half below
	if strings.Contains(belowHalf, k) {
		// Mark the note as hit
		g.changeLine(barIndex-1, strings.Replace(g.screen[barIndex-1], k, halfSymbol, 1))

		// Remove k from next half point string to prevent double count
		aboveHalf = strings.Replace(aboveHalf, k, "", 1)
		full = strings.Replace(full, k, "", 1)

		g.addScore(g.halfScore())
	}

	// Check full
	if strings.Contains(full, k) {
		// Mark the note as hit
		g.changeLine(barIndex, strings.Replace(g.screen[barIndex], k, fullSymbol, 1))

		g.addScore(g.fullScore())

		// Remove k from the half point strings to prevent double count
		aboveHalf = strings.Replace(aboveHalf, k, "", 1)
	}

	// Check half above
	if strings.Contains(aboveHalf, k) {
		// Mark the note as hit
		g.changeLine(barIndex+1, strings.Replace(g.screen[barIndex+1], k, halfSymbol, 1))

		g.addScore(g.halfScore())
	}

	// Add to streak
	g.streak += 1
}

func (g *Game) multiplier() int {
	return int(math.Min(float64(1+(g.streak/multiplierInc)), float64(maxStreak)))
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
			sidebar = fmt.Sprintf("score: %d (%d%%) %d", g.score, g.accuracy(), g.lastScore)
		}
		if i == timeIndex {
			sidebar = fmt.Sprintf("time: %d", int(g.timeLeft.Seconds()))
		}
		if i == streakIndex {
			sidebar = fmt.Sprintf("%d (%dx)", g.streak, g.multiplier())
		}
		if i == debugIndex {
			sidebar = fmt.Sprintf("%s", debugString)
		}

		// Draw flame
		f := strings.Split(flame, "\n")
		for j := 0; j < len(f) && g.multiplier() > 1; j++ {
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
		g.streak = 0
		g.mistakenNotes += 1
		g.addScore(-g.halfScore())
	}

	// Count total notes
	if strings.ContainsAny(missLine, strings.Join(g.keys, "")) {
		g.totalNotes += 1
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

			if g.timeLeft <= 0 {
				g.showFinalScreen()
				os.Exit(0)
			}
		}
	}
}

func (g *Game) showFinalScreen() {
	g.clear()

	fmt.Printf("Congratulations, your score was %d (%d%%)!\n", g.score, g.accuracy())
	fmt.Printf("Correct: %d, Mistakes: %d, Total: %d\n", g.correctNotes, g.mistakenNotes, g.totalNotes)
	fmt.Println()

	g.highscores.Add(g.name, g.score, g.correctNotes, g.totalNotes)
	fmt.Println(g.highscores.String())
}

func (g *Game) halfScore() int {
	return 50 * g.multiplier()
}

func (g *Game) fullScore() int {
	return 2 * g.halfScore()
}

func (g *Game) accuracy() int {
	if g.totalNotes == 0 {
		return 0
	}
	return int(math.Max(float64(g.correctNotes)/(float64(g.totalNotes)+float64(g.mistakenNotes))*100, 0))
}

func debug(s interface{}) {
	debugString = fmt.Sprintf("%v", s)
}

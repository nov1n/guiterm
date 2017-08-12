package highscore

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"sort"
	"strings"
)

var (
	fileName = "highscore"
	showN    = 5
)

type highscore struct {
	name    string
	speed   int
	score   int
	correct int
	total   int
}

func (h highscore) serialize() string {
	return fmt.Sprintf("%s,%d,%d,%d,%d\n", h.name, h.speed, h.score, h.correct, h.total)
}

func deserialize(s string) highscore {
	h := highscore{}

	nameEndIdx := strings.Index(s, ",")
	h.name = s[:nameEndIdx]

	_, err := fmt.Sscanf(s[nameEndIdx+1:], "%d,%d,%d,%d", &h.speed, &h.score, &h.correct, &h.total)
	if err != nil {
		panic(err)
	}
	return h
}

type Highscores struct {
	entries []highscore
}

func New() *Highscores {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDONLY, 0666)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}

	h := &Highscores{}
	s := bufio.NewScanner(strings.NewReader(string(b)))
	for s.Scan() {
		h.entries = append(h.entries, deserialize(s.Text()))
	}

	return h
}

func (h *Highscores) serialize() string {
	var res bytes.Buffer
	for _, hs := range h.entries {
		res.WriteString(hs.serialize())
	}
	return res.String()
}

func (h *Highscores) Add(n string, p, s, c, t int) {
	hs := highscore{
		name:    n,
		speed:   p,
		score:   s,
		correct: c,
		total:   t,
	}
	h.entries = append(h.entries, hs)
	sort.Sort(sort.Reverse(h))

	h.save()
}

func (h *Highscores) save() {
	f, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	defer f.Close()
	if err != nil {
		panic(err)
	}
	f.WriteString(h.serialize())
}

func (h *Highscores) String(speed int) string {
	var res bytes.Buffer
	res.WriteString(fmt.Sprintf("Highscores for speed %d:\n", speed))

	count := 0
	filteredEntries := h.filterSpeed(speed)
	for i := 0; i < int(math.Min(float64(showN), float64(len(filteredEntries)))); i++ {
		hs := filteredEntries[i]
		count += 1
		res.WriteString(fmt.Sprintf("  %d. %s %d (%d/%d)\n", count, hs.name, hs.score, hs.correct, hs.total))
	}
	return res.String()
}

func (h *Highscores) filterSpeed(speed int) (res []highscore) {
	for i := 0; i < len(h.entries); i++ {
		if h.entries[i].speed == speed {
			res = append(res, h.entries[i])
		}
	}
	return
}

func (h *Highscores) Len() int {
	return len(h.entries)
}

func (h *Highscores) Less(i, j int) bool {
	return h.entries[i].score < h.entries[j].score
}

func (h *Highscores) Swap(i, j int) {
	temp := h.entries[i]
	h.entries[i] = h.entries[j]
	h.entries[j] = temp
}

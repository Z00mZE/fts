package russia

import (
	"bufio"
	"embed"
	"math/rand"
	"strings"
	"time"
)

//go:embed russian.txt
var dic embed.FS

type Dictionary struct {
	words    []string
	capacity int
	rand     *rand.Rand
}

func NewDictionary() *Dictionary {
	f, _ := dic.Open("russian.txt")
	defer func() {
		_ = f.Close()
	}()

	sc := bufio.NewScanner(f)
	self := &Dictionary{words: make([]string, 0, 1_532_630)}
	for sc.Scan() {
		self.words = append(self.words, strings.ToLower(sc.Text()))
	}
	self.rand = rand.New(rand.NewSource(time.Now().Unix()))
	self.capacity = len(self.words)

	return self
}
func (p *Dictionary) Rand(count int) []string {
	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		out = append(out, p.words[p.rand.Intn(p.capacity)])
	}
	return out
}

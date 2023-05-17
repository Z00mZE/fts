package index

import (
	"math"
	"sort"
	"strings"
	"unicode"

	snowballRussian "github.com/kljensen/snowball/russian"

	"github.com/Z00mZE/fts/domain/entity"
)

var stopWords = map[string]struct{}{
	"в":     {},
	"без":   {},
	"до":    {},
	"из":    {},
	"к":     {},
	"на":    {},
	"по":    {},
	"о":     {},
	"от":    {},
	"перед": {},
	"при":   {},
	"через": {},
	"для":   {},
	"с":     {},
	"у":     {},
	"за":    {},
	"над":   {},
	"об":    {},
	"под":   {},
	"про":   {},
}

type Index struct {
	index    map[string][]string
	registry map[string]entity.Document
}

func NewIndex() *Index {
	return &Index{
		index:    make(map[string][]string),
		registry: make(map[string]entity.Document),
	}
}
func (i *Index) Search(text string) []entity.Document {
	var docIds []string
	text = strings.ToLower(text)
	for _, token := range i.analyze(text) {
		if ids, ok := i.index[token]; ok {
			if docIds == nil {
				docIds = ids
			} else {
				docIds = i.intersection(docIds, ids)
			}
		} else {
			return nil
		}
	}
	docIdsLength := len(docIds)
	out := make([]entity.Document, 0, docIdsLength)
	for ijk := 0; ijk < docIdsLength; ijk++ {
		out = append(out, i.registry[docIds[ijk]])
	}
	return out
}
func (i *Index) Add(data entity.Document) {
	i.registry[data.ID] = data
	for _, token := range i.analyze(data.Text) {
		ids := i.index[token]
		if ids != nil && ids[len(ids)-1] == data.ID {
			continue
		}
		i.index[token] = append(ids, data.ID)
	}
}
func (i *Index) analyze(text string) []string {
	tokens := i.tokenize(text)
	tokens = i.lowercaseFilter(tokens)
	tokens = i.stopWordFilter(tokens)
	tokens = i.stemmerFilter(tokens)
	return tokens
}

func (i *Index) tokenize(text string) []string {
	return strings.FieldsFunc(text, i.stringSpliter)
}

func (i *Index) stringSpliter(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsNumber(r)
}

func (i *Index) lowercaseFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = strings.ToLower(token)
	}
	return r
}

func (i *Index) stopWordFilter(tokens []string) []string {
	r := make([]string, 0, len(tokens))
	for _, token := range tokens {
		if _, ok := stopWords[token]; !ok {
			r = append(r, token)
		}
	}
	return r
}

func (i *Index) stemmerFilter(tokens []string) []string {
	r := make([]string, len(tokens))
	for i, token := range tokens {
		r[i] = snowballRussian.Stem(token, false)
	}
	return r
}

func (i *Index) intersection(a []string, b []string) []string {
	r := make([]string, 0, int(math.Max(float64(len(a)), float64(len(b)))))
	var j, k int
	for j < len(a) && k < len(b) {
		switch {
		case a[j] == b[k]:
			r = append(r, a[j])
			j++
			k++
		case sort.StringsAreSorted([]string{a[j], b[k]}):
			j++
		default:
			k++
		}
	}
	return r
}

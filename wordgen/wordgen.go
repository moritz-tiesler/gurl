package wordgen

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"strings"
)

//go:embed first.json
var firstNames []byte

//go:embed last.json
var lastNames []byte

//go:embed verb.json
var verbs []byte

//go:embed personal-noun.json
var personalNouns []byte

type NameGen struct {
	firsts    []string
	lasts     []string
	verbs     []string
	personals []string
}

func (n *NameGen) Generate(id int32) string {
	high := id >> 24
	highMid := id >> 16 & 0xFF
	lowMid := id >> 8 & 0xFF
	low := id & 0xFF

	var b strings.Builder
	b.Grow(4)
	b.WriteString(n.firsts[high])
	b.WriteString(n.lasts[highMid])
	v := n.verbs[lowMid]
	b.WriteString(strings.ToUpper(string(v[0])) + v[1:])
	p := n.personals[low]
	b.WriteString(strings.ToUpper(string(p[0])) + p[1:])
	return b.String()

}

func New() *NameGen {
	firsts, lasts, vs, personals := load()
	shuffle(firsts)
	shuffle(lasts)
	shuffle(vs)
	shuffle(personals)
	return &NameGen{
		firsts:    firsts,
		lasts:     lasts,
		verbs:     vs,
		personals: personals,
	}
}

func shuffle(xs []string) {
	r := rand.New(rand.NewSource(55))
	r.Shuffle(len(xs), func(i, j int) {
		xs[i], xs[j] = xs[j], xs[i]
	})
}

type words []string

func load() (firsts words, lasts words, vs words, personals words) {
	json.Unmarshal(firstNames, &firsts)

	json.Unmarshal(lastNames, &lasts)

	json.Unmarshal(verbs, &vs)

	json.Unmarshal(personalNouns, &personals)
	return
}

package wordgen

import (
	_ "embed"
	"encoding/json"
	"math/rand"
	"strings"
	"sync"
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
	sync.Mutex
	count     int32
	firsts    []string
	lasts     []string
	verbs     []string
	personals []string
}

func (n *NameGen) Generate() string {
	n.Lock()
	defer n.Unlock()
	n.count++
	high := n.count >> 24
	highMid := n.count >> 16 & 0xFF
	lowMid := n.count >> 8 & 0xFF
	low := n.count & 0xFF

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
	rand.New(rand.NewSource(55))
	rand.Shuffle(len(xs), func(i, j int) {
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

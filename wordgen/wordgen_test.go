package wordgen

import (
	"testing"
)

func TestLoad(t *testing.T) {
	wg := New()

	if len(wg.firsts) < 1024 {
		t.Errorf("expected more firsts, got=%d", len(wg.firsts))
	}
	if len(wg.lasts) < 1024 {
		t.Errorf("expected more lasts, got=%d", len(wg.lasts))
	}
	if len(wg.verbs) < 1024 {
		t.Errorf("expected more verbs, got=%d", len(wg.verbs))
	}
	if len(wg.personals) < 1024 {
		t.Errorf("expected more personals, got=%d", len(wg.personals))
	}
}

func TestGenerate(t *testing.T) {
	wg := New()
	id := wg.Generate(1)
	t.Log(id)
}

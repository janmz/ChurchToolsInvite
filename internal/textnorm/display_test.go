package textnorm

import "testing"

func TestPrintableRemovesControlAndBidi(t *testing.T) {
	esc := string(rune(0x1b))
	in := "Max" + esc + "[2J Muster" + string(rune(0x202E)) + "x"
	got := Printable(in)
	if got != "Max [2J Muster x" {
		t.Fatalf("Printable = %q", got)
	}
	if Printable("Jürgen Müller") != "Jürgen Müller" {
		t.Fatal("normal text must stay unchanged")
	}
}

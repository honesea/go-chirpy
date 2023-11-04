package main

import "testing"

func TestProfanityFilter(t *testing.T) {
	message := "I really need a kerfuffle with sharbert to go to bed sooner, Fornax !"

	actual := cleanProfanity(message)
	expected := "I really need a **** with **** to go to bed sooner, **** !"

	if actual != expected {
		t.Errorf("Expected '%v' but got '%v'", expected, actual)
	}
}

package main

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestIsHidden(t *testing.T) {
	assert.Equal(t, isHidden(".dog.png"), true)
	assert.Equal(t, isHidden("dog.png"), false)
}

func TestIsPng(t *testing.T) {
	assert.Equal(t, isPng("dog.png"), true)
	assert.Equal(t, isPng("dog.jpg"), false)
}

package main

import (
	"testing"

	"github.com/magiconair/properties/assert"
)

func TestIsHidden(t *testing.T) {
	assert.Equal(t, isHidden(".dog.png"), true)
	assert.Equal(t, isHidden("dog.png"), false)
}

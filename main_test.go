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

func TestNewConfig(t *testing.T) {
	cfg := newConfig()

	assert.Equal(t, cfg.Host, "")
	assert.Equal(t, cfg.Port, "")
	assert.Equal(t, cfg.Tls, false)
	assert.Equal(t, cfg.Username, "")
	assert.Equal(t, cfg.Password, "")
	assert.Equal(t, cfg.SourcePath, "")
	assert.Equal(t, cfg.BaseUrl, "")
}

package main

import (
	"testing"

	"github.com/eliona-smart-building-assistant/app-integration-tests/assert"
	"github.com/eliona-smart-building-assistant/app-integration-tests/docker"
)

func TestMain(m *testing.M) {
	docker.RunApp(m)
}

func TestAppWorks(t *testing.T) {
	assert.AppWorks(t)
}

package main

import (
	"github.com/eliona-smart-building-assistant/app-integration-tests/app"
	"github.com/eliona-smart-building-assistant/app-integration-tests/assert"
	"github.com/eliona-smart-building-assistant/app-integration-tests/test"
	"testing"
)

func TestApp(t *testing.T) {
	app.StartApp()
	test.AppWorks(t)
	t.Run("TestAssetTypes", assetTypes)
	t.Run("TestWidgetTypes", widgetTypes)
	t.Run("TestSchema", schema)
	app.StopApp()
}

func assetTypes(t *testing.T) {
	t.Parallel()

	assert.AssetTypeExists(t, "kontakt_io_badge", []string{"temperature", "pos_world", "battery_level", "model", "firmware"})
	assert.AssetTypeExists(t, "kontakt_io_beacon", []string{})
	assert.AssetTypeExists(t, "kontakt_io_building", []string{})
	assert.AssetTypeExists(t, "kontakt_io_floor", []string{"height"})
	assert.AssetTypeExists(t, "kontakt_io_portal_beam", []string{})
	assert.AssetTypeExists(t, "kontakt_io_room", []string{})
	assert.AssetTypeExists(t, "kontakt_io_root", []string{})
	assert.AssetTypeExists(t, "kontakt_io_tag", []string{})
}

func widgetTypes(t *testing.T) {
	t.Parallel()

	assert.WidgetTypeExists(t, "Kontakt.io Air Sensor")
	assert.WidgetTypeExists(t, "Kontakt.io floor settings")
}

func schema(t *testing.T) {
	t.Parallel()

	assert.SchemaExists(t, "kontakt_io", []string{"configuration", "location", "tag"})
}

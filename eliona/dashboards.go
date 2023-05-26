//  This file is part of the eliona project.
//  Copyright © 2022 LEICOM iTEC AG. All Rights Reserved.
//  ______ _ _
// |  ____| (_)
// | |__  | |_  ___  _ __   __ _
// |  __| | | |/ _ \| '_ \ / _` |
// | |____| | | (_) | | | | (_| |
// |______|_|_|\___/|_| |_|\__,_|
//
//  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING
//  BUT NOT LIMITED  TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
//  NON INFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM,
//  DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
//  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

package eliona

import (
	"fmt"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/client"
	"github.com/eliona-smart-building-assistant/go-utils/common"
)

func DevicesDashboard(projectId string) (api.Dashboard, error) {
	dashboard := api.Dashboard{}
	dashboard.Name = "Kontakt.io sensors"
	dashboard.ProjectId = projectId
	dashboard.Widgets = []api.Widget{}

	beacons, _, err := client.NewClient().AssetsApi.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName(beaconAssetType).
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching beacon assets: %v", err)
	}

	portalBeams, _, err := client.NewClient().AssetsApi.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName(portalBeamAssetType).
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching Portal Beam assets: %v", err)
	}

	var occupancyData []api.WidgetData
	for _, portalBeam := range portalBeams {
		occ := api.WidgetData{
			ElementSequence: nullableInt32(1),
			AssetId:         portalBeam.Id,
			Data: map[string]interface{}{
				"aggregatedDataType": "heap",
				"attribute":          "people_count",
				"description":        fmt.Sprintf("%v - occupancy", *portalBeam.Name.Get()),
				"subtype":            "input",
			},
		}
		occupancyData = append(occupancyData, occ)
	}
	occupancyWidget := api.Widget{
		WidgetTypeName: "Donut",
		Details: map[string]interface{}{
			"size":     1,
			"timespan": 7,
		},
		Data: occupancyData,
	}
	dashboard.Widgets = append(dashboard.Widgets, occupancyWidget)

	airSensorAssets := append(beacons, portalBeams...)
	for _, airSensor := range airSensorAssets {
		widget := api.Widget{
			WidgetTypeName: "Kontakt.io Air Sensor",
			AssetId:        airSensor.Id,
			Details: map[string]interface{}{
				"size":     1,
				"timespan": 7,
			},
			Data: []api.WidgetData{
				{
					ElementSequence: nullableInt32(1),
					AssetId:         airSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "temperature",
						"description":         "Temperature",
						"key":                 "temperature",
						"seq":                 0,
						"subtype":             "input",
					},
				},
				{
					ElementSequence: nullableInt32(1),
					AssetId:         airSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "humidity",
						"description":         "Humidity",
						"key":                 "humidity",
						"seq":                 0,
						"subtype":             "input",
					},
				},
				{
					ElementSequence: nullableInt32(2),
					AssetId:         airSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "air_quality",
						"description":         "Air quality",
						"key":                 "",
						"seq":                 0,
						"subtype":             "input",
					},
				},
				{
					ElementSequence: nullableInt32(2),
					AssetId:         airSensor.Id,
					Data: map[string]interface{}{
						"aggregatedDataField": nil,
						"aggregatedDataType":  "heap",
						"attribute":           "battery_level",
						"description":         "Battery level",
						"key":                 "",
						"seq":                 0,
						"subtype":             "status",
					},
				},
			},
		}
		dashboard.Widgets = append(dashboard.Widgets, widget)
	}

	floors, _, err := client.NewClient().AssetsApi.
		GetAssets(client.AuthenticationContext()).
		AssetTypeName(floorAssetType).
		ProjectId(projectId).
		Execute()
	if err != nil {
		return api.Dashboard{}, fmt.Errorf("fetching floor assets: %v", err)
	}

	floorsWidget := api.Widget{
		WidgetTypeName: "Kontakt.io floor settings",
		Details: map[string]interface{}{
			"size":     1,
			"timespan": 7,
		},
	}

	for i, floor := range floors {
		d := []api.WidgetData{
			{
				ElementSequence: nullableInt32(1),
				AssetId:         floor.Id,
				Data: map[string]interface{}{
					"attribute":   "height",
					"description": floor.Name,
					"key":         "_SETPOINT",
					"seq":         i,
					"subtype":     "output",
				},
			},
			{
				ElementSequence: nullableInt32(1),
				AssetId:         floor.Id,
				Data: map[string]interface{}{
					"attribute":   "height",
					"description": floor.Name,
					"key":         "_CURRENT",
					"seq":         i,
					"subtype":     "output",
				},
			},
		}
		floorsWidget.Data = append(floorsWidget.Data, d...)
	}

	dashboard.Widgets = append(dashboard.Widgets, floorsWidget)

	return dashboard, nil
}

func nullableInt32(val int32) api.NullableInt32 {
	return *api.NewNullableInt32(common.Ptr(val))
}

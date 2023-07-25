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
	"context"
	"fmt"
	"kontakt-io/apiserver"
	"kontakt-io/conf"
	kontaktio "kontakt-io/kontakt-io"
	"math"
	"time"

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

func UpsertLocationData(config apiserver.Configuration, rooms []kontaktio.Room) error {
	floors := make(map[int]kontaktio.Floor)
	buildings := make(map[int]kontaktio.Building)

	for _, room := range rooms {
		// Ensure floors and buildings are unique.
		floor := room.Floor
		floors[floor.ID] = floor

		building := floor.Building
		buildings[building.ID] = building
	}

	for _, projectId := range conf.ProjIds(config) {
		for _, room := range rooms {
			err := upsertRoomData(config, projectId, room)
			if err != nil {
				return err
			}
		}
		for _, floor := range floors {
			err := upsertFloorData(config, projectId, floor)
			if err != nil {
				return err
			}
		}
		for _, building := range buildings {
			err := upsertBuildingData(config, projectId, building)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type roomInfoDataPayload struct{}

func upsertRoomData(config apiserver.Configuration, projectId string, room kontaktio.Room) error {
	log.Debug("Eliona", "upserting data for room: config %d and room '%v'", config.Id, room.ID)
	assetId, err := conf.GetLocationAssetId(context.Background(), config, projectId, kontaktio.RoomAssetType+fmt.Sprint(room.ID))
	if err != nil {
		return err
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	if err := upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		roomInfoDataPayload{},
	); err != nil {
		return fmt.Errorf("upserting info data: %v", err)
	}
	return nil
}

type floorInfoDataPayload struct{}

func upsertFloorData(config apiserver.Configuration, projectId string, floor kontaktio.Floor) error {
	log.Debug("Eliona", "upserting data for floor: config %d and floor '%v'", config.Id, floor.ID)
	assetId, err := conf.GetLocationAssetId(context.Background(), config, projectId, kontaktio.FloorAssetType+fmt.Sprint(floor.ID))
	if err != nil {
		return err
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	if err := upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		floorInfoDataPayload{},
	); err != nil {
		return fmt.Errorf("upserting info data: %v", err)
	}
	return nil
}

type buildingInfoDataPayload struct{}

func upsertBuildingData(config apiserver.Configuration, projectId string, building kontaktio.Building) error {
	log.Debug("Eliona", "upserting data for building: config %d and building '%v'", config.Id, building.ID)
	assetId, err := conf.GetLocationAssetId(context.Background(), config, projectId, kontaktio.BuildingAssetType+fmt.Sprint(building.ID))
	if err != nil {
		return err
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	if err := upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		buildingInfoDataPayload{},
	); err != nil {
		return fmt.Errorf("upserting info data: %v", err)
	}
	return nil
}

func UpsertDeviceData(config apiserver.Configuration, tags []kontaktio.Device) error {
	for _, projectId := range conf.ProjIds(config) {
		for _, tag := range tags {
			if err := upsertTagData(config, projectId, tag); err != nil {
				return fmt.Errorf("upserting tag data: %v", err)
			}
		}
	}
	return nil
}

type deviceInfoDataPayload struct {
	Firmware string `json:"firmware"`
	Model    string `json:"model"`
}

type deviceStatusDataPayload struct {
	BatteryLevel string `json:"battery_level"`
}

type badgeInputDataPayload struct {
	WorldPosition []float64 `json:"pos_world"`
	Temperature   string    `json:"temperature"`
}

type beaconInputDataPayload struct {
	AirPressure    string `json:"air_pressure"`
	Humidity       string `json:"humidity"`
	LightIntensity string `json:"light_intensity"`
	Temperature    string `json:"temperature"`
	AirQuality     string `json:"air_quality"`
}

type portalBeamInputDataPayload struct {
	AirPressure    string `json:"air_pressure"`
	Humidity       string `json:"humidity"`
	LightIntensity string `json:"light_intensity"`
	Temperature    string `json:"temperature"`
	AirQuality     string `json:"air_quality"`
	PeopleCount    string `json:"people_count"`
}

type tagInputDataPayload struct {
	WorldPosition []float64 `json:"pos_world"`
}

func upsertTagData(config apiserver.Configuration, projectId string, device kontaktio.Device) error {
	log.Debug("Eliona", "upserting data for device %+v", device)
	assetId, err := conf.GetTagAssetId(context.Background(), config, projectId, device.Type+fmt.Sprint(device.ID))
	if err != nil {
		return fmt.Errorf("getting asset id: %v", err)
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	if err := upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		deviceInfoDataPayload{
			Firmware: device.Firmware,
			Model:    fmt.Sprint(device.Product),
		},
	); err != nil {
		return err
	}
	if err := upsertData(
		api.SUBTYPE_STATUS,
		*assetId,
		deviceStatusDataPayload{
			BatteryLevel: fmt.Sprint(device.BatteryLevel),
		},
	); err != nil {
		return err
	}

	var inputData any
	switch device.Type {
	case kontaktio.TagAssetType:
		inputData = tagInputDataPayload{
			WorldPosition: device.WorldPosition,
		}
	case kontaktio.BeaconAssetType:
		inputData = beaconInputDataPayload{
			Humidity:       fmt.Sprint(device.Humidity),
			LightIntensity: fmt.Sprint(device.LightIntensity),
			Temperature:    fmt.Sprint(math.Round(device.Temperature*10) / 10),
			AirQuality:     fmt.Sprint(device.AirQuality),
			AirPressure:    fmt.Sprint(device.AirPressure),
		}
	case kontaktio.PortalBeamAssetType:
		inputData = portalBeamInputDataPayload{
			Humidity:       fmt.Sprint(device.Humidity),
			LightIntensity: fmt.Sprint(device.LightIntensity),
			Temperature:    fmt.Sprint(math.Round(device.Temperature*10) / 10),
			AirQuality:     fmt.Sprint(device.AirQuality),
			AirPressure:    fmt.Sprint(device.AirPressure),
			PeopleCount:    fmt.Sprint(device.PeopleCount),
		}
	case kontaktio.BadgeAssetType:
		inputData = badgeInputDataPayload{
			WorldPosition: device.WorldPosition,
			Temperature:   fmt.Sprint(math.Round(device.Temperature*10) / 10),
		}
	default:
		return fmt.Errorf("unknown asset type \"%s\"", device.Type)
	}
	if err := upsertData(api.SUBTYPE_INPUT, *assetId, inputData); err != nil {
		return err
	}
	return nil
}

func upsertData(subtype api.DataSubtype, assetId int32, payload any) error {
	var statusData api.Data
	statusData.Subtype = subtype
	now := time.Now()
	statusData.Timestamp = *api.NewNullableTime(&now)
	statusData.AssetId = assetId
	statusData.Data = common.StructToMap(payload)
	if err := asset.UpsertDataIfAssetExists(statusData); err != nil {
		return fmt.Errorf("upserting data: %v", err)
	}
	return nil
}

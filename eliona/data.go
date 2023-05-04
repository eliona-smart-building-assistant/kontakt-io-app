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

type roomInfoDataPayload struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func upsertRoomData(config apiserver.Configuration, projectId string, room kontaktio.Room) error {
	log.Debug("Eliona", "upserting data for room: config %d and room '%s'", config.Id, room.ID)
	assetId, err := conf.GetLocationAssetId(context.Background(), config, projectId, int32(room.ID))
	if err != nil {
		return err
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	if err := upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		roomInfoDataPayload{
			ID:   room.ID,
			Name: room.Name,
		},
	); err != nil {
		return fmt.Errorf("upserting info data: %v", err)
	}
	return nil
}

type floorInfoDataPayload struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Level int    `json:"level"`
}

func upsertFloorData(config apiserver.Configuration, projectId string, floor kontaktio.Floor) error {
	log.Debug("Eliona", "upserting data for floor: config %d and floor '%s'", config.Id, floor.ID)
	assetId, err := conf.GetLocationAssetId(context.Background(), config, projectId, int32(floor.ID))
	if err != nil {
		return err
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	if err := upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		floorInfoDataPayload{
			ID:    floor.ID,
			Name:  floor.Name,
			Level: floor.Level,
		},
	); err != nil {
		return fmt.Errorf("upserting info data: %v", err)
	}
	return nil
}

type buildingInfoDataPayload struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

func upsertBuildingData(config apiserver.Configuration, projectId string, building kontaktio.Building) error {
	log.Debug("Eliona", "upserting data for building: config %d and building '%s'", config.Id, building.ID)
	assetId, err := conf.GetLocationAssetId(context.Background(), config, projectId, int32(building.ID))
	if err != nil {
		return err
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	if err := upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		buildingInfoDataPayload{
			ID:          building.ID,
			Name:        building.Name,
			Address:     building.Address,
			Description: building.Description,
		},
	); err != nil {
		return fmt.Errorf("upserting info data: %v", err)
	}
	return nil
}

func UpsertTagData(config apiserver.Configuration, tags []kontaktio.Tag) error {
	for _, projectId := range conf.ProjIds(config) {
		for _, tag := range tags {
			if err := upsertTagData(config, projectId, tag); err != nil {
				return fmt.Errorf("upserting MultiSensor data: %v", err)
			}
		}
	}
	return nil
}

type tagInfoDataPayload struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Firmware string `json:"firmware"`
	Model    string `json:"model"`
}

type tagStatusDataPayload struct {
	BatteryLevel string `json:"battery_level"`
}

type tagInputDataPayload struct {
	PosX           string `json:"pos_x"`
	PosY           string `json:"pos_y"`
	Humidity       string `json:"humidity"`
	LigthIntensity string `json:"light_intensity"`
	Temperature    string `json:"temperature"`
	AirQuality     string `json:"air_quality"`
	AirPressure    string `json:"air_pressure"`
}

func upsertTagData(config apiserver.Configuration, projectId string, tag kontaktio.Tag) error {
	log.Debug("Eliona", "upserting data for tag %s", tag.Name)
	assetId, err := conf.GetTagAssetId(context.Background(), config, projectId, tag.ID)
	if err != nil {
		return fmt.Errorf("getting asset id: %v", err)
	}
	if assetId == nil {
		return fmt.Errorf("unable to find asset ID")
	}
	if err := upsertData(
		api.SUBTYPE_INFO,
		*assetId,
		tagInfoDataPayload{
			ID:       tag.ID,
			Name:     tag.Name,
			Firmware: tag.Firmware,
			Model:    fmt.Sprint(tag.Model),
		},
	); err != nil {
		return err
	}
	if err := upsertData(
		api.SUBTYPE_STATUS,
		*assetId,
		tagStatusDataPayload{
			BatteryLevel: fmt.Sprint(tag.BatteryLevel),
		},
	); err != nil {
		return err
	}
	if err := upsertData(
		api.SUBTYPE_INPUT,
		*assetId,
		tagInputDataPayload{
			PosX:           fmt.Sprint(tag.PositionX),
			PosY:           fmt.Sprint(tag.PositionY),
			Humidity:       fmt.Sprint(tag.Humidity),
			LigthIntensity: fmt.Sprint(tag.LightIntensity),
			Temperature:    fmt.Sprint(tag.Temperature),
			AirQuality:     fmt.Sprint(tag.AirQuality),
			AirPressure:    fmt.Sprint(tag.AirPressure),
		},
	); err != nil {
		return err
	}
	return nil
}

//

func upsertData(subtype api.DataSubtype, assetId int32, payload any) error {
	var statusData api.Data
	statusData.Subtype = subtype
	now := time.Now()
	statusData.Timestamp = *api.NewNullableTime(&now)
	statusData.AssetId = assetId
	statusData.Data = common.StructToMap(payload)
	if err := asset.UpsertDataIfAssetExists[any](statusData); err != nil {
		return fmt.Errorf("upserting data: %v", err)
	}
	return nil
}

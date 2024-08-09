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

package kontaktio

import (
	"context"
	"fmt"
	"kontakt-io/apiserver"
	"kontakt-io/conf"
	"net/url"
	"strings"
	"time"

	"github.com/eliona-smart-building-assistant/go-eliona/utils"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

const TagAssetType = "kontakt_io_tag"
const BadgeAssetType = "kontakt_io_badge"
const BeaconAssetType = "kontakt_io_beacon"
const PortalBeamAssetType = "kontakt_io_portal_beam"
const RoomAssetType = "kontakt_io_room"
const FloorAssetType = "kontakt_io_floor"
const BuildingAssetType = "kontakt_io_building"

const RootAssetType = "kontakt_io_root"

const productAnchorBeacon = "Anchor Beacon 2"
const productAssetTag = "Asset Tag 2"
const productNanoTag = "Nano Tag"
const productPortalBeam = "Portal Beam"
const productPortalLight = "Portal Light AC EU (Plug F)"
const productPuckBeacon = "Puck Beacon"
const productSmartBadge = "Smart Badge"

type Building struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

type Floor struct {
	ID       int      `json:"id"`
	Name     string   `json:"name"`
	Building Building `json:"building"`
	Level    int      `json:"level"`
}

type Room struct {
	ID         int    `json:"id"`
	RoomNumber int32  `json:"roomNumber"`
	Name       string `json:"name"`
	Floor      Floor  `json:"floor"`
}

type locationsResponse struct {
	Content []Room `json:"content"`
}

func GetRooms(config apiserver.Configuration) ([]Room, error) {
	u := "https://apps.cloud.us.kontakt.io/v2/locations/rooms?size=2000"
	r, err := http.NewRequestWithApiKey(u, "API-Key", config.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("creating request to %s: %v", u, err)
	}
	locationsResponse, err := http.Read[locationsResponse](r, time.Duration(*config.RequestTimeout)*time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %v", u, err)
	}
	return locationsResponse.Content, nil
}

type Device struct {
	ID             string  `json:"trackingId"`
	Name           string  `json:"uniqueId"`
	Firmware       string  `json:"firmware"`
	Product        string  `json:"product"`
	BatteryLevel   int     `json:"batteryLevel"`
	Humidity       int     `json:"humidity"`
	LightIntensity int     `json:"lightIntensity"`
	Temperature    float64 `json:"temperature"`
	AirQuality     int     `json:"airQuality"`
	AirPressure    float64 `json:"airPressure"`
	PeopleCount    int     `json:"numberOfPeopleDetected"`
	RoomNumberIr   *int32  `json:"-"`

	Type          string
	WorldPosition []float64

	PositionX float64   `json:"x"`
	PositionY float64   `json:"y"`
	FloorID   int       `json:"floorId"`
	Timestamp time.Time `json:"timestamp"`
}

type deviceInfo struct {
	Model        string `json:"model" eliona:"model,filterable"`
	ID           string `json:"id" eliona:"id,filterable"`
	BatteryLevel int    `json:"batteryLevel" eliona:"battery_level,filterable"`
	Product      string `json:"product" eliona:"product,filterable"`
	Name         string `json:"name" eliona:"name,filterable"`
	Mac          string `json:"mac" eliona:"mac,filterable"`
	Firmware     string `json:"firmware" eliona:"firmware,filterable"`
	RoomNumberIr *int32 `json:"irRoomNumber"`
}

type deviceResponse struct {
	Devices []deviceInfo `json:"devices"`
}

type telemetryResponse struct {
	Content []Device `json:"content"`
}

type positionsResponse struct {
	Content []Device `json:"content"`
}

func fetchDevices(config apiserver.Configuration) (map[string]Device, error) {
	headers := map[string]string{
		"API-Key": config.ApiKey,
		"Accept":  "application/vnd.com.kontakt+json;version=10",
	}
	deviceUrl := "https://api.kontakt.io/device"
	r, err := http.NewRequestWithHeaders(deviceUrl, headers)
	if err != nil {
		return nil, fmt.Errorf("creating request to %s: %v", deviceUrl, err)
	}
	deviceResponse, err := http.Read[deviceResponse](r, time.Duration(*config.RequestTimeout)*time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %v", deviceUrl, err)
	}
	tags := make(map[string]Device)
	for _, device := range deviceResponse.Devices {
		if adheres, err := device.AdheresToFilter(config); err != nil {
			return nil, fmt.Errorf("checking if device adheres to a device filter: %v", err)
		} else if !adheres {
			log.Debug("kontaktio", "Device %v - %v skipped, does not adhere to asset filter.", device.Name, device.Product)
			continue
		}
		uid := strings.ToLower(device.Mac)
		tags[uid] = Device{
			ID:           uid,
			Name:         fmt.Sprintf("%v %v", device.Product, device.Name),
			BatteryLevel: device.BatteryLevel,
			Firmware:     device.Firmware,
			Product:      device.Product,
			RoomNumberIr: device.RoomNumberIr,
		}
	}

	return tags, nil
}

func fetchTelemetry(config apiserver.Configuration, potentialTags map[string]Device) ([]Device, error) {
	telemetryUrl := "https://apps.cloud.us.kontakt.io/v3/telemetry"
	u, err := url.Parse(telemetryUrl)
	if err != nil {
		return nil, fmt.Errorf("shouldn't happen: parsing telemetry URL: %v", err)
	}
	trackingIDs := make([]string, 0, len(potentialTags))
	for id := range potentialTags {
		trackingIDs = append(trackingIDs, id)
	}
	trackingIDsFormatted := strings.Join(trackingIDs, ",")
	now := time.Now().UTC()
	startTime := now.Add(-2 * time.Minute) // The devices should report themselves every 1 minute, so we should give some margin.
	startTimeFormatted := startTime.Format(time.RFC3339)
	endTimeFormatted := now.Format(time.RFC3339)

	q := u.Query()
	q.Set("trackingId", trackingIDsFormatted)
	q.Set("startTime", startTimeFormatted)
	q.Set("endTime", endTimeFormatted)
	q.Set("size", "2000") // 2000 is the biggest allowed page size
	// q.Set("sort", "timestamp,desc") - not respected at all, for some reason.
	u.RawQuery = q.Encode()

	r, err := http.NewRequestWithApiKey(u.String(), "API-Key", config.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("creating request to %s: %v", u.String(), err)
	}
	telemetryResponse, err := http.Read[telemetryResponse](r, time.Duration(*config.RequestTimeout)*time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %v", u.String(), err)
	}

	return telemetryResponse.Content, nil
}

func fetchPositions(config apiserver.Configuration) ([]Device, error) {
	positionsUrl := "https://apps.cloud.us.kontakt.io/v2/positions?size=2000"
	r, err := http.NewRequestWithApiKey(positionsUrl, "API-Key", config.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("creating request to %s: %v", positionsUrl, err)
	}
	positionsResponse, err := http.Read[positionsResponse](r, time.Duration(*config.RequestTimeout)*time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %v", positionsUrl, err)
	}

	return positionsResponse.Content, nil
}

func GetDevices(config apiserver.Configuration) ([]Device, error) {
	devices, err := fetchDevices(config)
	if err != nil {
		return nil, fmt.Errorf("fetching devices: %v", err)
	}

	telemetry, err := fetchTelemetry(config, devices)
	if err != nil {
		return nil, fmt.Errorf("fetching telemetry: %v", err)
	}

	tags := make(map[string]Device, len(telemetry))
	for _, t := range telemetry {
		if tt, ok := tags[t.ID]; ok {
			if tt.Timestamp.After(t.Timestamp) {
				// Already got newer data
				continue
			}
		}
		tags[t.ID] = t
	}

	positions, err := fetchPositions(config)
	if err != nil {
		return nil, fmt.Errorf("fetching positions: %v", err)
	}

	for _, p := range positions {
		f, err := conf.GetLocationIrrespectibleOfProject(context.Background(), config, FloorAssetType+fmt.Sprint(p.FloorID))
		if err != nil {
			return nil, fmt.Errorf("finding floor %v (irrespectible of project): %v", p.FloorID, err)
		}
		if f == nil {
			log.Error("kontakt-io", "found no corresponding location for tag %v floor %v", p.ID, p.FloorID)
			continue
		}
		floor := *f
		floorHeight := floor.FloorHeight.Float64
		if floor.FloorHeight.Valid == false {
			log.Info("kontakt-io", "floor %v has no height set, assuming 0", floor.AssetID.Int32)
			floorHeight = 0
		}

		x := p.PositionX - config.AbsoluteX
		y := p.PositionY - config.AbsoluteY
		p.WorldPosition = []float64{x, y, floorHeight}
		if t, ok := tags[p.ID]; ok {
			t.WorldPosition = p.WorldPosition
			p = t
		}
		tags[p.ID] = p
	}
	tagsSlice := make([]Device, 0, len(tags))
	for _, tag := range tags {
		t, ok := devices[tag.ID]
		if !ok {
			// This happens due to matching MAC address with trackingID.
			// As this should only be the case for portal lights that provide no valuable
			// information, we ignore the error.
			//
			// Response from kontakt.io support:
			// In telemetry trackingID is always mac address of beacon or Portal Light.
			// For Portal Lights you may see difference by +2. Basically BLE mac = WiFi mac + 2
			log.Debug("kontakt-io", "A tracking ID %v was not matched with a device.", tag.ID)
			continue
		}
		switch t.Product {
		case productSmartBadge, productAssetTag:
			tag.Type = BadgeAssetType
		case productNanoTag:
			tag.Type = TagAssetType
		case productAnchorBeacon, productPuckBeacon:
			tag.Type = BeaconAssetType
		case productPortalBeam:
			tag.Type = PortalBeamAssetType
		case productPortalLight:
			// Provides no valuable information.
			continue
		default:
			log.Debug("kontakt-io", "Skipped unsupported product: %s", t.Product)
			continue
		}

		tag.Name = t.Name
		tag.BatteryLevel = t.BatteryLevel
		tag.Firmware = t.Firmware
		tag.RoomNumberIr = t.RoomNumberIr
		tagsSlice = append(tagsSlice, tag)
	}

	return tagsSlice, nil
}

func (device *deviceInfo) AdheresToFilter(config apiserver.Configuration) (bool, error) {
	f := apiFilterToCommonFilter(config.AssetFilter)
	fp, err := utils.StructToMap(device)
	if err != nil {
		return false, fmt.Errorf("converting struct to map: %v", err)
	}
	adheres, err := common.Filter(f, fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
}

func apiFilterToCommonFilter(input [][]apiserver.FilterRule) [][]common.FilterRule {
	result := make([][]common.FilterRule, len(input))
	for i := 0; i < len(input); i++ {
		result[i] = make([]common.FilterRule, len(input[i]))
		for j := 0; j < len(input[i]); j++ {
			result[i][j] = common.FilterRule{
				Parameter: input[i][j].Parameter,
				Regex:     input[i][j].Regex,
			}
		}
	}
	return result
}

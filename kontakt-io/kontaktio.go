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
	"fmt"
	"kontakt-io/apiserver"
	"net/url"
	"reflect"
	"strings"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	"github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

// Keep in sync with ../eliona/assets.go
const tagAssetType = "kontakt_io_tag"
const badgeAssetType = "kontakt_io_badge"
const beaconAssetType = "kontakt_io_beacon"
const portalBeamAssetType = "kontakt_io_portal_beam"
const roomAssetType = "kontakt_io_room"
const floorAssetType = "kontakt_io_floor"
const buildingAssetType = "kontakt_io_building"

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
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Floor Floor  `json:"floor"`
}

type locationsResponse struct {
	Content []Room `json:"content"`
}

func GetRooms(config apiserver.Configuration) ([]Room, error) {
	url := "https://apps.cloud.us.kontakt.io/v2/locations/rooms?size=2000"
	r, err := http.NewRequestWithApiKey(url, "API-Key", config.ApiKey)
	if err != nil {
		return nil, fmt.Errorf("creating request to %s: %v", url, err)
	}
	locationsResponse, err := http.Read[locationsResponse](r, time.Duration(*config.RequestTimeout)*time.Second, false)
	if err != nil {
		return nil, fmt.Errorf("reading response from %s: %v", url, err)
	}
	return locationsResponse.Content, nil
}

type Device struct {
	ID             string  `json:"trackingId"`
	Name           string  `json:"uniqueId"`
	Firmware       string  `json:"firmware"`
	Product        string  `json:"product"`
	BatteryLevel   int     `json:"batteryLevel"`
	PositionX      float64 `json:"x"`
	PositionY      float64 `json:"y"`
	Humidity       int     `json:"humidity"`
	LightIntensity int     `json:"lightIntensity"`
	Temperature    float64 `json:"temperature"`
	AirQuality     int     `json:"airQuality"`
	AirPressure    float64 `json:"airPressure"`
	PeopleCount    int     `json:"numberOfPeopleDetected"`

	Type string

	timestamp time.Time `json:"timestamp"`
}

type deviceInfo struct {
	Model        string `json:"model"`
	ID           string `json:"id"`
	BatteryLevel int    `json:"batteryLevel"`
	Product      string `json:"product"`
	Name         string `json:"name"`
	Mac          string `json:"mac"`
	Firmware     string `json:"firmware"`
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
			log.Debug("kontaktio", "Device %v skipped, does not adhere to asset filter.", device.Name)
			continue
		}
		uid := strings.ToLower(device.Mac)
		tags[uid] = Device{
			ID:           uid,
			Name:         fmt.Sprintf("%v %v", device.Product, device.Name),
			BatteryLevel: device.BatteryLevel,
			Firmware:     device.Firmware,
			Product:      device.Product,
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
	startTime := now.Add(-5 * time.Minute)
	startTimeFormatted := startTime.Format(time.RFC3339)
	endTimeFormatted := now.Format(time.RFC3339)

	q := u.Query()
	q.Set("trackingId", trackingIDsFormatted)
	q.Set("startTime", startTimeFormatted)
	q.Set("endTime", endTimeFormatted)
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

func fetchPositions(config apiserver.Configuration, tags map[string]Device) ([]Device, error) {
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
			if tt.timestamp.After(t.timestamp) {
				// Already got newer data
				continue
			}
		}
		tags[t.ID] = t
	}

	positions, err := fetchPositions(config, tags)
	if err != nil {
		return nil, fmt.Errorf("fetching positions: %v", err)
	}

	for _, p := range positions {
		if t, ok := tags[p.ID]; ok {
			t.PositionX = p.PositionX
			t.PositionY = p.PositionY
			p = t
		}
		tags[p.ID] = p
	}
	tagsSlice := make([]Device, 0, len(tags))
	for _, tag := range tags {
		t, ok := devices[tag.ID]
		if !ok {
			// This happens due to matching Mac address with trackingID.
			// As this should only be the case for portal lights that provide no valuable
			// information, we ignore the error.
			//
			// Response from kontakt.io support:
			// In telemetry trackingID is always mac address of beacon or Portal Light.
			// For Portal Lights you may see difference by +2. Basically BLE mac = WiFi mac + 2
			log.Debug("kontakt-io", "A tracking ID was %v not matched with a device.", tag.ID)
			continue
		}
		switch t.Product {
		case productSmartBadge, productAssetTag:
			tag.Type = badgeAssetType
		case productNanoTag:
			tag.Type = tagAssetType
		case productAnchorBeacon, productPuckBeacon:
			tag.Type = beaconAssetType
		case productPortalBeam:
			tag.Type = portalBeamAssetType
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
		tagsSlice = append(tagsSlice, tag)
	}

	return tagsSlice, nil
}

func (device *deviceInfo) AdheresToFilter(config apiserver.Configuration) (bool, error) {
	f := apiFilterToCommonFilter(config.AssetFilter)
	fp, err := structToMap(device)
	if err != nil {
		return false, fmt.Errorf("converting strict to map: %v", err)
	}
	adheres, err := common.Filter(f, fp)
	if err != nil {
		return false, err
	}
	return adheres, nil
}

func structToMap(input interface{}) (map[string]string, error) {
	if input == nil {
		return nil, fmt.Errorf("input is nil")
	}

	inputValue := reflect.ValueOf(input)
	inputType := reflect.TypeOf(input)

	if inputValue.Kind() == reflect.Ptr {
		inputValue = inputValue.Elem()
		inputType = inputType.Elem()
	}

	if inputValue.Kind() != reflect.Struct {
		return nil, fmt.Errorf("input is not a struct")
	}

	output := make(map[string]string)
	for i := 0; i < inputValue.NumField(); i++ {
		fieldValue := inputValue.Field(i)
		fieldType := inputType.Field(i)
		output[fieldType.Name] = fieldValue.String()
	}

	return output, nil
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

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

import "kontakt-io/apiserver"

type Building struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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

func GetRooms(config apiserver.Configuration) ([]Room, error) {
	return nil, nil
}

type Tag struct {
	ID             string  `json:"trackingId"`
	Name           string  `json:"uniqueId"`
	Firmware       string  `json:"firmware"`
	Model          int     `json:"model"`
	BatteryLevel   int     `json:"batteryLevel"`
	PositionX      float64 `json:"pos_x"`
	PositionY      float64 `json:"pos_y"`
	Humidity       int     `json:"humidity"`
	LightIntensity int     `json:"lightIntensity"`
	Temperature    float64 `json:"temperature"`
	AirQuality     int     `json:"airQuality"`
	AirPressure    float64 `json:"airPressure"`
}

func GetTags(config apiserver.Configuration) ([]Tag, error) {
	return nil, nil
}

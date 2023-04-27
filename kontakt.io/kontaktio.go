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
	"reflect"

	"github.com/eliona-smart-building-assistant/go-utils/common"
)

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

func (tag *Tag) AdheresToFilter(config apiserver.Configuration) (bool, error) {
	f := apiFilterToCommonFilter(config.AssetFilter)
	fp, err := structToMap(tag)
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

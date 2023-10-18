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

package main

import (
	"context"
	"kontakt-io/apiserver"
	"kontakt-io/apiservices"
	"kontakt-io/conf"
	"kontakt-io/eliona"
	kontaktio "kontakt-io/kontakt-io"
	"net/http"
	"time"

	"github.com/eliona-smart-building-assistant/go-utils/common"
	utilshttp "github.com/eliona-smart-building-assistant/go-utils/http"
	"github.com/eliona-smart-building-assistant/go-utils/log"
)

// collectData is the main app function which is called periodically
func collectData() {
	configs, err := conf.GetConfigs(context.Background())
	if err != nil {
		log.Fatal("conf", "Couldn't read configs from DB: %v", err)
		return
	}
	if len(configs) == 0 {
		log.Info("conf", "No configs in DB")
		return
	}

	for _, config := range configs {
		// Skip config if disabled and set inactive
		if !conf.IsConfigEnabled(config) {
			if conf.IsConfigActive(config) {
				_, err := conf.SetConfigActiveState(context.Background(), config, false)
				if err != nil {
					log.Fatal("conf", "Couldn't set config active state to DB: %v", err)
					return
				}
			}
			continue
		}

		// Signals that this config is active
		if !conf.IsConfigActive(config) {
			_, err := conf.SetConfigActiveState(context.Background(), config, true)
			if err != nil {
				log.Fatal("conf", "Couldn't set config active state to DB: %v", err)
				return
			}
			log.Info("conf", "Collecting initialized with Configuration %d:\n"+
				"API Key: %s\n"+
				"Enable: %t\n"+
				"Refresh Interval: %d\n"+
				"Request Timeout: %d\n"+
				"Active: %t\n"+
				"Project IDs: %v\n",
				*config.Id,
				config.ApiKey,
				*config.Enable,
				config.RefreshInterval,
				*config.RequestTimeout,
				*config.Active,
				*config.ProjectIDs)
		}

		common.RunOnceWithParam(func(config apiserver.Configuration) {
			log.Info("main", "Collecting %d started", *config.Id)

			if err := collectLocations(config); err != nil {
				return // Error is handled in the method itself.
			}
			if err := collectDevices(config); err != nil {
				return // Error is handled in the method itself.
			}

			log.Info("main", "Collecting %d finished", *config.Id)

			time.Sleep(time.Second * time.Duration(config.RefreshInterval))
		}, config, *config.Id)
	}
}

func collectLocations(config apiserver.Configuration) error {
	rooms, err := kontaktio.GetRooms(config)
	if err != nil {
		log.Error("kontakt-io", "getting rooms: %v", err)
		return err
	}
	if err := eliona.CreateLocationAssetsIfNecessary(config, rooms); err != nil {
		log.Error("eliona", "creating location assets: %v", err)
		return err
	}

	if err := eliona.UpsertLocationData(config, rooms); err != nil {
		log.Error("eliona", "inserting location data into Eliona: %v", err)
		return err
	}
	return nil
}

func collectDevices(config apiserver.Configuration) error {
	devices, err := kontaktio.GetDevices(config)
	if err != nil {
		log.Error("kontakt-io", "getting devices info: %v", err)
		return err
	}
	if err := eliona.CreateDeviceAssetsIfNecessary(config, devices); err != nil {
		log.Error("eliona", "creating tag assets: %v", err)
		return err
	}
	if err := eliona.UpsertDeviceData(config, devices); err != nil {
		log.Error("eliona", "inserting location data into Eliona: %v", err)
		return err
	}
	return nil
}

func listenForOutputChanges() {
	for { // We want to restart listening in case something breaks.
		outputs, err := eliona.ListenForOutputChanges()
		if err != nil {
			log.Error("eliona", "listening for output changes: %v", err)
			return
		}
		for output := range outputs {
			// TODO: Filter for only own asset types.
			height, ok := output.Data["height"]
			log.Info("output", "%+v\n%+v", output.Data, height)
			if !ok {
				log.Debug("eliona", "no 'height' attribute in data: %+v", output)
				continue
			}
			if err := conf.SetFloorHeight(output.AssetId, height.(float64)); err != nil {
				log.Error("conf", "setting floor height: %v", err)
				continue
			}
		}
		time.Sleep(time.Second * 5) // Give the server a little break.
	}
}

// listenApi starts the API server and listen for requests
func listenApi() {
	err := http.ListenAndServe(":"+common.Getenv("API_SERVER_PORT", "3000"), utilshttp.NewCORSEnabledHandler(
		apiserver.NewRouter(
			apiserver.NewConfigurationApiController(apiservices.NewConfigurationApiService()),
			apiserver.NewVersionApiController(apiservices.NewVersionApiService()),
			apiserver.NewCustomizationApiController(apiservices.NewCustomizationApiService()),
		)),
	)
	log.Fatal("main", "API server: %v", err)
}

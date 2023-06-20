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

	api "github.com/eliona-smart-building-assistant/go-eliona-api-client/v2"
	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/common"
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

func isLocation(assetType string) bool {
	return assetType == RoomAssetType || assetType == FloorAssetType || assetType == BuildingAssetType
}

func isTracker(assetType string) bool {
	return assetType == TagAssetType || assetType == BadgeAssetType
}

func createAssetIfNecessary(config apiserver.Configuration, projectId string, id string, parentId *int32, assetType string, name string) (int32, error) {
	assetData := assetData{
		config:                  config,
		projectId:               projectId,
		parentLocationalAssetId: parentId,
		identifier:              assetType + id,
		assetType:               assetType,
		name:                    name,
		description:             fmt.Sprintf("%s (%v)", name, id),
	}
	_, assetID, err := upsertAsset(assetData)
	if err != nil {
		return 0, fmt.Errorf("creating asset for %s %s and project %v: %v", assetType, name, projectId, err)
	}
	return assetID, nil
}

func CreateLocationAssetsIfNecessary(config apiserver.Configuration, rooms []kontaktio.Room) error {
	for _, projectId := range conf.ProjIds(config) {
		rootAssetID, err := createAssetIfNecessary(config, projectId, "", nil, RootAssetType, "Kontakt.io")
		if err != nil {
			return err
		}
		for _, room := range rooms {
			buildingAssetID, err := createAssetIfNecessary(config, projectId, fmt.Sprint(room.Floor.Building.ID), &rootAssetID, BuildingAssetType, room.Floor.Building.Name)
			if err != nil {
				return err
			}
			floorAssetID, err := createAssetIfNecessary(config, projectId, fmt.Sprint(room.Floor.ID), &buildingAssetID, FloorAssetType, room.Floor.Name)
			if err != nil {
				return err
			}
			if _, err := createAssetIfNecessary(config, projectId, fmt.Sprint(room.ID), &floorAssetID, RoomAssetType, room.Name); err != nil {
				return err
			}
		}
	}
	return nil
}

func CreateDeviceAssetsIfNecessary(config apiserver.Configuration, devices []kontaktio.Device) error {
	for _, projectId := range conf.ProjIds(config) {
		for _, device := range devices {
			_, err := createAssetIfNecessary(config, projectId, device.ID, nil, device.Type, device.Name)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type assetData struct {
	config                  apiserver.Configuration
	projectId               string
	parentFunctionalAssetId *int32
	parentLocationalAssetId *int32
	identifier              string
	assetType               string
	name                    string
	description             string
}

func upsertAsset(d assetData) (created bool, assetID int32, err error) {
	// Get known asset id from configuration
	currentAssetID, err := conf.GetTagAssetId(context.Background(), d.config, d.projectId, d.identifier)
	if isLocation(d.assetType) {
		currentAssetID, err = conf.GetLocationAssetId(context.Background(), d.config, d.projectId, d.identifier)
	}
	if err != nil {
		return false, 0, fmt.Errorf("finding asset ID: %v", err)
	}
	if currentAssetID != nil {
		return false, *currentAssetID, nil
	}

	a := api.Asset{
		ProjectId:               d.projectId,
		GlobalAssetIdentifier:   d.identifier,
		Name:                    *api.NewNullableString(common.Ptr(d.name)),
		AssetType:               d.assetType,
		Description:             *api.NewNullableString(common.Ptr(d.description)),
		ParentFunctionalAssetId: *api.NewNullableInt32(d.parentFunctionalAssetId),
		ParentLocationalAssetId: *api.NewNullableInt32(d.parentLocationalAssetId),
		IsTracker:               *api.NewNullableBool(common.Ptr(isTracker(d.assetType))),
	}
	newID, err := asset.UpsertAsset(a)
	if err != nil {
		return false, 0, fmt.Errorf("upserting asset %+v into Eliona: %v", a, err)
	}
	if newID == nil {
		return false, 0, fmt.Errorf("cannot create asset %s", d.name)
	}

	// Remember the asset id for further usage
	if !isLocation(d.assetType) {
		if err := conf.InsertDevice(context.Background(), d.config, d.projectId, d.identifier, *newID); err != nil {
			return false, 0, fmt.Errorf("inserting asset to config db: %v", err)
		}
	} else {
		if err := conf.InsertLocation(context.Background(), d.config, d.projectId, d.identifier, *newID); err != nil {
			return false, 0, fmt.Errorf("inserting asset to config db: %v", err)
		}
	}

	log.Debug("eliona", "Created new asset for project %s and device %s.", d.projectId, d.identifier)

	return true, *newID, nil
}

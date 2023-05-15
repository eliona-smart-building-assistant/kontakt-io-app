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

	"github.com/eliona-smart-building-assistant/go-eliona/asset"
	"github.com/eliona-smart-building-assistant/go-utils/db"
)

// InitEliona initializes the app in eliona
func InitEliona(connection db.Connection) error {
	if err := asset.InitAssetTypeFile("eliona/asset-type-building.json")(connection); err != nil {
		return fmt.Errorf("init building asset type: %v", err)
	}
	if err := asset.InitAssetTypeFile("eliona/asset-type-floor.json")(connection); err != nil {
		return fmt.Errorf("init floor asset type: %v", err)
	}
	if err := asset.InitAssetTypeFile("eliona/asset-type-room.json")(connection); err != nil {
		return fmt.Errorf("init room asset type: %v", err)
	}
	if err := asset.InitAssetTypeFile("eliona/asset-type-tag.json")(connection); err != nil {
		return fmt.Errorf("init tag asset type: %v", err)
	}
	if err := asset.InitAssetTypeFile("eliona/asset-type-beacon.json")(connection); err != nil {
		return fmt.Errorf("init beacon asset type: %v", err)
	}
	if err := asset.InitAssetTypeFile("eliona/asset-type-badge.json")(connection); err != nil {
		return fmt.Errorf("init badge asset type: %v", err)
	}
	return nil
}

package modrinth

import (
	"errors"

	"github.com/comp500/packwiz/core"
	"github.com/mitchellh/mapstructure"
)

type mrUpdateData struct {
	ModID            string `mapstructure:"mod-id"`
	InstalledVersion string `mapstructure:"version"`
}

func (u mrUpdateData) ToMap() (map[string]interface{}, error) {
	newMap := make(map[string]interface{})
	err := mapstructure.Decode(u, &newMap)
	return newMap, err
}

type mrUpdater struct{}

func (u mrUpdater) ParseUpdate(updateUnparsed map[string]interface{}) (interface{}, error) {
	var updateData mrUpdateData
	err := mapstructure.Decode(updateUnparsed, &updateData)
	return updateData, err
}

type cachedStateStore struct {
	ModID string
	Version Version
}

func (u mrUpdater) CheckUpdate(mods []core.Mod, mcVersion string) ([]core.UpdateCheck, error) {
	results := make([]core.UpdateCheck, len(mods))

	pack, err := core.LoadPack()
	if err != nil {
		return results, err
	}

	for i, mod := range mods {
		rawData, ok := mod.GetParsedUpdateData("modrinth")
		if !ok {
			results[i] = core.UpdateCheck{Error: errors.New("couldn't parse mod data")}
			continue
		}

		data := rawData.(mrUpdateData)

		newVersion, err := getLatestVersion(data.ModID, pack)
		if err != nil {
			results[i] = core.UpdateCheck{Error: err}
			continue
		}

		if newVersion.ID == "" { //There is no version available for this minecraft version or loader.
			results[i] = core.UpdateCheck{UpdateAvailable: false}
        	continue
		}

		if newVersion.ID == data.InstalledVersion { //The latest version from the site is the same as the installed one
			results[i] = core.UpdateCheck{UpdateAvailable: false}
			continue
		}

		if len(newVersion.Files) == 0 {
			results[i] = core.UpdateCheck{Error: errors.New("new version doesn't have any files")}
			continue
		}

		results[i] = core.UpdateCheck{
			UpdateAvailable: true,
			UpdateString:    mod.FileName + " -> " + newVersion.Files[0].Filename,
			CachedState:     cachedStateStore{data.ModID, newVersion},
		}
	}

	return results, nil
}

func (u mrUpdater) DoUpdate(mods []*core.Mod, cachedState []interface{}) error {
	for i, mod := range mods {
		modState := cachedState[i].(cachedStateStore)
		var version = modState.Version

		var file = version.Files[0]

		algorithm, hash := file.getBestHash()
		if algorithm == "" {
			return errors.New("file for mod "+mod.Name+" doesn't have a hash")
		}

		mod.FileName = file.Filename
		mod.Download = core.ModDownload {
			URL:        file.Url,
			HashFormat: algorithm,
			Hash:       hash,
		}
		mod.Update["modrinth"]["version"] = version.ID
	}

	return nil
}
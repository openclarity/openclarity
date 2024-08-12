// Copyright © 2024 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package windows

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/openclarity/vmclarity/core/to"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"

	"www.velocidex.com/golang/regparser"
)

// Documentation about registries, its structure and further details can be found at:
//	- https://en.wikipedia.org/wiki/Windows_Registry
//	- https://github.com/msuhanov/regf/blob/master/Windows%20registry%20file%20format%20specification.md
//	- https://techdirectarchive.com/2020/02/07/how-to-check-if-windows-updates-were-installed-on-your-device-via-the-registry-editor/
//	- https://jgmes.com/webstart/library/qr_windowsxp.htm#:~:text=In%20Windows%20XP%2C%20the%20registry,corresponding%20location%20of%20each%20hive.
//
// System SOFTWARE registry keys accessed:
//	- system info: 		Microsoft\Windows NT\CurrentVersion
//	- updates: 			Microsoft\Windows\CurrentVersion\Component Based Servicing\Packages\*
//	- profiles: 		Microsoft\Windows NT\CurrentVersion\ProfileList\*
//	- system apps: 		Microsoft\Windows\CurrentVersion\Uninstall\*
//	- system apps:		Wow6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*
//	- system apps:		WOW6432Node\Microsoft\Windows\CurrentVersion\Uninstall\*
//
// User NTUSER.DAT registry keys accessed:
//	- user apps: 		SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\*

var defaultRegistryRootPaths = []string{
	"/Windows/System32/config/SOFTWARE", // Windows Vista and newer
	"/WINDOWS/system32/config/software", // Windows XP and older
}

type Registry struct {
	mountPath   string              // root path to Windows drive
	softwareReg *regparser.Registry // HKEY_LOCAL_MACHINE/SOFTWARE registry
	cleanup     func() error
	logger      *log.Entry
}

func NewRegistryForMount(mountPath string, logger *log.Entry) (*Registry, error) {
	// The registry key structure is identical for all Windows NT distributions, so
	// try all registry combinations. If the registry is not found under any default
	// paths, it might be a custom system installation or unsupported version.
	var errs error
	for _, defaultRootPath := range defaultRegistryRootPaths {
		registryFilePath := path.Join(mountPath, defaultRootPath)
		registry, err := NewRegistry(registryFilePath, logger)
		if err == nil {
			return registry, nil // found, return
		}
		errs = errors.Join(errs, err) // collect errors, might be file-related
	}

	return nil, fmt.Errorf("cannot find registry in mount %s: %w", mountPath, errs)
}

func NewRegistry(registryFilePath string, logger *log.Entry) (*Registry, error) {
	// Use filepath clean to ensure path is platform-independent
	registryFile, err := os.Open(filepath.Clean(registryFilePath))
	if err != nil {
		return nil, fmt.Errorf("cannot open registry file: %w", err)
	}

	// Registry file must remain open as it is read on-the-fly
	softwareReg, err := regparser.NewRegistry(registryFile)
	if err != nil {
		return nil, fmt.Errorf("cannot create registry reader: %w", err)
	}

	// Extract mount path from registry path by removing the default root paths. For
	// registry under /var/snapshot/Windows/System32/config/SOFTWARE, this should
	// extract /var/snapshot mount path
	mountPath := registryFilePath
	for _, defaultRootPath := range defaultRegistryRootPaths {
		mountPath = strings.TrimSuffix(mountPath, defaultRootPath)
	}

	return &Registry{
		mountPath:   mountPath,
		softwareReg: softwareReg,
		cleanup:     registryFile.Close,
		logger:      logger,
	}, nil
}

// Close needs to be called when done to free up resources. Registry is not
// usable once closed.
func (r *Registry) Close() error {
	if err := r.cleanup(); err != nil {
		return fmt.Errorf("unable to close registry: %w", err)
	}
	return nil
}

// GetPlatform returns OS-specific data from the registry.
func (r *Registry) GetPlatform() (map[string]string, error) {
	// Open key to fetch operating system version and configuration data
	platformKey, err := openKey(r.softwareReg, "Microsoft/Windows NT/CurrentVersion")
	if err != nil {
		return nil, err
	}

	// Extract all platform data from the registry
	platform := getValuesMap(platformKey)

	// Strip information about the product key hash
	delete(platform, "DigitalProductId")
	delete(platform, "DigitalProductId4")

	return platform, nil
}

// GetUpdates returns a slice of all installed system updates from the registry.
func (r *Registry) GetUpdates() ([]string, error) {
	// Open key to fetch CBS data about packages (updates and components)
	packagesKey, err := openKey(r.softwareReg, "Microsoft/Windows/CurrentVersion/Component Based Servicing/Packages")
	if err != nil {
		return nil, err
	}

	// Extract all updates from installed packages
	updates := map[string]string{}
	updateRegex := regexp.MustCompile("KB[0-9]{7,}")
	for _, pkgKey := range packagesKey.Subkeys() {
		pkgName := pkgKey.Name()
		pkgValues := getValuesMap(pkgKey)

		// Ignore packages that were not installed as system components or via updates
		_, isComponent := pkgValues["InstallClient"]
		_, isUpdate := pkgValues["UpdateAgentLCU"]
		if !isComponent && !isUpdate {
			continue
		}

		// Install location value for a given package can contain update identifier such
		// as "C:\Windows\CbsTemp\31075171_2144217839\Windows10.0-KB5032189-x64.cab\"
		if location, ok := pkgValues["InstallLocation"]; ok {
			if kb := updateRegex.FindString(location); kb != "" {
				updates[kb] = kb
			}
		}

		// If the installed package contains state value, it indicates a potential system
		// update. We are only curious about "112" state code which translates to
		// successful package installation. When this is the case, package registry key
		// contains update identifier such as "Package_10_for_KB5011048..."
		if state, ok := pkgValues["CurrentState"]; ok && state == "112" {
			if kb := updateRegex.FindString(pkgName); kb != "" {
				updates[kb] = kb
			}
		}
	}

	return to.Keys(updates), nil
}

// GetUsersApps returns installed apps from the registry for all users. This
// method will not error in case user profile cannot be loaded. It only errors if
// the system registry cannot be accessed to get the list of user profiles.
func (r *Registry) GetUsersApps() ([]map[string]string, error) {
	// Open key to fetch system user profiles in order to get profile dir paths
	profilesKey, err := openKey(r.softwareReg, "Microsoft/Windows NT/CurrentVersion/ProfileList")
	if err != nil {
		return nil, err
	}

	// Extract all installed applications for each user
	apps := []map[string]string{}
	for _, profileKey := range profilesKey.Subkeys() {
		// Run in a function to allow cleanup
		func(profileValues map[string]string) {
			// Extract profile path from the registry key values. The path is
			// Windows-specific, but the mount path must be Unix-specific.
			profileLocation, ok := profileValues["ProfileImagePath"]
			if !ok {
				return // silent skip, not a user profile
			}
			profileLocation = strings.ReplaceAll(profileLocation, "\\", "/")

			// The actual user location in the registry is specified as "C:\Users\...".
			// However, due to the mount location, the actual path could be
			// "/var/mounts/Users/...". Strip everything before the "/Users/" to construct a
			// valid mount path.
			if prefixIdx := strings.Index(profileLocation, "/Users/"); prefixIdx >= 0 {
				baseProfileLocation := profileLocation[prefixIdx:]
				profileLocation = path.Join(r.mountPath, baseProfileLocation)
			} else {
				return // silent skip, not a user profile
			}

			// Open profile registry file to access profile-specific registry.
			// Use filepath clean to ensure path is platform-independent.
			profileRegPath := path.Join(profileLocation, "NTUSER.DAT")
			profileRegFile, err := os.Open(filepath.Clean(profileRegPath))
			if err != nil {
				r.logger.Warnf("failed to open user profile: %v", err)
				return
			}
			defer profileRegFile.Close()

			profileReg, err := regparser.NewRegistry(profileRegFile)
			if err != nil {
				r.logger.Warnf("failed to create user registry reader: %v", err)
				return
			}

			// Open key to fetch installed profile apps
			profileAppsKey, err := openKey(profileReg, "SOFTWARE/Microsoft/Windows/CurrentVersion/Uninstall")
			if err != nil {
				r.logger.Warnf("failed to open key: %v", err)
				return
			}

			// Extract all apps from user registry key. When the application registry key
			// values contain application name, add them to the result.
			for _, appKey := range profileAppsKey.Subkeys() {
				appValues := getValuesMap(appKey)
				if _, ok := appValues["DisplayName"]; ok {
					apps = append(apps, appValues)
				}
			}
		}(getValuesMap(profileKey))
	}

	return apps, nil
}

// GetSystemApps returns installed system-wide apps from the registry.
func (r *Registry) GetSystemApps() ([]map[string]string, error) {
	// Try multiple keys to fetch installed system apps
	apps := []map[string]string{}
	for _, appsKey := range []string{
		"Microsoft/Windows/CurrentVersion/Uninstall",             // for newer Windows NT
		"Wow6432Node/Microsoft/Windows/CurrentVersion/Uninstall", // store for 32-bit apps on 64-bit systems
		"WOW6432Node/Microsoft/Windows/CurrentVersion/Uninstall", // same as before, resolves compatibility issues
	} {
		appsKey, err := openKey(r.softwareReg, appsKey)
		if err != nil {
			r.logger.Warnf("failed to get installed system apps: %v", err)
			continue
		}

		// Extract all apps from system registry. When the application registry key
		// values contain application name, add them to the result.
		for _, appKey := range appsKey.Subkeys() {
			appValues := getValuesMap(appKey)
			if _, ok := appValues["DisplayName"]; ok {
				apps = append(apps, appValues)
			}
		}
	}

	return apps, nil
}

// GetBOM returns cyclone database from the registry data.
// TODO(ramizpolic): Other registry fetch methods should return a struct instead of maps.
func (r *Registry) GetBOM() (*cdx.BOM, error) {
	bom := cdx.NewBOM()

	// Inject platform data to BOM
	{
		// Get platform registry data
		platformData, err := r.GetPlatform()
		if err != nil {
			return nil, fmt.Errorf("unable to get platform data: %w", err)
		}

		// Extract manufacturer if available
		var manufacturer *cdx.OrganizationalEntity
		if name, ok := platformData["SystemManufacturer"]; ok {
			manufacturer = &cdx.OrganizationalEntity{
				Name: name,
			}
		}

		// Set immutable serial number from the unique installation ID
		bom.SerialNumber = uuid.NewSHA1(uuid.Nil, []byte(platformData["ProductId"])).URN()

		// Set BOM metadata
		bom.Metadata = &cdx.Metadata{
			Timestamp: platformData["InstallDate"],
			Component: &cdx.Component{
				Type: cdx.ComponentTypeOS,
				Name: platformData["ProductName"], // Windows 10 Pro
				Version: fmt.Sprintf("%s.%s.%s", // 10.0.22000
					platformData["CurrentMajorVersionNumber"],
					platformData["CurrentMinorVersionNumber"],
					platformData["CurrentBuildNumber"],
				),
			},
			Manufacture: manufacturer,
			Supplier: &cdx.OrganizationalEntity{
				Name: "Microsoft Corporation",
			},
			Properties: &[]cdx.Property{
				{
					Name:  "analyzers",
					Value: AnalyzerName,
				},
			},
		}
	}

	// Inject updates, system and user apps to BOM
	{
		var components []cdx.Component

		// Add applications to cyclonedx components
		systemApps, err := r.GetSystemApps()
		if err != nil {
			return nil, fmt.Errorf("unable to get system apps: %w", err)
		}

		usersApps, err := r.GetUsersApps()
		if err != nil {
			return nil, fmt.Errorf("unable to get users apps: %w", err)
		}

		apps := append(systemApps, usersApps...)
		for _, app := range apps {
			components = append(components, cdx.Component{
				Type:      cdx.ComponentTypeApplication,
				Publisher: app["Publisher"],
				Name:      app["DisplayName"],
				Version:   app["DisplayVersion"],
			})
		}

		// Add updates to cyclonedx components. This is optional since Windows XP stores
		// update data in the system package registry.
		//
		// TODO(ramizpolic): Error skip should only happen when we know it is XP,
		// otherwise this should continue regularly.
		systemUpdates, err := r.GetUpdates()
		if err == nil {
			for _, update := range systemUpdates {
				components = append(components, cdx.Component{
					Type:      cdx.ComponentTypeApplication,
					Publisher: "Microsoft Corporation",
					Name:      update,
				})
			}
		} else {
			r.logger.Warnf("could not fetch updates from registry: %v", err)
		}

		// Set BOM components
		bom.Components = &components
	}

	return bom, nil
}

// openKey opens a given registry key from the given registry or returns an error.
// Returned key can have multiple sub-keys and values specified.
func openKey(registry *regparser.Registry, key string) (*regparser.CM_KEY_NODE, error) {
	keyNode := registry.OpenKey(key)
	if keyNode == nil {
		return nil, fmt.Errorf("cannot open key %s", key)
	}
	return keyNode, nil
}

// getValuesMap returns all registry key values for a given registry key as a map
// of value name and its data.
func getValuesMap(key *regparser.CM_KEY_NODE) map[string]string {
	valuesMap := map[string]string{}
	for _, keyValue := range key.Values() {
		valuesMap[keyValue.ValueName()] = convertKVData(keyValue.ValueData())
	}
	return valuesMap
}

// convertKVData returns the registry key value data as a valid string.
func convertKVData(value *regparser.ValueData) string {
	switch value.Type {
	case regparser.REG_SZ, regparser.REG_EXPAND_SZ: // null-terminated string
		return strings.TrimRightFunc(value.String, func(r rune) bool {
			return r == 0
		})

	case regparser.REG_MULTI_SZ: // multi-part string
		return strings.Join(value.MultiSz, " ")

	case regparser.REG_DWORD, regparser.REG_DWORD_BIG_ENDIAN, regparser.REG_QWORD: // unsigned 32/64-bit value
		return strconv.FormatUint(value.Uint64, 10)

	case regparser.REG_BINARY: // non-stringable binary value
		// Return as hex to preserve buffer; we don't really care about this value
		return fmt.Sprintf("%X", value.Data)

	case
		regparser.REG_LINK,                       // unicode symbolic link
		regparser.REG_RESOURCE_LIST,              // device-driver resource list
		regparser.REG_FULL_RESOURCE_DESCRIPTOR,   // hardware setting
		regparser.REG_RESOURCE_REQUIREMENTS_LIST, // hardware resource list
		regparser.REG_UNKNOWN:                    // no-type
		fallthrough

	default:
		return ""
	}
}

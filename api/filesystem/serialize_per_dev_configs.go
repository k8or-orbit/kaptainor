package filesystem

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	portainer "github.com/portainer/portainer/api"
)

type MultiFilterArgs []struct {
	FilterKey  string
	FilterType portainer.PerDevConfigsFilterType
}

// MultiFilterDirForPerDevConfigs filers the given dirEntries with multiple filter args, returns the merged entries for the given device
func MultiFilterDirForPerDevConfigs(dirEntries []DirEntry, configPath string, multiFilterArgs MultiFilterArgs) []DirEntry {
	var filteredDirEntries []DirEntry

	for _, multiFilterArg := range multiFilterArgs {
		tmp := FilterDirForPerDevConfigs(dirEntries, multiFilterArg.FilterKey, configPath, multiFilterArg.FilterType)
		filteredDirEntries = append(filteredDirEntries, tmp...)
	}

	return deduplicate(filteredDirEntries)
}

func deduplicate(dirEntries []DirEntry) []DirEntry {
	var deduplicatedDirEntries []DirEntry

	marks := make(map[string]struct{})

	for _, dirEntry := range dirEntries {
		_, ok := marks[dirEntry.Name]
		if !ok {
			marks[dirEntry.Name] = struct{}{}
			deduplicatedDirEntries = append(deduplicatedDirEntries, dirEntry)
		}
	}

	return deduplicatedDirEntries
}

// FilterDirForPerDevConfigs filers the given dirEntries, returns entries for the given device
// For given configPath A/B/C, return entries:
//  1. all entries outside of dir A/B/C
//  2. For filterType file:
//     file entries: A/B/C/<deviceName> and A/B/C/<deviceName>.*
//  3. For filterType dir:
//     dir entry:   A/B/C/<deviceName>
//     all entries: A/B/C/<deviceName>/*
func FilterDirForPerDevConfigs(dirEntries []DirEntry, deviceName, configPath string, filterType portainer.PerDevConfigsFilterType) []DirEntry {
	var filteredDirEntries []DirEntry

	for _, dirEntry := range dirEntries {
		if shouldIncludeEntry(dirEntry, deviceName, configPath, filterType) {
			filteredDirEntries = append(filteredDirEntries, dirEntry)
		}
	}

	return filteredDirEntries
}

func shouldIncludeEntry(dirEntry DirEntry, deviceName, configPath string, filterType portainer.PerDevConfigsFilterType) bool {

	// Include all entries outside of dir A
	if !isInConfigDir(dirEntry, configPath) {
		return true
	}

	if filterType == portainer.PerDevConfigsTypeFile {
		// Include file entries A/B/C/<deviceName> or A/B/C/<deviceName>.*
		return shouldIncludeFile(dirEntry, deviceName, configPath)
	}

	if filterType == portainer.PerDevConfigsTypeDir {
		// Include:
		// dir entry A/B/C/<deviceName>
		// all entries A/B/C/<deviceName>/*
		return shouldIncludeDir(dirEntry, deviceName, configPath)
	}

	return false
}

func isInConfigDir(dirEntry DirEntry, configPath string) bool {
	// return true if entry name starts with "A/B"
	return strings.HasPrefix(dirEntry.Name, appendTailSeparator(configPath))
}

func shouldIncludeFile(dirEntry DirEntry, deviceName, configPath string) bool {
	if !dirEntry.IsFile {
		return false
	}

	// example: A/B/C/<deviceName>
	filterEqual := filepath.Join(configPath, deviceName)

	// example: A/B/C/<deviceName>/
	filterPrefix := filterEqual + "."

	// include file entries: A/B/C/<deviceName> or A/B/C/<deviceName>.*
	return dirEntry.Name == filterEqual || strings.HasPrefix(dirEntry.Name, filterPrefix)
}

func shouldIncludeDir(dirEntry DirEntry, deviceName, configPath string) bool {
	// example: A/B/C/'/<deviceName>
	filterEqual := filepath.Join(configPath, deviceName)

	// example: A/B/C/<deviceName>/
	filterPrefix := appendTailSeparator(filterEqual)

	// include dir entry: A/B/C/<deviceName>
	if !dirEntry.IsFile && dirEntry.Name == filterEqual {
		return true
	}

	// include all entries A/B/C/<deviceName>/*
	return strings.HasPrefix(dirEntry.Name, filterPrefix)
}

func appendTailSeparator(path string) string {
	return fmt.Sprintf("%s%c", path, os.PathSeparator)
}

package persistence

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"sort"
)

var allStorages = make(map[string]Storage)

// Storages returns a list containing the names of all supported storages.
func Storages() []string {
	keys := make([]string, 0, len(allStorages))
	for k := range allStorages {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// RegisterStorage registers a new storage.
func RegisterStorage(name string, storage Storage) {
	allStorages[name] = storage
}

// GetStorage returns the storage with the given name.
func GetStorage(name string) Storage {
	if storage, ok := allStorages[name]; ok {
		return storage
	}
	return nil
}

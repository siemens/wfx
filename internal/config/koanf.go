package config

/*
 * SPDX-FileCopyrightText: 2023 Siemens AG
 *
 * SPDX-License-Identifier: Apache-2.0
 *
 * Author: Michael Adler <michael.adler@siemens.com>
 */

import (
	"sync"

	"github.com/knadh/koanf/v2"
)

// ThreadSafeKoanf is a thread-safe wrapper around Koanf.
type ThreadSafeKoanf struct {
	instance *koanf.Koanf
	lock     sync.RWMutex
}

func New() *ThreadSafeKoanf {
	result := new(ThreadSafeKoanf)
	result.instance = koanf.New(".")
	return result
}

// Read provides thread-safe read access to the Koanf instance.
func (t *ThreadSafeKoanf) Read(cb func(k *koanf.Koanf)) {
	t.lock.RLock()
	defer t.lock.RUnlock()
	cb(t.instance)
}

// Write provides thread-safe write access to the Koanf instance.
func (t *ThreadSafeKoanf) Write(cb func(k *koanf.Koanf)) {
	t.lock.Lock()
	defer t.lock.Unlock()
	cb(t.instance)
}

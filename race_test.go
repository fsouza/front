// Copyright 2013 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build race

package main

import (
	"sync"
	"testing"
)

func TestLoadRulesIsSafe(t *testing.T) {
	var wg sync.WaitGroup
	n := 5
	s := Server{}
	var load = func() {
		s.loadRules("testdata/rules.json")
		wg.Done()
	}
	wg.Add(n)
	for i := 0; i < n; i++ {
		go load()
	}
	wg.Wait()
}

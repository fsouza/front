// Copyright 2013 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"os"
	"sync"
)

type Rule struct {
	Domain  string
	Backend string
}

type Server struct {
	rules []Rule
	rmut  sync.RWMutex
}

func (s *Server) LoadRules(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	s.rmut.Lock()
	err = json.NewDecoder(f).Decode(&s.rules)
	s.rmut.Unlock()
	if err != nil {
		return &invalidRuleError{err}
	}
	return nil
}

type invalidRuleError struct {
	err error
}

func (e *invalidRuleError) Error() string {
	return "Invalid rule file: " + e.err.Error()
}

func main() {}

// Copyright 2013 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"os"
)

type Rule struct {
	Domain  string
	Backend string
}

type Server struct {
	rules []Rule
}

func (s *Server) LoadRules(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(&s.rules)
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

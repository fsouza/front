// Copyright 2013 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"reflect"
	"testing"
)

func TestLoadRules(t *testing.T) {
	s := Server{}
	err := s.LoadRules("testdata/rules.json")
	if err != nil {
		t.Fatal(err)
	}
	want := []Rule{
		{Domain: "souza.cc", Backend: "localhost:3232"},
		{Domain: "golang.org", Backend: "localhost:3131"},
		{Domain: "globo.com", Backend: "localhost:3030"},
	}
	if !reflect.DeepEqual(s.rules, want) {
		t.Errorf("LoadRules:\nWant %#v.\nGot %#v.", want, s.rules)
	}
}

func TestLoadRulesFailures(t *testing.T) {
	tests := []struct {
		filename string
		msg      string
	}{
		{"testdata/invalidrules.json", "Invalid rule file: invalid character '}' looking for beginning of object key string"},
		{"testdata/file-not-found.json", "open testdata/file-not-found.json: no such file or directory"},
	}
	s := Server{}
	for _, tt := range tests {
		err := s.LoadRules(tt.filename)
		if err == nil {
			t.Errorf("LoadRules(%q): Want %q. Got %v.", tt.filename, tt.msg, nil)
			continue
		}
		if err.Error() != tt.msg {
			t.Errorf("LoadRules(%q): Want %q. Got %q.", tt.filename, tt.msg, err.Error())
		}
	}
}

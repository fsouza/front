// Copyright 2013 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"github.com/fsouza/lb"
	"testing"
)

func TestLoadRules(t *testing.T) {
	s := Server{}
	rules, err := s.loadRules("testdata/rules.json")
	if err != nil {
		t.Fatal(err)
	}
	b1, _ := lb.NewLoadBalancer("localhost:3232")
	b2, _ := lb.NewLoadBalancer("localhost:3131")
	b3, _ := lb.NewLoadBalancer("localhost:3030", "localhost:2929", "localhost:2121")
	want := []Rule{
		{Domain: "souza.cc", Backend: b1},
		{Domain: "golang.org", Backend: b2},
		{Domain: "globo.com", Backend: b3},
	}
	for i, rule := range rules {
		wanted := want[i]
		if wanted.Domain != rule.Domain {
			t.Errorf("LoadRules: Wrong domain. Want %q. Got %q", wanted.Domain, rule.Domain)
		}
		if rule.Backend == nil {
			t.Errorf("LoadRules. Wante non-nil load balancer. Got <nil>.")
		}
	}
}

func TestLoadRulesFailures(t *testing.T) {
	tests := []struct {
		filename string
		msg      string
	}{
		{"testdata/invalidrules.json", "Invalid rule file: invalid character '}' looking for beginning of object key string"},
		{"testdata/file-not-found.json", "open testdata/file-not-found.json: no such file or directory"},
		{"testdata/invaliddomain.json", "Invalid rule file: parse http://%%%%: hexadecimal escape in host"},
	}
	s := Server{}
	for _, tt := range tests {
		rules, err := s.loadRules(tt.filename)
		if rules != nil {
			t.Errorf("LoadRules(%q). Want <nil>. Got %#v.", tt.filename, rules)
		}
		if err == nil {
			t.Errorf("LoadRules(%q): Want %q. Got %v.", tt.filename, tt.msg, nil)
			continue
		}
		if err.Error() != tt.msg {
			t.Errorf("LoadRules(%q): Want %q. Got %q.", tt.filename, tt.msg, err.Error())
		}
	}
}

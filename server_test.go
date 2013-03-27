// Copyright 2013 Francisco Souza. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"github.com/fsouza/lb"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	rules := []jsonRule{
		{Domain: "souza.cc", Backends: []string{"localhost:3232"}},
		{Domain: "golang.org", Backends: []string{"localhost:3131"}},
		{Domain: "globo.com", Backends: []string{"localhost:3030", "localhost:2929", "localhost:2121"}},
		{Domain: "rust-lang.org", Backends: []string{"10.10.10.10:8080"}},
	}
	exec.Command("cp", "testdata/rules.json", "/tmp/rules-server.json").Run()
	defer exec.Command("rm", "/tmp/rules-server.json").Run()
	s, err := NewServer("/tmp/rules-server.json")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"souza.cc", "golang.org", "globo.com"}
	for i, w := range want {
		if s.rules[i].Domain != w {
			t.Errorf("NewServer: Wrong domain. Want %q. Got %q.", w, s.rules[i])
		}
	}
	f, err := os.OpenFile("/tmp/rules-server.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	if err = json.NewEncoder(f).Encode(rules); err != nil {
		t.Fatal(err)
	}
	f.Close()
	time.Sleep(1e6)
	want = append(want, "rust-lang.org")
	if len(want) != len(s.rules) {
		t.Fatalf("NewServer did not watch the file for changes. Want %#v. Got %#v.", want, s.rules)
	}
	for i, w := range want {
		if s.rules[i].Domain != w {
			t.Errorf("NewServer: Wrong domain. Want %q. Got %q.", w, s.rules[i])
		}
	}
}

func TestNewServerWatchingFileWithInvalidJSON(t *testing.T) {
	exec.Command("cp", "testdata/rules.json", "/tmp/rules-server.json").Run()
	defer exec.Command("rm", "/tmp/rules-server.json").Run()
	s, err := NewServer("/tmp/rules-server.json")
	if err != nil {
		t.Fatal(err)
	}
	old := s.rules
	f, err := os.OpenFile("/tmp/rules-server.json", os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		t.Fatal(err)
	}
	f.Write([]byte("------"))
	f.Close()
	time.Sleep(1e6)
	if !reflect.DeepEqual(s.rules, old) {
		t.Errorf("NewServer(%q) reloaded broken file. Want %#v. Got %#v.", "/tmp/rules-server.json", old, s.rules)
	}
}

func TestNewServerErrors(t *testing.T) {
	tests := []struct {
		filename string
		msg      string
	}{
		{"testdata/invalidrules.json", "Invalid rule file: invalid character '}' looking for beginning of object key string"},
		{"testdata/file-not-found.json", "open testdata/file-not-found.json: no such file or directory"},
		{"testdata/invaliddomain.json", "Invalid rule file: parse http://%%%%: hexadecimal escape in host"},
	}
	for _, tt := range tests {
		server, err := NewServer(tt.filename)
		if server != nil {
			t.Errorf("NewServer(%q). Want <nil>. Got %#v.", tt.filename, server)
		}
		if err == nil {
			t.Errorf("NewServer(%q): Want %q. Got %v.", tt.filename, tt.msg, nil)
			continue
		}
		if err.Error() != tt.msg {
			t.Errorf("NewServer(%q): Want %q. Got %q.", tt.filename, tt.msg, err.Error())
		}
	}
}

func TestServeHTTP(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	}))
	defer server1.Close()
	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world from server 2!"))
	}))
	defer server2.Close()
	b1, err := lb.NewLoadBalancer(server1.URL)
	if err != nil {
		t.Fatal(err)
	}
	b2, err := lb.NewLoadBalancer(server2.URL)
	if err != nil {
		t.Fatal(err)
	}
	s := Server{
		rules: []Rule{
			{Domain: "souza.cc", Backend: b1},
			{Domain: "globo.com", Backend: b2},
		},
	}
	var tests = []struct {
		host string
		body string
	}{
		{"souza.cc", "Hello world!"},
		{"f.souza.cc", "Hello world!"},
		{"globo.com", "Hello world from server 2!"},
		{"www.globo.com", "Hello world from server 2!"},
		{"g1.globo.com", "Hello world from server 2!"},
	}
	for _, tt := range tests {
		request, _ := http.NewRequest("GET", "/", nil)
		request.Header.Set("Host", tt.host)
		recorder := httptest.NewRecorder()
		s.ServeHTTP(recorder, request)
		if recorder.Code != http.StatusOK {
			t.Errorf("ServerHTTP returned wrong status. Want %d. Got %d.", http.StatusOK, recorder.Code)
		}
		if recorder.Body.String() != tt.body {
			t.Errorf("ServeHTTP did not return the proper body. Want %q. Got %q.", tt.body, recorder.Body.String())
		}
	}
}

func TestServeHTTPErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello world!"))
	}))
	defer server.Close()
	b, _ := lb.NewLoadBalancer(server.URL)
	s := Server{rules: []Rule{{Domain: "souza.cc", Backend: b}}}
	var tests = []struct {
		host string
		code int
		body string
	}{
		{"", http.StatusBadRequest, "Missing Host header\n"},
		{"globo.com", http.StatusNotFound, "Page not found\n"},
	}
	for _, tt := range tests {
		request, _ := http.NewRequest("GET", "/", nil)
		request.Header.Set("Host", tt.host)
		recorder := httptest.NewRecorder()
		s.ServeHTTP(recorder, request)
		if recorder.Code != tt.code {
			t.Errorf("ServerHTTP with Host %q: Want %d. Got %d.", tt.host, tt.code, recorder.Code)
		}
		if recorder.Body.String() != tt.body {
			t.Errorf("ServerHTTP with Host %q wrong body: Want %q. Got %q.", tt.host, tt.body, recorder.Body.String())
		}
	}
}

func TestLoadRules(t *testing.T) {
	rules, err := loadRules("testdata/rules.json")
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
	for _, tt := range tests {
		rules, err := loadRules(tt.filename)
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

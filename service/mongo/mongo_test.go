package mongo

import (
	"strings"
	"testing"

	"miaoverse/consts"
)

func TestBuildClientOptionsRequiresHost(t *testing.T) {
	_, err := buildClientOptions(Config{Database: "miaoverse"})
	if !strings.Contains(err.Error(), "host is required") {
		t.Fatalf("buildClientOptions() error = %v, want host required error", err)
	}
}

func TestBuildClientOptionsRequiresDatabase(t *testing.T) {
	_, err := buildClientOptions(Config{Host: "127.0.0.1"})
	if !strings.Contains(err.Error(), "db_name is required") {
		t.Fatalf("buildClientOptions() error = %v, want db_name required error", err)
	}
}

func TestBuildClientOptionsRejectsBadPort(t *testing.T) {
	_, err := buildClientOptions(Config{Host: "127.0.0.1", Port: 70000, Database: "miaoverse"})
	if !strings.Contains(err.Error(), "port must be in range") {
		t.Fatalf("buildClientOptions() error = %v, want port range error", err)
	}
}

func TestBuildClientOptionsDefaults(t *testing.T) {
	opts, err := buildClientOptions(Config{Host: "127.0.0.1", Database: "miaoverse"})
	if err != nil {
		t.Fatal(err)
	}
	if len(opts.Hosts) != 1 || opts.Hosts[0] != "127.0.0.1:27017" {
		t.Fatalf("hosts = %v, want [127.0.0.1:27017]", opts.Hosts)
	}
	if timeout := opts.ConnectTimeout; timeout == nil || *timeout != consts.MongoDefaultConnectTimeout {
		t.Fatalf("connect timeout = %v, want %v", timeout, consts.MongoDefaultConnectTimeout)
	}
	if selTimeout := opts.ServerSelectionTimeout; selTimeout == nil || *selTimeout != consts.MongoDefaultServerSelectionTimeout {
		t.Fatalf("server selection timeout = %v, want %v", selTimeout, consts.MongoDefaultServerSelectionTimeout)
	}
}

func TestBuildClientOptionsAuth(t *testing.T) {
	opts, err := buildClientOptions(Config{
		Host:       "127.0.0.1",
		Database:   "miaoverse",
		Username:   "user",
		Password:   "pass",
		AuthSource: "admin",
	})
	if err != nil {
		t.Fatal(err)
	}
	cred := opts.Auth
	if cred == nil {
		t.Fatal("auth should be set when username provided")
	}
	if cred.Username != "user" || cred.Password != "pass" || cred.AuthSource != "admin" {
		t.Fatalf("auth = %+v, want user/pass/admin", cred)
	}
}

func TestBuildClientOptionsNoAuthWhenUsernameEmpty(t *testing.T) {
	opts, err := buildClientOptions(Config{Host: "127.0.0.1", Database: "miaoverse"})
	if err != nil {
		t.Fatal(err)
	}
	if cred := opts.Auth; cred != nil {
		t.Fatalf("auth should be nil when username empty, got %+v", cred)
	}
}

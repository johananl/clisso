package aws

import (
	"bytes"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-ini/ini"
)

func TestWriteToFile(t *testing.T) {
	id := "testkey"
	sec := "testsecret"
	tok := "testtoken"
	exp := time.Now()

	c := Credentials{
		AccessKeyID:     id,
		SecretAccessKey: sec,
		SessionToken:    tok,
		Expiration:      exp,
	}

	fn := "test_creds.txt"
	p := "testprofile"

	// Write credentials
	err := WriteToFile(&c, fn, p)
	if err != nil {
		t.Fatal("Could not write credentials to file: ", err)
	}

	// Read credentials
	cfg, err := ini.Load(fn)
	if err != nil {
		t.Fatal("Could not load INI file: ", err)
	}

	s := cfg.Section(p)

	// Verify credentials
	k, err := s.GetKey("aws_access_key_id")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != id {
		t.Fatalf("Wrong access key ID: got %s, want %s", k.String(), id)
	}

	k, err = s.GetKey("aws_secret_access_key")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != sec {
		t.Fatalf("Wrong secret access key: got %s, want %s", k.String(), sec)
	}

	k, err = s.GetKey("aws_session_token")
	if err != nil {
		t.Fatal(err)
	}
	if k.String() != tok {
		t.Fatalf("Wrong session token: got %s, want %s", k.String(), tok)
	}

	k, err = s.GetKey("aws_expiration")
	if err != nil {
		t.Fatal(err)
	}
	f := exp.UTC().Format(time.RFC3339)
	if k.String() != f {
		t.Fatalf("Wrong expiration: got %s, want %s", k.String(), f)
	}

	err = os.Remove(fn)
	if err != nil {
		t.Fatalf("Could not remove file %v during cleanup", fn)
	}
}

func TestWriteToShellUnix(t *testing.T) {
	id := "testkey"
	sec := "testsecret"
	tok := "testtoken"
	exp := time.Now()

	c := Credentials{
		AccessKeyID:     id,
		SecretAccessKey: sec,
		SessionToken:    tok,
		Expiration:      exp,
	}
	var b bytes.Buffer

	WriteToShell(&c, false, &b)

	got := b.String()
	want := fmt.Sprintf(
		"export AWS_ACCESS_KEY_ID=%v\nexport AWS_SECRET_ACCESS_KEY=%v\nexport AWS_SESSION_TOKEN=%v\n",
		id,
		sec,
		tok,
	)

	if got != want {
		t.Fatalf("Wrong info written to shell: got %v want %v", got, want)
	}
}

func TestWriteToShellWindows(t *testing.T) {
	id := "testkey"
	sec := "testsecret"
	tok := "testtoken"
	exp := time.Now()

	c := Credentials{
		AccessKeyID:     id,
		SecretAccessKey: sec,
		SessionToken:    tok,
		Expiration:      exp,
	}
	var b bytes.Buffer

	WriteToShell(&c, true, &b)

	got := b.String()
	want := fmt.Sprintf(
		"set AWS_ACCESS_KEY_ID=%v\nset AWS_SECRET_ACCESS_KEY=%v\nset AWS_SESSION_TOKEN=%v\n",
		id,
		sec,
		tok,
	)

	if got != want {
		t.Fatalf("Wrong info written to shell: got %v want %v", got, want)
	}
}

package main

import (
	"bytes"
	"testing"
)

func TestRewriteToHeader(t *testing.T) {
	name := "foo"
	addr := name + "@example.com"
	testEmail := "X-A-Header: Test\nTo: " + addr + "\nSubject: Test\n\nHello World!"
	buf := bytes.NewBufferString(testEmail)
	msg, err := rewriteToHeader(buf)
	if err != nil {
		t.Errorf("e rewriting headers: %v", err)
		t.FailNow()
	}

	to, err := msg.Header.Text("To")
	if err != nil {
		t.Errorf("could not grab To header: %v\n", err)
		t.FailNow()
	}

	if to != name {
		t.Errorf("for To header expected %s got %s", name, to)
	}
}

func TestRewriteToHeaderNoTo(t *testing.T) {
	testEmail := "X-A-Header: Test\nSubject: Test\n\nHello World!"
	buf := bytes.NewBufferString(testEmail)
	msg, err := rewriteToHeader(buf)
	if err != nil {
		t.Errorf("e rewriting headers: %v", err)
		t.FailNow()
	}

	to, err := msg.Header.Text("To")
	if err != nil {
		t.Errorf("could not grab To header: %v\n", err)
		t.FailNow()
	}

	if to != "" {
		t.Error("expected empty To header")
	}
}

func TestRewriteToHeaderEmptyTo(t *testing.T) {
	testEmail := "X-A-Header: Test\nTo: \nSubject: Test\n\nHello World!"
	buf := bytes.NewBufferString(testEmail)
	msg, err := rewriteToHeader(buf)
	if err != nil {
		t.Errorf("e rewriting headers: %v", err)
		t.FailNow()
	}

	to, err := msg.Header.Text("To")
	if err != nil {
		t.Errorf("could not grab To header: %v\n", err)
		t.FailNow()
	}

	if to != "" {
		t.Error("expected empty To header")
	}
}

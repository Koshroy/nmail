package main

import (
	"bytes"
	"testing"

	"github.com/emersion/go-message/mail"
)

func TestRewriteFromHeader(t *testing.T) {
	testEmail := "X-A-Header: Test\nFrom: foo@example.com\nSubject: Test\n\nHello World!"
	buf := bytes.NewBufferString(testEmail)
	nodeID := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	newFrom := "foo@" + nodeID + ".id.nncp"
	msg, err := rewriteFromHeader(buf, nodeID)
	if err != nil {
		t.Errorf("e rewriting headers: %v", err)
		t.FailNow()
	}

	from, err := msg.Header.Text("From")
	if err != nil {
		t.Errorf("could not grab From header: %v\n", err)
		t.FailNow()
	}

	addr, err := mail.ParseAddress(from)
	if err != nil {
		t.Errorf("Could not parse new from header: %s because %v", from, err)
		t.FailNow()
	}

	if addr.Address != newFrom {
		t.Errorf("Expected From to be %s but got %s", newFrom, addr.Address)
	}
}

func TestRewriteFromHeaderNoFrom(t *testing.T) {
	testEmail := "X-A-Header: Test\nSubject: Test\n\nHello World!"
	buf := bytes.NewBufferString(testEmail)
	nodeID := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	msg, err := rewriteFromHeader(buf, nodeID)
	if err != nil {
		t.Errorf("e rewriting headers: %v", err)
		t.FailNow()
	}

	from, err := msg.Header.Text("From")
	if err != nil {
		t.Errorf("could not grab From header: %v\n", err)
		t.FailNow()
	}

	if from != "" {
		t.Errorf("expected no From header got %s", from)
	}
}

func TestRewriteFromHeaderEmptyFrom(t *testing.T) {
	testEmail := "X-A-Header: Test\nFrom: \nSubject: Test\n\nHello World!"
	buf := bytes.NewBufferString(testEmail)
	nodeID := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	msg, err := rewriteFromHeader(buf, nodeID)
	if err != nil {
		t.Errorf("e rewriting headers: %v", err)
		t.FailNow()
	}

	from, err := msg.Header.Text("From")
	if err != nil {
		t.Errorf("could not grab From header: %v\n", err)
		t.FailNow()
	}

	if from != "" {
		t.Errorf("expected no From header got %s", from)
	}
}

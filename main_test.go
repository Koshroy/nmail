package main

import (
	"testing"
	// "github.com/emersion/go-message/mail"
)

func TestNNCPMailAddress(t *testing.T) {
	addr := NNCPMailAddress{"foo", "nncpexample"}
	expected := "foo@nncpexample.nncp"
	got := addr.String()
	if got != expected {
		t.Errorf("expected: %s got: %s", expected, got)
	}
}

func TestParseRecipientAlias(t *testing.T) {
	emailRecipient := "foo@alice.nncp"
	expectedLocalPart := "foo"
	expectedNodeName := "alice"

	nncpAddr, err := parseRecipient(emailRecipient)
	if err != nil {
		t.Errorf("error parsing: %v", err)
	}

	if nncpAddr.LocalPart != expectedLocalPart {
		t.Errorf("expected localpart of %s but got %s", expectedLocalPart, nncpAddr.LocalPart)
	}

	if nncpAddr.NodeName != expectedNodeName {
		t.Errorf("expected nodename of %s but got %s", expectedNodeName, nncpAddr.NodeName)
	}
}

func TestParseRecipientBadDomain(t *testing.T) {
	emailRecipient := "foo@alice.example.com"

	_, err := parseRecipient(emailRecipient)
	if err == nil {
		t.Error("expected error parsing but got no error")
	}
}

func TestParseRecipientBadNNCPDomain(t *testing.T) {
	emailRecipient := "foo@example.foo.nncp"

	_, err := parseRecipient(emailRecipient)
	if err == nil {
		t.Error("expected error parsing but got no error")
	}
}

func TestParseRecipientNodeId(t *testing.T) {
	emailRecipient := "foo@AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.id.nncp"
	expectedLocalPart := "foo"
	expectedNodeName := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

	nncpAddr, err := parseRecipient(emailRecipient)
	if err != nil {
		t.Errorf("error parsing: %v", err)
	}

	if nncpAddr.LocalPart != expectedLocalPart {
		t.Errorf("expected localpart of %s but got %s", expectedLocalPart, nncpAddr.LocalPart)
	}

	if nncpAddr.NodeName != expectedNodeName {
		t.Errorf("expected nodename of %s but got %s", expectedNodeName, nncpAddr.NodeName)
	}
}

func TestParseRecipientBadNNCPIdDomain(t *testing.T) {
	emailRecipient := "foo@AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.id.foo.nncp"

	_, err := parseRecipient(emailRecipient)
	if err == nil {
		t.Error("expected error parsing but got no error")
	}
}

func TestParseRecipientInvalidNNCPId(t *testing.T) {
	emailRecipient := "foo@nodeid.id.nncp"

	_, err := parseRecipient(emailRecipient)
	if err == nil {
		t.Error("expected error parsing but got no error")
	}
}

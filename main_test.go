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

func TestParseRecipient(t *testing.T) {
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

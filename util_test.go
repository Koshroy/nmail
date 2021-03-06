package main

import (
	"testing"
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

func TestParseRecipientAliasWithName(t *testing.T) {
	emailRecipient := "Alice Example <foo@alice.nncp>"
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
	nodeID := "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	emailRecipient := "foo@" + nodeID + ".id.nncp"
	expectedLocalPart := "foo"
	expectedNodeName := nodeID

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
	_, err := parseRecipient("foo@AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.id.foo.nncp")
	if err == nil {
		t.Error("expected error parsing but got no error")
	}

	_, err = parseRecipient("foo@AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA.foo.id.nncp")
	if err == nil {
		t.Error("expected error parsing but got no error")
	}
}

func TestSplitEmailAddress(t *testing.T) {
	addr, err := splitEmailAddress("foo@example.com")
	localPart := "foo"
	domain := "example.com"

	if err != nil {
		t.Errorf("could not split email address: %v", err)
	}
	if addr.LocalPart != localPart {
		t.Errorf("expected local part %s got %s", localPart, addr.LocalPart)
	}
	if addr.Domain != domain {
		t.Errorf("expected local part %s got %s", domain, addr.Domain)
	}
}

func TestSplitEmailAddressTLD(t *testing.T) {
	addr, err := splitEmailAddress("foo@example")
	localPart := "foo"
	domain := "example"

	if err != nil {
		t.Errorf("could not split email address: %v", err)
	}
	if addr.LocalPart != localPart {
		t.Errorf("expected local part %s got %s", localPart, addr.LocalPart)
	}
	if addr.Domain != domain {
		t.Errorf("expected local part %s got %s", domain, addr.Domain)
	}
}

func TestSplitEmailAddressErrInvalidAddr(t *testing.T) {
	_, err := splitEmailAddress("foo@")
	if err == nil {
		t.Error("expected error parsing")
	}
}

func TestSplitEmailAddressErrNoDomain(t *testing.T) {
	_, err := splitEmailAddress("foo")
	if err == nil {
		t.Error("expected error parsing")
	}
}

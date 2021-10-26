package main

import (
	"errors"
	"fmt"
	"strings"

	message "github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
)

type EmailAddress struct {
	LocalPart string
	Domain    string
}

func (a EmailAddress) String() string {
	return a.LocalPart + "@" + a.Domain
}

type NNCPMailAddress struct {
	LocalPart string
	NodeName  string
}

func (a NNCPMailAddress) String() string {
	return a.LocalPart + "@" + a.NodeName + ".nncp"
}

func splitEmailAddress(addr string) (EmailAddress, error) {
	zero := EmailAddress{"", ""}
	splits := strings.SplitN(addr, "@", 2)
	if len(splits) != 2 {
		return zero, errors.New("could not split email into localpart and domain")
	}
	if splits[1] == "" {
		return zero, errors.New("could not split email into localpart and domain")
	}

	return EmailAddress{splits[0], splits[1]}, nil
}

func parseRecipient(addr string) (NNCPMailAddress, error) {
	zero := NNCPMailAddress{"", ""}
	address, err := mail.ParseAddress(addr)
	if err != nil {
		return zero, err
	}

	emailAddr, err := splitEmailAddress(address.Address)
	if err != nil {
		return zero, err
	}

	if !strings.HasSuffix(emailAddr.Domain, ".nncp") {
		return zero, errors.New("email domain must use .nncp TLD (must end in .nncp)")
	}

	isNodeId := strings.HasSuffix(emailAddr.Domain, ".id.nncp")

	numDots := strings.Count(emailAddr.Domain, ".")
	if isNodeId {
		if numDots != 2 {
			return zero, errors.New("email domain for node ID must be of the form <ID>.id.nncp")
		}
	} else {
		if numDots != 1 {
			return zero, errors.New("email domain for node alias must be of form <alias>.nncp")
		}
	}
	var nodeName string
	if isNodeId {
		nodeName = strings.ToUpper(strings.TrimSuffix(emailAddr.Domain, ".id.nncp"))
	} else {
		nodeName = strings.TrimSuffix(emailAddr.Domain, ".nncp")
	}

	return NNCPMailAddress{emailAddr.LocalPart, nodeName}, nil
}

func rewriteHeader(entity *message.Entity, header string, rewriteFunc func(string) (string, error)) error {
	oldHeader, err := entity.Header.Text(header)
	if err != nil && !message.IsUnknownCharset(err) {
		return fmt.Errorf("could not parse From header: %w", err)
	}

	newHeader, err := rewriteFunc(oldHeader)
	if err != nil {
		return fmt.Errorf("error rewriting header: %w", err)
	}

	entity.Header.SetText(header, newHeader)
	return nil
}

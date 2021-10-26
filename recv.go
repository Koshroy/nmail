package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/emersion/go-message/mail"
	message "github.com/emersion/go-message"
)

func recvMail(sender string, debug bool) {
	if debug {
		log.Printf("mail receive")
	}

	if sender == "" {
		log.Fatalln("No sender provided in NNCP_SENDER, aborting")
	}

	msg, err := rewriteFromHeader(os.Stdin, sender)
	if err != nil {
		log.Fatalf("error rewriting message: %v\n", err)
		return
	}

	err = msg.WriteTo(os.Stdout)
	if err != nil {
		log.Fatalf("error writing mail to stdout: %v\n", err)
	}
}

func rewriteFromHeader(r io.Reader, srcNode string) (*message.Entity, error) {
	if srcNode == "" {
		return nil, errors.New("a valid new from address is required")
	}

	msg, err := message.Read(r)
	if err != nil && !message.IsUnknownCharset(err) {
		return nil, fmt.Errorf("could not read mail message: %w", err)
	}

	rewriteFunc := func(old string) (string, error) {
		return mungeFrom(old, srcNode)
	}

	err = rewriteHeader(msg, "From", rewriteFunc)
	if err != nil {
		return nil, fmt.Errorf("could not rewrite From header: %w", err)
	}

	return msg, nil
}

func mungeFrom(old, srcNode string) (string, error) {
	if old == "" {
		return "", nil
	}

	oldAddr, err := mail.ParseAddress(old)
	if err != nil {
		return "", fmt.Errorf("could not parse From header as address: %w", err)
	}

	split, err := splitEmailAddress(oldAddr.Address)
	if err != nil {
		return "", fmt.Errorf("could not split From header value: %w", err)
	}

	newAddr := EmailAddress{
		LocalPart: split.LocalPart,
		Domain:    srcNode + ".id.nncp",
	}

	newMailAddr := mail.Address{
		Name:    oldAddr.Name,
		Address: newAddr.String(),
	}

	return newMailAddr.String(), nil
}

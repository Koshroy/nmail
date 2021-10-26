package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	message "github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
)

func sendMail(rcpt, nncpCfgPath, handle string, debug bool) {
	if debug {
		log.Println("send mail")
	}

	if rcpt == "" {
		log.Fatalln("No recipient provided")
	}

	address, err := parseRecipient(rcpt)
	if err != nil {
		log.Fatalf("error parsing recipient address %s: %v\n", rcpt, err)
	}

	entity, err := rewriteToHeader(os.Stdin)
	if err != nil {
		log.Fatalf("could not rewrite To header: %v\n", err)
	}

	var mail bytes.Buffer
	err = entity.WriteTo(&mail)
	if err != nil {
		log.Fatalf("error writing mail to temp buffer: %v\n", err)
	}

	err = nncpSendmail(nncpCfgPath, address, handle, &mail, debug)
	if err != nil {
		log.Fatalf("error sending mail via nncp: %v\n", err)
	}

	return
}

func nncpSendmail(nncpCfgPath string, recipient NNCPMailAddress, handle string, reader io.Reader, debug bool) error {
	var cmd *exec.Cmd
	var out *bytes.Buffer

	if debug {
		path := nncpCfgPath
		if path == "" {
			path = "<empty>"
		}
		log.Printf(
			"Sending mail through nncp-exec at config-path: %s to %s:%s\n",
			path,
			recipient.LocalPart,
			recipient.NodeName,
		)

		out = new(bytes.Buffer)
	}

	if nncpCfgPath == "" {
		cmd = exec.Command("nncp-exec", recipient.NodeName, handle, recipient.LocalPart)
	} else {
		cmd = exec.Command("nncp-exec", "-cfg", nncpCfgPath, recipient.NodeName, handle, recipient.LocalPart)
	}
	if debug {
		cmd.Stderr = out
	}
	cmd.Stdin = reader

	err := cmd.Run()
	if debug {
		log.Println("nncp stderr:", strings.TrimSpace(out.String()))
	}
	return err
}

func rewriteToHeader(r io.Reader) (*message.Entity, error) {
	msg, err := message.Read(r)
	if err != nil && !message.IsUnknownCharset(err) {
		return nil, fmt.Errorf("could not read mail message: %w", err)
	}

	err = rewriteHeader(msg, "To", mungeTo)
	if err != nil {
		return nil, fmt.Errorf("could not rewrite To header: %w", err)
	}

	return msg, nil
}

func mungeTo(old string) (string, error) {
	if old == "" {
		return "", nil
	}

	oldAddr, err := mail.ParseAddress(old)
	if err != nil {
		return "", fmt.Errorf("could not parse To header as address: %w", err)
	}

	split, err := splitEmailAddress(oldAddr.Address)
	if err != nil {
		return "", fmt.Errorf("could not split From header value: %w", err)
	}

	return split.LocalPart, nil
}

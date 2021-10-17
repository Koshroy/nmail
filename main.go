package main

import (
	"bytes"
	"errors"
	"flag"
	"io"
	"log"
	"net/mail"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type EmailAddress struct {
	LocalPart string
	Domain    string
}

type NNCPMailAddress struct {
	LocalPart string
	NodeName  string
}

func main() {
	rcpt := flag.String("rcpt", "", "mail recipient")
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()

	nncpCfgPath := getCfgPath(*debug)
	if *rcpt == "" {
		log.Fatalln("No recipient provided")
	}

	address, err := parseRecipient(*rcpt)
	if err != nil {
		log.Fatalf("Error parsing recipient address %s: %v\n", *rcpt, err)
	}

	err = nncpSendmail(nncpCfgPath, address, os.Stdin, *debug)
	if err != nil {
		log.Fatalf("error sending mail via nncp: %v\n", err)
	}

	return
}

func getCfgPath(debug bool) string {
	nncpCfgPath := os.Getenv("NNCP_CFG_PATH")
	if nncpCfgPath != "" {
		absPath, err := filepath.Abs(nncpCfgPath)
		if err == nil {
			nncpCfgPath = absPath
		} else {
			if debug {
				log.Printf("error canonicalizing config path: %v\n", err)
			}
		}
	}

	return nncpCfgPath
}

func parseEmailAddress(addr string) (EmailAddress, error) {
	zero := EmailAddress{"", ""}
	splits := strings.SplitN(addr, "@", 2)
	if len(splits) != 2 {
		return zero, errors.New("Could not split email into localpart and domain")
	}

	return EmailAddress{splits[0], splits[1]}, nil
}

func parseRecipient(addr string) (NNCPMailAddress, error) {
	zero := NNCPMailAddress{"", ""}
	lower := strings.ToLower(addr)
	address, err := mail.ParseAddress(lower)
	if err != nil {
		return zero, err
	}

	emailAddr, err := parseEmailAddress(address.Address)
	if err != nil {
		return zero, err
	}

	if !strings.HasSuffix(emailAddr.Domain, ".nncp") {
		return zero, errors.New("Email domain must use .nncp TLD (must end in .nncp)")
	}

	nodeName := strings.TrimSuffix(emailAddr.Domain, ".nncp")

	return NNCPMailAddress{emailAddr.LocalPart, nodeName}, nil
}

func nncpSendmail(nncpCfgPath string, recipient NNCPMailAddress, reader io.Reader, debug bool) error {
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
		cmd = exec.Command("nncp-exec", recipient.NodeName, "sendmail", recipient.LocalPart)
	} else {
		cmd = exec.Command("nncp-exec", "-cfg", nncpCfgPath, recipient.NodeName, "sendmail", recipient.LocalPart)
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

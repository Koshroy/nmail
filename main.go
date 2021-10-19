package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	message "github.com/emersion/go-message"
	"github.com/emersion/go-message/mail"
)

type EmailAddress struct {
	LocalPart string
	Domain    string
}

type NNCPMailAddress struct {
	LocalPart string
	NodeName  string
}

func (a NNCPMailAddress) String() string {
	return a.LocalPart + "@" + a.NodeName + ".nncp"
}

func main() {
	log.SetOutput(os.Stderr)
	rcpt := flag.String("rcpt", "", "mail recipient")
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()
	sendRecv := strings.ToLower(flag.Arg(0))
	nncpCfgPath := getCfgPath(*debug)
	sender := os.Getenv("NNCP_SENDER")

	if sendRecv == "" {
		sendMail(*rcpt, nncpCfgPath, *debug)
	} else if sendRecv == "send" {
		sendMail(*rcpt, nncpCfgPath, *debug)
	} else if sendRecv == "receive" || sendRecv == "recv" {
		recvMail(sender, *debug)
	} else {
		log.Fatalln(sendRecv, "is not a valid mode")
	}
}

func recvMail(sender string, debug bool) {
	if debug {
		log.Printf("mail receive")
	}

	if sender == "" {
		log.Fatalln("No sender provided in NNCP_SENDER, aborting")
	}

	msg, err := mungeFrom(os.Stdin, sender, debug)
	if err != nil {
		log.Fatalf("Error rewriting message: %v\n", err)
		return
	}

	err = msg.WriteTo(os.Stdout)
	if err != nil {
		log.Fatalf("error writing mail to stdout: %v\n", err)
	}
}

func sendMail(rcpt, nncpCfgPath string, debug bool) {
	if debug {
		log.Println("send mail")
	}

	if rcpt == "" {
		log.Fatalln("No recipient provided")
	}

	address, err := parseRecipient(rcpt)
	if err != nil {
		log.Fatalf("Error parsing recipient address %s: %v\n", rcpt, err)
	}

	err = nncpSendmail(nncpCfgPath, address, os.Stdin, debug)
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

func setDomain(addr *mail.Address, domain string) mail.Address {
	mailAddr := addr.Address
	if mailAddr == "" {
		return mail.Address{"", ""}
	}

	splits := strings.SplitN(mailAddr, "@", 2)
	if len(splits) >= 1 {
		return mail.Address{addr.Name, splits[0] + "@" + domain}
	} else {
		return mail.Address{"", ""}
	}
}

func mungeFrom(r io.Reader, srcNode string, debug bool) (*message.Entity, error) {
	if srcNode == "" {
		return nil, errors.New("a valid new from address is required")
	}

	msg, err := message.Read(r)
	if err != nil && !message.IsUnknownCharset(err) {
		return nil, fmt.Errorf("could not read mail message: %w", err)
	}

	oldFromRaw, err := msg.Header.Text("From")
	if err != nil && !message.IsUnknownCharset(err) {
		return nil, fmt.Errorf("could not parse From header: %w", err)
	}

	// If there is no From address in the originating email, then we do
	// not need to perform header munging, so we should just return the email
	if oldFromRaw == "" {
		if debug {
			log.Println("No From header in source email so no need to munge headers")
		}
		return msg, nil
	}

	oldFrom, err := mail.ParseAddress(oldFromRaw)
	if err != nil && !message.IsUnknownCharset(err) {
		return nil, fmt.Errorf("could not parse From address: %w", err)
	}

	newFrom := setDomain(oldFrom, srcNode+".nncp")

	if debug {
		log.Println("old From header:", oldFrom.String())
		log.Println("new From header:", newFrom.String())
	}

	msg.Header.SetText("From", newFrom.String())
	return msg, nil
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

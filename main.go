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

func main() {
	log.SetOutput(os.Stderr)
	debug := flag.Bool("debug", false, "debug mode")
	flag.Parse()
	sendRecv := strings.ToLower(flag.Arg(0))
	rcpt := flag.Arg(1)
	nncpCfgPath := getCfgPath(*debug)
	sender := os.Getenv("NNCP_SENDER")

	if sendRecv == "" {
		sendMail(rcpt, nncpCfgPath, *debug)
	} else if sendRecv == "send" {
		sendMail(rcpt, nncpCfgPath, *debug)
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

	msg, err := rewriteHeaders(os.Stdin, sender, debug)
	if err != nil {
		log.Fatalf("error rewriting message: %v\n", err)
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
		log.Fatalf("error parsing recipient address %s: %v\n", rcpt, err)
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

func rewriteHeaders(r io.Reader, srcNode string, debug bool) (*message.Entity, error) {
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

	oldFromAddr, err := splitEmailAddress(oldFrom.Address)
	if err != nil {
		return nil, fmt.Errorf("Error parsing From address: %w", err)
	}

	// On receipt, the sender is always in node ID form. So we always want to
	// create the new From header in ID form
	newFromAddr := EmailAddress{
		LocalPart: oldFromAddr.LocalPart,
		Domain: srcNode + ".id.nncp",
	}
	newFrom := mail.Address{
		Name: oldFrom.Name,
		Address: newFromAddr.String(),
	}

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

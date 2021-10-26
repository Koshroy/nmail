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
	sendMailHandle := flag.String("handle", "sendmail", "handle used to send mail")
	flag.Parse()
	sendRecv := strings.ToLower(flag.Arg(0))
	rcpt := flag.Arg(1)
	nncpCfgPath := getCfgPath(*debug)
	sender := os.Getenv("NNCP_SENDER")

	if sendRecv == "" {
		sendMail(rcpt, nncpCfgPath, *sendMailHandle, *debug)
	} else if sendRecv == "send" {
		sendMail(rcpt, nncpCfgPath, *sendMailHandle, *debug)
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

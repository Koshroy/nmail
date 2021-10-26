package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

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

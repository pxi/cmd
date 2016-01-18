// Mfa is a simple tool for managing mfa passwords in OSX keychain.
package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/hgfischer/go-otp"
	"github.com/keybase/go-keychain"
)

func usage() {
	fmt.Fprintf(os.Stderr, "usage: mfa [SERVICE [SECRET]]\n")
	os.Exit(1)
}

func main() {
	flag.Parse()

	log.SetFlags(0)
	log.SetPrefix("mfa: ")

	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService("mfa")

	var err error
	var res []keychain.QueryResult
	switch flag.NArg() {
	case 0:
		// List stored accounts.
		item.SetMatchLimit(keychain.MatchLimitAll)
		item.SetReturnAttributes(true)
		res, err = keychain.QueryItem(item)
		if err != nil {
			break
		}
		for _, qres := range res {
			fmt.Println(qres.Account)
		}
	case 1:
		// Get mfa password for an account.
		item.SetMatchLimit(keychain.MatchLimitOne)
		item.SetReturnData(true)
		item.SetAccount(flag.Arg(0))
		res, err = keychain.QueryItem(item)
		if err == nil && len(res) == 0 {
			err = errors.New("account not found")
		}
		if err != nil {
			break
		}
		for _, qres := range res {
			totp := &otp.TOTP{
				Secret:         string(qres.Data),
				Length:         otp.DefaultLength,
				Period:         otp.DefaultPeriod,
				IsBase32Secret: true,
			}
			fmt.Println(totp.Get())
		}
	case 2:
		// Create a new account.
		item.SetAccount(flag.Arg(0))
		sk := strings.ToUpper(flag.Arg(1))
		if !check(sk) {
			log.Printf("warning: secret is not compatible with google authenticator\n")
		}
		item.SetData([]byte(sk))
		err = keychain.AddItem(item)
	default:
		usage()
	}
	if err != nil {
		log.Fatal(err)
	}
}

func check(s string) bool {
	s = strings.Replace(s, "=", "", -1)
	s = strings.Replace(s, " ", "", -1)
	return len(s) == 16
}

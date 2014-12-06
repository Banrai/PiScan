// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/Banrai/PiScan/server/database/barcodes"
	"github.com/Banrai/PiScan/server/digest"
	"github.com/Banrai/PiScan/server/emailer"
	"net/http"
	"net/url"
	"text/template"
)

const (
	VERIFY_SENDER  = "openproductdata@saruzai.com"
	VERIFY_SUBJECT = "Please verify your email address"
	VERIFY_MESSAGE = `<p>Thank you for registering to contribute to the Open Product Database.</p>
<p>Please confirm your email address by clicking on this link:</p>
<p><a href="{{.APIServer}}/verify/{{.APICode}}">http://{{.APIServer}}/verify/{{.APICode}}</a></p>
<p>You only have to do this once per email address.</p>`
)

type RegistrationLink struct {
	APIServer string
	APICode   string
}

func SendVerificationEmail(server, email, code string) error {
	var msg bytes.Buffer
	t := template.Must(template.New("VERIFY_MESSAGE").Parse(VERIFY_MESSAGE))
	context := RegistrationLink{server, code}
	err := t.Execute(&msg, context)
	if err == nil {
		sender := emailer.EmailAddress{Address: VERIFY_SENDER}
		recipient := emailer.EmailAddress{Address: email}
		err = emailer.Send(VERIFY_SUBJECT, msg.String(), "text/html", &sender, &recipient, []*emailer.EmailAttachment{})
	}
	return err
}

func RegisterAccount(r *http.Request, db DBConnection, server string) string {
	// the result is a simple json ack
	ack := new(SimpleMessage)

	// this function only responds to GET requests with
	// an hmac digest of the contents
	if "GET" == r.Method {
		params, paramErr := url.ParseQuery(r.URL.RawQuery)
		if paramErr != nil {
			ack.Err = paramErr
		} else {
			email := params.Get("email")
			apiCode := params.Get("api")
			hmacDigest := params.Get("hmac")

			if email != "" && apiCode != "" {
				// confirm the digest matches
				params.Del("hmac") // separate the digest from the rest
				if digest.DigestMatches(email, params.Encode(), hmacDigest) {
					// the request is valid
					registerFn := func(statements map[string]*sql.Stmt) {
						lookupStmt, lookupStmtExists := statements[barcodes.ACCOUNT_LOOKUP_BY_EMAIL]
						insertStmt, insertStmtExists := statements[barcodes.ACCOUNT_INSERT]
						if lookupStmtExists && insertStmtExists {
							// see if the email is available
							acc, accErr := barcodes.LookupAccount(lookupStmt, email, false)
							if accErr != nil {
								ack.Err = accErr
							} else {
								if acc.Id == "" {
									// can proceed with the registration (add this email + api combination)
									acc.Email = email
									acc.APICode = apiCode
									pk, addErr := acc.Add(insertStmt)
									if addErr != nil {
										ack.Err = accErr
										// the account is created, but unverified

										// send an email for verfication
										verifyErr := SendVerificationEmail(server, email, apiCode)

										// and update this json reply
										ack.Ack = fmt.Sprintf("ok: %s", pk)
										ack.Err = verifyErr
									}
								}
							}
						}
					}
					WithServerDatabase(db, registerFn)
				}
			}
		}
	}

	result, err := json.Marshal(ack)
	if err != nil {
		fmt.Println(err)
	}
	return string(result)
}

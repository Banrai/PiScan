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
	SERVER_SENDER = "openproductdata@saruzai.com" // used by other email notifications

	VERIFY_SUBJECT = "Please verify your email address"
	VERIFY_MESSAGE = `<p>Thank you for registering to contribute to the Open Product Database.</p>
<p>Please confirm your email address by clicking on this link:</p>
<p><a href="{{.APIServer}}/verify/{{.APICode}}">{{.APIServer}}/verify/{{.APICode}}</a></p>
<p>You only have to do this once per email address.</p>`
)

type RegistrationLink struct {
	APIServer string
	APICode   string
}

func SendVerificationEmail(server, email, code string, port int) error {
	var msg bytes.Buffer
	t := template.Must(template.New("VERIFY_MESSAGE").Parse(VERIFY_MESSAGE))
	context := RegistrationLink{fmt.Sprintf("https://api.%s:%d", server, port), code}
	err := t.Execute(&msg, context)
	if err == nil {
		sender := emailer.EmailAddress{Address: SERVER_SENDER}
		recipient := emailer.EmailAddress{Address: email}
		err = emailer.Send(VERIFY_SUBJECT, msg.String(), "text/html", &sender, &recipient, []*emailer.EmailAttachment{})
	}
	return err
}

func RegisterAccount(r *http.Request, db DBConnection, server string, port int) string {
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
								if acc.Id != "" {
									// this account has already been registered
									ack.Ack = fmt.Sprintf("exists: %s", acc.Id)

									if !acc.Verified {
										// but it has yet to be verified, so send an email
										ack.Err = SendVerificationEmail(server, email, acc.Id, port)
									}
								} else {
									// can proceed with the registration (add this email + api combination)
									acc.Email = email
									acc.APICode = apiCode
									pk, addErr := acc.Add(insertStmt)
									if addErr != nil {
										ack.Err = accErr
									} else {
										// the account is created, but unverified

										// send an email for verfication
										ack.Err = SendVerificationEmail(server, email, pk, port)

										// and update this json reply
										ack.Ack = fmt.Sprintf("ok: %s", pk)
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

func VerifyAccount(r *http.Request, db DBConnection) string {
	// accumulate the result in a simple ack struct
	ack := new(SimpleMessage)

	// this function only responds to GET requests
	if "GET" == r.Method {
		// get the verification code from the query string
		code := r.URL.Path[len("/verify/"):]

		verifyFn := func(statements map[string]*sql.Stmt) {
			lookupStmt, lookupStmtExists := statements[barcodes.ACCOUNT_LOOKUP_BY_ID]
			updateStmt, updateStmtExists := statements[barcodes.ACCOUNT_UPDATE]
			if lookupStmtExists && updateStmtExists {
				// see if the account for this code exists
				acc, accErr := barcodes.LookupAccount(lookupStmt, code, true)
				if accErr != nil {
					ack.Err = accErr
				} else {
					if acc.Id == code {
						// can proceed with the verification
						acc.Verified = true
						ack.Err = acc.Update(updateStmt)
					}
				}
			}
		}
		WithServerDatabase(db, verifyFn)
	}

	// need to return a simple html string in reply
	if ack.Err != nil {
		return ack.Err.Error()
	} else {
		return "Thank you for verifying your email address"
	}
}

func GetAccountStatus(r *http.Request, db DBConnection) string {
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
			hmacDigest := params.Get("hmac")

			if email != "" && hmacDigest != "" {
				// confirm the digest matches
				params.Del("hmac") // separate the digest from the rest
				if digest.DigestMatches(email, params.Encode(), hmacDigest) {
					// the request is valid
					statusFn := func(statements map[string]*sql.Stmt) {
						lookupStmt, lookupStmtExists := statements[barcodes.ACCOUNT_LOOKUP_BY_EMAIL]
						if lookupStmtExists {
							// see if the email corresponds to an account
							acc, accErr := barcodes.LookupAccount(lookupStmt, email, false)
							if accErr != nil {
								ack.Err = accErr
							} else {
								if acc.Id != "" {
									// this account has already been registered
									// so return its verified status in the message
									if acc.Verified {
										ack.Ack = "true"
									} else {
										ack.Ack = "false"
									}
									ack.Err = nil
								}
							}
						}
					}
					WithServerDatabase(db, statusFn)
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

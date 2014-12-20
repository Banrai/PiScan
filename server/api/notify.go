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
	"text/template"
)

const (
	SHOPPING_LIST_SUBJECT = "My Items"
	SHOPPING_LIST_MESSAGE = `<p>Here are the items you selected to send to {{.Email}}</p>

	{{range $i, $item := .Items}}
	<p>
	  {{$i}}. {{$item.Desc}}
	</p>
	{{end}}`
)

type EmailedItems struct {
	Items []string
	Email string
}

func SendEmailedItems(context EmailedItems) error {
	var msg bytes.Buffer
	t := template.Must(template.New("SHOPPING_LIST_MESSAGE").Parse(SHOPPING_LIST_MESSAGE))
	err := t.Execute(&msg, context)
	if err == nil {
		sender := emailer.EmailAddress{Address: SERVER_SENDER}
		recipient := emailer.EmailAddress{Address: context.Email}
		err = emailer.Send(SHOPPING_LIST_SUBJECT, msg.String(), "text/html", &sender, &recipient, []*emailer.EmailAttachment{})
	}
	return err
}

func EmailSelectedItems(r *http.Request, db DBConnection) string {
	// the result is a simple json ack
	ack := new(SimpleMessage)

	// this function only responds to POST requests
	if "POST" == r.Method {
		r.ParseForm()

		emailVal, emailValExists := r.PostForm["email"]
		items, itemsExist := r.PostForm["item"]
		hmacDigest, hmacDigestExists := r.PostForm["hmac"]

		if emailValExists && itemsExist && hmacDigestExists {
			processFn := func(statements map[string]*sql.Stmt) {
				// see if the account exists
				accountLookupStmt, accountLookupStmtExists := statements[barcodes.ACCOUNT_LOOKUP_BY_EMAIL]
				if accountLookupStmtExists {
					// see if the email is available
					acc, accErr := barcodes.LookupAccount(accountLookupStmt, emailVal[0], false)
					if accErr != nil {
						ack.Err = accErr
					} else {
						// check the hmac digest
						r.PostForm.Del("hmac") // separate the digest from the rest
						if digest.DigestMatches(acc.APICode, r.PostForm.Encode(), hmacDigest[0]) {
							// hmac is correct

							// email the list of items
							content := EmailedItems{Email: acc.Email, Items: items}
							ack.Err = SendEmailedItems(content)

							// and update this json reply
							ack.Ack = "ok"
						}
					}
				}
			}
			WithServerDatabase(db, processFn)
		}
	}

	result, err := json.Marshal(ack)
	if err != nil {
		fmt.Println(err)
	}
	return string(result)
}

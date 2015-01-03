// Copyright Banrai LLC. All rights reserved. Use of this source code is
// governed by the license that can be found in the LICENSE file.

// Package ui provides http request handlers for the Pi client WebApp

package ui

import (
	"bytes"
	"fmt"
	"net/http"
	"os/exec"
)

func doShutdown() (string, error) {
	// Issue the system shutdown command and return the command line output,
	// which, if successful, it will be the standard system message. This
	// works on the pi b/c the default user has sudo privilege; on another
	// client device, this needs to be changed, else a sudoers update is needed.
	shutdownCmd := []string{"sudo", "shutdown", "-h", "now"}
	cmd := exec.Command(shutdownCmd[0], shutdownCmd[1:]...)

	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	return out.String(), err
}

// Issue the shutdown and write back the system command output
func ShutdownClientHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		result, err := doShutdown()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, result)
	}
}

// Copyright (c) 2022 Tailscale Inc & AUTHORS All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cli

import (
	"flag"
	"fmt"
	"strings"
	"time"
)

var (
	riskTypes     []string
	acceptedRisks string
	riskLoseSSH   = registerRiskType("lose-ssh")
)

func registerRiskType(riskType string) string {
	riskTypes = append(riskTypes, riskType)
	return riskType
}

// registerAcceptRiskFlag registers the --accept-risk flag. Accepted risks are accounted for
// in presentRiskToUser.
func registerAcceptRiskFlag(f *flag.FlagSet) {
	f.StringVar(&acceptedRisks, "accept-risk", "", "accept risk and skip confirmation for risk types: "+strings.Join(riskTypes, ","))
}

// riskAccepted reports whether riskType is in acceptedRisks.
func riskAccepted(riskType string) bool {
	for _, r := range strings.Split(acceptedRisks, ",") {
		if r == riskType {
			return true
		}
	}
	return false
}

// riskAbortTimeSeconds is the number of seconds to wait after displaying the
// risk message before continuing with the operation.
// It is used by the presentRiskToUser function below.
const riskAbortTimeSeconds = 5

// presentRiskToUser displays the risk message and waits for the user to
// cancel.
func presentRiskToUser(riskType, riskMessage string) {
	if riskAccepted(riskType) {
		return
	}
	fmt.Println(riskMessage)
	fmt.Printf("To skip this, use --accept-risk=%s\n", riskType)

	var msgLen int
	for left := riskAbortTimeSeconds; left > 0; left-- {
		msg := fmt.Sprintf("\rContinuing in %d seconds...", left)
		msgLen = len(msg)
		fmt.Print(msg)
		time.Sleep(time.Second)
	}
	fmt.Printf("\r%s\r", strings.Repeat(" ", msgLen))
}

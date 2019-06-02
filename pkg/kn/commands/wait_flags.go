package commands

import (
	"fmt"
	"github.com/spf13/cobra"
)

// Flags for tuning wait behaviour
type WaitFlags struct {
	Timeout int

	// Don't access directly, use IsWait()
	noWaitFlag bool
	waitFlag   bool
}

// Add flags which influence the sync/async behaviour when creating or updating
// resources. Set `waitDefault` argument if the default behaviour is synchronous.
// Use `what` for describing what is waited for.
func (p *WaitFlags) AddWaitFlags(command *cobra.Command, waitDefault bool, waitTimeoutDefault int, what string) {
	if waitDefault {
		noWaitUsage := fmt.Sprintf("Don't wait for %s to become ready.", what)
		command.Flags().BoolVar(&p.noWaitFlag, "no-wait", false, noWaitUsage)
	} else {
		waitUsage := fmt.Sprintf("Wait for %s to become ready.", what)
		command.Flags().BoolVar(&p.waitFlag, "wait", false, waitUsage)
	}

	timeoutUsage := fmt.Sprintf("Seconds to wait before stop checking for %s to become ready (default: %d).", what, waitTimeoutDefault)
	command.Flags().IntVar(&p.Timeout, "wait-timeout", waitTimeoutDefault, timeoutUsage)
}

// Return true if sync behaviour is requested, false otherwise. Use `waitDefault`
// for the value to be returned if the user hasn't specified any flag.
func (p *WaitFlags) IsWait(waitDefault bool) bool {
	if p.waitFlag {
		return true
	}
	if p.noWaitFlag {
		return false
	}
	return waitDefault
}

// +build !linux

package main

import (
	"github.com/rjeczalik/notify"
)

func IsInterestingEvent(e notify.Event) bool {
	switch e {
	case notify.Create, notify.Write:
		return true
	}
	return false
}

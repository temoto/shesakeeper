// +build linux

package main

import (
	"github.com/rjeczalik/notify"
)

func IsInterestingEvent(e notify.Event) bool {
	switch e {
	case notify.Create, notify.InAttrib, notify.InCloseWrite, notify.InCloseNowrite:
		return true
	}
	return false
}

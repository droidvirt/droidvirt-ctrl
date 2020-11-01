package controller

import (
	"github.com/droidvirt/droidvirt-ctrl/pkg/controller/droidvirt"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, droidvirt.Add)
}

package controller

import (
	"github.com/droidvirt/droidvirt-ctrl/pkg/controller/droidvirtvolume"
)

func init() {
	// AddToManagerFuncs is a list of functions to create controllers and add them to a manager.
	AddToManagerFuncs = append(AddToManagerFuncs, droidvirtvolume.Add)
}

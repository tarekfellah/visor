package main

import (
	"github.com/soundcloud/visor"
)

func procTypeList(snapshot visor.Snapshot, rev *visor.Revision) (procTypes string) {
	separator := ""

	if procType, err := visor.RevisionProcTypes(snapshot, rev); err == nil {
		for _, pt := range procType {
			procTypes = procTypes + separator + string(pt.Name)
			separator = ", "
		}
	}

	return
}

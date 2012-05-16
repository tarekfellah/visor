// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

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

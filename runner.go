// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"path"
	"strconv"
	"strings"
)

const runnersPath = "runners"

type Runner struct {
	Dir        dir
	Addr       string
	InstanceId int64
}

func NewRunner(addr string, instanceId int64, s Snapshot) *Runner {
	return &Runner{
		Dir:        dir{s, runnerPath(addr)},
		Addr:       addr,
		InstanceId: instanceId,
	}
}

func (r *Runner) Register() (*Runner, error) {
	f, err := createFile(r.Dir.Snapshot, r.Dir.Name, []string{strconv.FormatInt(r.InstanceId, 10)}, new(listCodec))
	if err != nil {
		return nil, err
	}
	return r.FastForward(f.Snapshot.Rev), nil
}

func (r *Runner) createSnapshot(rev int64) snapshotable {
	tmp := *r
	tmp.Dir.Snapshot = Snapshot{rev, r.Dir.Snapshot.conn}
	return &tmp
}

// FastForward advances the runner in time. It returns
// a new instance of Runner with the supplied revision.
func (r *Runner) FastForward(rev int64) *Runner {
	return r.Dir.Snapshot.fastForward(r, rev).(*Runner)
}

func RunnersByHost(s Snapshot, host string) (runners []*Runner, err error) {
	ids, err := s.getdir(path.Join(runnersPath, host))
	if err != nil {
		return
	}
	ch, errch := getSnapshotables(ids, func(id string) (snapshotable, error) {
		return GetRunner(s, runnerAddr(host, id))
	})
	for i := 0; i < len(ids); i++ {
		select {
		case r := <-ch:
			runners = append(runners, r.(*Runner))
		case e := <-errch:
			if err == nil {
				err = e
			} else {
				err = fmt.Errorf("%s\n%s", err, e)
			}
		}
	}
	return
}

func GetRunner(s Snapshot, addr string) (*Runner, error) {
	f, err := s.getFile(runnerPath(addr), new(listCodec))
	if err != nil {
		return nil, err
	}
	data := f.Value.([]string)
	insIdStr := data[0]
	insId, err := parseInstanceId(insIdStr)
	if err != nil {
		return nil, err
	}

	return NewRunner(addr, insId, s), nil
}

func WatchRunnerStart(host string, s Snapshot, ch chan *Runner, errch chan error) {
	rev := s.Rev

	for {
		ev, err := s.conn.Wait(path.Join(runnersPath, host, "*"), rev+1)
		if err != nil {
			errch <- err
			return
		}
		rev = ev.Rev

		if !ev.IsSet() {
			continue
		}
		port := strings.Split(ev.Path, "/")[3]
		addr := runnerAddr(host, port)

		runner, err := GetRunner(s.FastForward(rev), addr)
		if err != nil {
			errch <- err
			return
		}
		ch <- runner
	}
}

func runnerAddr(host, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}

func runnerPath(addr string) string {
	parts := strings.Split(addr, ":")
	return path.Join(runnersPath, parts[0], parts[1])
}

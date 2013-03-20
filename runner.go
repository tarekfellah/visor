// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"bufio"
	"fmt"
	cp "github.com/soundcloud/cotterpin"
	"github.com/soundcloud/visor/net"
	"io"
	"path"
	"strconv"
	"strings"
	"time"
)

const runnersPath = "runners"

type RStatus string

const (
	RunnerUp      RStatus = "up"
	RunnerDown    RStatus = "down"
	RunnerUnknown RStatus = "unknown"
)

type Runner struct {
	Dir        cp.Dir
	Addr       string
	InstanceId int64
	conn       io.ReadWriteCloser
	net        net.Network
}

func (s *Store) NewRunner(addr string, instanceId int64, network net.Network) *Runner {
	return &Runner{
		Dir:        cp.Dir{s.GetSnapshot(), runnerPath(addr)},
		Addr:       addr,
		InstanceId: instanceId,
		net:        network,
	}
}

func (r *Runner) GetSnapshot() cp.Snapshot {
	return r.Dir.Snapshot
}

// Join advances the Runner in time. It returns
// a new instance of Runner at the rev of the
// supplied cp.Snapshotable.
func (r *Runner) Join(s cp.Snapshotable) *Runner {
	tmp := *r
	tmp.Dir.Snapshot = s.GetSnapshot()
	return &tmp
}

func (r *Runner) Register() (*Runner, error) {
	f := cp.NewFile(r.Dir.Name, []string{strconv.FormatInt(r.InstanceId, 10)}, new(cp.ListCodec), r.Dir.Snapshot)
	f, err := f.Save()
	if err != nil {
		return nil, err
	}
	return r.Join(f), nil
}

func (r *Runner) Unregister() error {
	return r.Dir.Del("/")
}

func (r *Runner) Connect() (err error) {
	if r.conn != nil {
		return
	}

	for i := 0; i < 3; i++ {
		var conn io.ReadWriteCloser

		conn, err = r.net.Dial(r.Addr)
		if err == nil {
			r.conn = conn
			_, err = r.cmd("raw") // Activate 'raw' mode
			break
		}
		time.Sleep(time.Second / 2)
	}
	return
}

func (r *Runner) Disconnect() {
	r.conn.Close()
	r.conn = nil
}

func (r *Runner) send(cmd string) (err error) {
	_, err = fmt.Fprintf(r.conn, "%s\n", cmd)

	return
}

func (r *Runner) cmd(c string) (out string, err error) {
	if err = r.send(c); err != nil {
		return
	}
	out, err = bufio.NewReader(r.conn).ReadString('\n')

	return
}

func (r *Runner) GetPid() (pid int, err error) {
	out, err := r.cmd("pid")
	if err != nil {
		return -1, err
	}
	return strconv.Atoi(strings.TrimSpace(out))
}

func (r *Runner) GetStatus() (s RStatus, srvRestarts, logRestarts int, err error) {
	out, err := r.cmd("status")
	if err != nil {
		return RunnerUnknown, -1, -1, err
	}
	fields := strings.Fields(out) // example: "up 6 0"

	s = RStatus(fields[0])

	srvRestarts, err = strconv.Atoi(fields[1])
	if err != nil {
		return RunnerUnknown, -1, -1, err
	}

	logRestarts, err = strconv.Atoi(fields[2])
	if err != nil {
		return RunnerUnknown, -1, -1, err
	}

	return
}

func (r *Runner) Exit() error {
	return r.send("kill")
}

func (r *Runner) Down() error {
	out, err := r.cmd("down")
	if out != "OK\n" {
		return fmt.Errorf(out)
	}
	return err
}

func (s *Store) Runners() (runners []*Runner, err error) {
	hosts, err := s.GetSnapshot().Getdir(runnersPath)
	if err != nil {
		return
	}

	for _, host := range hosts {
		rns, err := s.RunnersByHost(host)
		if err != nil {
			return runners, err
		}
		runners = append(runners, rns...)
	}
	return
}

func (s *Store) RunnersByHost(host string) (runners []*Runner, err error) {
	ids, err := s.GetSnapshot().Getdir(path.Join(runnersPath, host))
	if err != nil {
		return
	}
	ch, errch := cp.GetSnapshotables(ids, func(id string) (cp.Snapshotable, error) {
		return s.GetRunner(runnerAddr(host, id))
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

func (s *Store) GetRunner(addr string) (*Runner, error) {
	f, err := s.GetSnapshot().GetFile(runnerPath(addr), new(cp.ListCodec))
	if err != nil {
		return nil, err
	}
	data := f.Value.([]string)
	insIdStr := data[0]
	insId, err := parseInstanceId(insIdStr)
	if err != nil {
		return nil, err
	}

	return s.NewRunner(addr, insId, new(net.Net)), nil
}

func (s *Store) WatchRunnerStart(host string, ch chan *Runner, errch chan error) {
	rev := s.GetSnapshot().Rev
	for {
		ev, err := waitRunnersByHost(s, host, rev)
		if err != nil {
			errch <- err
			return
		}
		rev = ev.Rev

		if !ev.IsSet() {
			continue
		}
		addr := addrFromPath(ev.Path)

		runner, err := s.Join(ev).GetRunner(addr)
		if err != nil {
			errch <- err
			return
		}
		ch <- runner
	}
}

func (s *Store) WatchRunnerStop(host string, ch chan string, errch chan error) {
	rev := s.GetSnapshot().Rev
	for {
		ev, err := waitRunnersByHost(s, host, rev)
		if err != nil {
			errch <- err
			return
		}
		rev = ev.Rev

		if !ev.IsDel() {
			continue
		}
		ch <- addrFromPath(ev.Path)
	}
}

func addrFromPath(path string) string {
	parts := strings.Split(path, "/")
	addr := runnerAddr(parts[2], parts[3])

	return addr
}

func waitRunnersByHost(s *Store, host string, rev int64) (cp.Event, error) {
	sp := s.GetSnapshot()
	return sp.Wait(path.Join(runnersPath, host, "*"), rev+1)
}

func runnerAddr(host, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}

func runnerPath(addr string) string {
	parts := strings.Split(addr, ":")
	return path.Join(runnersPath, parts[0], parts[1])
}

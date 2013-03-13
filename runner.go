// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"bufio"
	"fmt"
	"github.com/soundcloud/doozer"
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
	Dir        dir
	Addr       string
	InstanceId int64
	conn       io.ReadWriteCloser
	net        net.Network
}

func NewRunner(addr string, instanceId int64, network net.Network, s Snapshot) *Runner {
	return &Runner{
		Dir:        dir{s, runnerPath(addr)},
		Addr:       addr,
		InstanceId: instanceId,
		net:        network,
	}
}

func (r *Runner) Register() (*Runner, error) {
	f, err := createFile(r.Dir.Snapshot, r.Dir.Name, []string{strconv.FormatInt(r.InstanceId, 10)}, new(listCodec))
	if err != nil {
		return nil, err
	}
	return r.FastForward(f.Snapshot.Rev), nil
}

func (r *Runner) Unregister() error {
	return r.Dir.del("/")
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

func Runners(s Snapshot) (runners []*Runner, err error) {
	hosts, err := s.getdir(runnersPath)
	if err != nil {
		return
	}

	for _, host := range hosts {
		rns, err := RunnersByHost(s, host)
		if err != nil {
			return runners, err
		}
		runners = append(runners, rns...)
	}
	return
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

	return NewRunner(addr, insId, new(net.Net), s), nil
}

func WatchRunnerStart(host string, s Snapshot, ch chan *Runner, errch chan error) {
	rev := s.Rev

	for {
		ev, err := waitRunnersByHost(host, rev, s)
		if err != nil {
			errch <- err
			return
		}
		rev = ev.Rev

		if !ev.IsSet() {
			continue
		}
		addr := addrFromPath(ev.Path)

		runner, err := GetRunner(s.FastForward(rev), addr)
		if err != nil {
			errch <- err
			return
		}
		ch <- runner
	}
}

func WatchRunnerStop(host string, s Snapshot, ch chan string, errch chan error) {
	rev := s.Rev

	for {
		ev, err := waitRunnersByHost(host, rev, s)
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

func waitRunnersByHost(host string, rev int64, s Snapshot) (doozer.Event, error) {
	return s.conn.Wait(path.Join(runnersPath, host, "*"), rev+1)
}

func runnerAddr(host, port string) string {
	return fmt.Sprintf("%s:%s", host, port)
}

func runnerPath(addr string) string {
	parts := strings.Split(addr, ":")
	return path.Join(runnersPath, parts[0], parts[1])
}

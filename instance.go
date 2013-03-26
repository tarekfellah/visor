// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	cp "github.com/soundcloud/cotterpin"
	"path"
	"strconv"
	"strings"
)

const claimsPath = "claims"
const instancesPath = "instances"
const failedPath = "failed"
const startPath = "start"
const statusPath = "status"
const stopPath = "stop"
const restartsPath = "restarts"

type InsRestarts struct {
	OOM, Fail int
}

func (r *InsRestarts) Fields() []int {
	return []int{r.Fail, r.OOM}
}

type RestartReason string

const (
	RestartFail = "restart-fail"
	RestartOOM  = "restart-oom"
)

const (
	restartFailField = 0
	restartOOMField  = 1
)

type InsStatus string

const (
	InsStatusPending  InsStatus = "pending"
	InsStatusClaimed            = "claimed"
	InsStatusRunning            = "running"
	InsStatusStopping           = "stopping"

	InsStatusFailed = "failed"
	InsStatusExited = "exited"
)

// Instance represents application instances.
type Instance struct {
	Dir          cp.Dir
	Id           int64
	AppName      string
	RevisionName string
	ProcessName  string
	Ip           string
	Port         int
	Host         string
	Status       InsStatus
	Restarts     *InsRestarts
}

func (s *Store) Instances() (ins []*Instance, err error) {
	ids, err := s.GetSnapshot().Getdir("instances")
	if err != nil {
		return
	}
	ch, errch := cp.GetSnapshotables(ids, func(idstr string) (cp.Snapshotable, error) {
		id, err := parseInstanceId(idstr)
		if err != nil {
			return nil, err
		}
		return s.GetInstance(id)
	})
	for i := 0; i < len(ids); i++ {
		select {
		case r := <-ch:
			ins = append(ins, r.(*Instance))
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

// GetInstance returns an Instance from the given id
func (s *Store) GetInstance(id int64) (ins *Instance, err error) {
	p := instancePath(id)
	status := InsStatusPending

	var (
		ip   string
		port int
		host string
	)

	f, err := s.GetSnapshot().GetFile(p+"/start", new(cp.ListCodec))
	if cp.IsErrNoEnt(err) {
		// Ignore
	} else if err != nil {
		return
	} else {
		fields := f.Value.([]string)

		if len(fields) > 0 { // IP
			status = InsStatusClaimed
			ip = fields[0]
		}
		if len(fields) > 1 { // Port
			status = InsStatusRunning
			port, err = strconv.Atoi(fields[1])
			if err != nil {
				panic("invalid port number: " + fields[1])
			}
		}
		if len(fields) > 2 { // Hostname
			host = fields[2]
		}
	}

	statusStr, _, err := s.GetSnapshot().Get(p + "/status")
	if cp.IsErrNoEnt(err) {
		err = nil
	} else if err == nil {
		status = InsStatus(statusStr)
	} else {
		return
	}

	if status == InsStatusRunning {
		_, _, err := s.GetSnapshot().Get(p + "/stop")
		if err == nil {
			status = InsStatusStopping
		} else if !cp.IsErrNoEnt(err) {
			return nil, err
		}
	}

	f, err = s.GetSnapshot().GetFile(p+"/object", new(cp.ListCodec))
	if err != nil {
		return nil, ErrNotFound
	}
	fields := f.Value.([]string)

	ins = &Instance{
		Id:           id,
		AppName:      fields[0],
		RevisionName: fields[1],
		ProcessName:  fields[2],
		Status:       status,
		Ip:           ip,
		Port:         port,
		Host:         host,
		Dir:          cp.Dir{s.GetSnapshot(), instancePath(id)},
	}

	restarts, _, err := ins.getRestarts()
	if err != nil {
		return
	}
	ins.Restarts = restarts

	return
}

func getInstanceIds(sp cp.Snapshotable, app, rev, pty string) (ids []int64, err error) {
	s := sp.GetSnapshot()
	p := ptyInstancesPath(app, rev, pty)
	exists, _, err := s.Exists(p)
	if err != nil || !exists {
		return
	}

	dir, err := s.Getdir(p)
	if err != nil {
		return
	}
	ids = []int64{}
	for _, f := range dir {
		id, e := parseInstanceId(f)
		if e != nil {
			return nil, e
		}
		ids = append(ids, id)
	}
	return
}

func (s *Store) RegisterInstance(app string, rev string, pty string) (ins *Instance, err error) {
	//
	//   instances/
	//       6868/
	// +         object = <app> <rev> <proc>
	// +         start  =
	//
	//   apps/<app>/procs/<proc>/instances/<rev>
	// +     6868 = 2012-07-19 16:41 UTC
	//
	id, err := s.GetSnapshot().Getuid()
	if err != nil {
		return
	}
	ins = &Instance{
		Id:           id,
		AppName:      app,
		RevisionName: rev,
		ProcessName:  pty,
		Status:       InsStatusPending,
		Dir:          cp.Dir{s.GetSnapshot(), instancePath(id)},
		Restarts:     new(InsRestarts),
	}

	object := cp.NewFile(ins.Dir.Prefix("object"), ins.objectArray(), new(cp.ListCodec), s.GetSnapshot())
	object, err = object.Save()
	if err != nil {
		return nil, err
	}
	_, err = s.GetSnapshot().Set(ins.ptyInstancesPath(), timestamp())
	if err != nil {
		return
	}
	start := cp.NewFile(ins.Dir.Prefix(startPath), "", new(cp.StringCodec), s.GetSnapshot())
	start, err = start.Save()
	if err != nil {
		return nil, err
	}
	ins = ins.Join(start.Snapshot)

	return
}

func (s *Store) StopInstance(id int64) (*Store, error) {
	//
	//   instances/
	//       6868/
	//           ...
	// +         stop =
	//
	// TODO Check that instance is started
	d := &cp.Dir{s.GetSnapshot(), instancePath(id)}
	d, err := d.Set("stop", "")
	if err != nil {
		return nil, err
	}

	return s.Join(d), nil
}

func instancePath(id int64) string {
	return path.Join(instancesPath, strconv.FormatInt(id, 10))
}

func (i *Instance) GetSnapshot() cp.Snapshot {
	return i.Dir.Snapshot
}

// Join advances the Instance in time. It returns a new
// instance of Instance at the rev of the supplied
// cp.Snapshotable.
func (i *Instance) Join(s cp.Snapshotable) *Instance {
	tmp := *i
	tmp.Dir.Snapshot = s.GetSnapshot()
	return &tmp
}

// Claims returns the list of claimers.
func (i *Instance) Claims() (claims []string, err error) {
	claims, err = i.Dir.Snapshot.Getdir(i.Dir.Prefix("claims"))
	if cp.IsErrNoEnt(err) {
		claims = []string{}
		err = nil
	}
	return
}

// Claim locks the instance to the specified host.
func (i *Instance) Claim(host string) (*Instance, error) {
	//
	//   instances/
	//       6868/
	//           claims/
	// +             10.0.0.1 = 2012-07-19 16:22 UTC
	//           object = <app> <rev> <proc>
	// -         start  =
	// +         start  = 10.0.0.1
	//
	f, err := i.Dir.GetFile(startPath, new(cp.ListCodec))
	if err != nil {
		return nil, err
	}
	fields := f.Value.([]string)
	if len(fields) > 0 {
		return nil, ErrInsClaimed
	}
	d := i.Dir.Join(f)

	d, err = d.Set(startPath, host)
	if err != nil {
		return i, err
	}

	d, err = i.claimDir().Join(d).Set(host, timestamp())
	if err != nil {
		return i, err
	}
	return i.Join(d), err
}

func (i *Instance) Unregister() (err error) {
	var path string

	if i.Status == InsStatusFailed {
		path = i.ptyFailedPath()
	} else {
		path = i.ptyInstancesPath()
	}
	err = i.Dir.Snapshot.Del(path)
	if err != nil {
		if cp.IsErrNoEnt(err) {
			err = nil
		} else {
			return
		}
	}
	err = i.Dir.Del("/")
	return
}

// Exited tells the coordinator that the instance has exited.
func (i *Instance) Exited(host string) (i1 *Instance, err error) {
	if err = i.verifyClaimer(host); err != nil {
		return
	}
	i1, err = i.updateStatus(InsStatusExited)
	if err != nil {
		return nil, err
	}
	err = i.Dir.Snapshot.Del(i.ptyInstancesPath())

	return
}

func (i *Instance) started(ip string, port int, host string) {
	i.Ip = ip
	i.Port = port
	i.Host = host
	i.Status = InsStatusRunning
}

func (i *Instance) claimed(ip string) {
	i.Ip = ip
	i.Status = InsStatusClaimed
}

func (i *Instance) getRestarts() (*InsRestarts, *cp.File, error) {
	restarts := new(InsRestarts)

	f, err := i.Dir.GetFile(i.Dir.Prefix(restartsPath), new(cp.ListIntCodec))
	if err == nil {
		fields := f.Value.([]int)

		restarts.Fail = fields[restartFailField]

		if len(fields) > 1 {
			restarts.OOM = fields[restartOOMField]
		}
	} else if !cp.IsErrNoEnt(err) {
		return nil, nil, err
	}
	return restarts, f, nil
}

func (i *Instance) Started(host string, port int, hostname string) (i1 *Instance, err error) {
	//
	//   instances/
	//       6868/
	//           object = <app> <rev> <proc>
	// -         start  = 10.0.0.1
	// +         start  = 10.0.0.1 24690 localhost
	//
	if i.Status == InsStatusRunning {
		return i, nil
	}
	if err = i.verifyClaimer(host); err != nil {
		return
	}
	i1 = i.Join(&i.Dir) // Create a copy
	i1.started(host, port, hostname)

	start := cp.NewFile(i1.Dir.Prefix(startPath), i1.startArray(), new(cp.ListCodec), i1.Dir.Snapshot)
	start, err = start.Save()
	if err != nil {
		return
	}
	i1 = i1.Join(start)

	return
}

// Restarted tells the coordinator that the instance has been restarted.
func (i *Instance) Restarted(reason RestartReason, count int) (i1 *Instance, err error) {
	//
	//   instances/
	//       6868/
	//           object   = <app> <rev> <proc>
	//           start    = 10.0.0.1 24690 localhost
	// -         restarts = 1 4
	// +         restarts = 2 4
	//
	//   instances/
	//       6869/
	//           object   = <app> <rev> <proc>
	//           start    = 10.0.0.1 24691 localhost
	// +         restarts = 1 0
	//
	if i.Status != InsStatusRunning {
		return i, nil
	}

	restarts, f, err := i.getRestarts()
	if err != nil {
		return
	}

	switch reason {
	case RestartFail:
		i.Restarts.Fail = restarts.Fail + count
	case RestartOOM:
		i.Restarts.OOM = restarts.OOM + count
	}

	f, err = f.Set(i.Restarts.Fields())
	if err != nil {
		return
	}

	i1 = i.Join(f)

	return
}

func (i *Instance) updateStatus(s InsStatus) (i1 *Instance, err error) {
	d, err := i.Dir.Set("status", string(s))
	if err != nil {
		return
	}
	i.Status = s

	return i.Join(d), err
}

func (i *Instance) getClaimer() (*string, error) {
	f, err := i.Dir.Snapshot.GetFile(i.Dir.Prefix(startPath), new(cp.ListCodec))

	if cp.IsErrNoEnt(err) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	fields := f.Value.([]string)

	if len(fields) == 0 {
		return nil, nil
	}
	return &fields[0], nil
}

func (i *Instance) setClaimer(claimer string) (*cp.Dir, error) {
	d, err := i.Dir.Set(startPath, claimer)
	if err != nil {
		return nil, err
	}
	return d, nil
}

// Unclaim removes the lock applied by Claim of the Ticket.
func (i *Instance) Unclaim(host string) (i1 *Instance, err error) {
	//
	//   instances/
	//       6868/
	// -         start = 10.0.0.1
	// +         start =
	//
	if err = i.verifyClaimer(host); err != nil {
		return
	}

	d, err := i.setClaimer("")
	if err != nil {
		return
	}
	i1 = i.Join(d)

	return
}

func (i *Instance) verifyClaimer(host string) error {
	claimer, err := i.getClaimer()
	if err != nil {
		return err
	} else if claimer == nil || *claimer != host {
		return ErrUnauthorized
	}
	return nil
}

func (i *Instance) Failed(host string, reason error) (i1 *Instance, err error) {
	if err = i.verifyClaimer(host); err != nil {
		return
	}
	_, err = i.updateStatus(InsStatusFailed)
	if err != nil {
		return
	}
	s, err := i.Dir.Snapshot.Set(i.ptyFailedPath(), timestamp()+" "+reason.Error())
	if err != nil {
		return
	}
	err = i.Dir.Snapshot.Del(i.ptyInstancesPath())
	if err != nil {
		return
	}
	i1 = i.Join(s)

	return
}

func (s *Store) WatchInstanceStart(listener chan *Instance, errors chan error) {
	// instances/*/start =
	rev := s.GetSnapshot().Rev
	for {
		ev, err := s.snapshot.Wait(path.Join(instancesPath, "*", startPath), rev+1)
		rev = ev.Rev

		if !ev.IsSet() || string(ev.Body) != "" {
			continue
		}
		idstr := strings.Split(ev.Path, "/")[2]

		id, err := parseInstanceId(idstr)
		if err != nil {
			panic(err)
		}
		ins, err := s.Join(ev).GetInstance(id)
		if err != nil {
			panic(err)
		}
		listener <- ins
	}
}

func (i *Instance) WaitStop() (i1 *Instance, err error) {
	p := path.Join(instancesPath, strconv.FormatInt(i.Id, 10), stopPath)

	ev, err := i.Dir.Snapshot.Wait(p, i.Dir.Snapshot.Rev+1)
	if err != nil {
		return
	}
	i1 = i.Join(ev)

	return
}

func (i *Instance) WaitStatus() (i1 *Instance, err error) {
	p := path.Join(instancesPath, strconv.FormatInt(i.Id, 10), statusPath)
	ev, err := i.Dir.Snapshot.Wait(p, i.Dir.Snapshot.Rev+1)
	if err != nil {
		return
	}
	i1 = i.Join(ev)
	i1.Status = InsStatus(string(ev.Body))

	return
}

func (i *Instance) WaitClaimed() (i1 *Instance, err error) {
	return i.waitStartPathStatus(InsStatusClaimed)
}

func (i *Instance) WaitStarted() (i1 *Instance, err error) {
	return i.waitStartPathStatus(InsStatusRunning)
}

func (i *Instance) WaitExited() (*Instance, error) {
	for {
		i, err := i.WaitStatus()
		if err != nil {
			return nil, err
		}
		if i.Status == InsStatusExited {
			break
		}
	}
	return i, nil
}

func (i *Instance) WaitFailed() (*Instance, error) {
	for {
		i, err := i.WaitStatus()
		if err != nil {
			return nil, err
		}
		if i.Status == InsStatusFailed {
			break
		}
	}
	return i, nil
}

func (i *Instance) GetStatusInfo() (info string, err error) {
	if i.Status == InsStatusFailed {
		info, _, err = i.Dir.Snapshot.Get(i.ptyFailedPath())
		info = strings.TrimSpace(info)
	}
	return
}

func (i *Instance) waitStartPathStatus(s InsStatus) (i1 *Instance, err error) {
	for {
		i, err = i.waitStartPath()
		if err != nil {
			return i, err
		}
		if i.Status == s {
			break
		}
	}
	return i, nil
}

func (i *Instance) waitStartPath() (i1 *Instance, err error) {
	p := path.Join(instancesPath, strconv.FormatInt(i.Id, 10), startPath)

	ev, err := i.Dir.Snapshot.Wait(p, i.Dir.Snapshot.Rev+1)
	if err != nil {
		return
	}
	i1 = i.Join(ev)
	parts, err := new(cp.ListCodec).Decode(ev.Body)
	if err != nil {
		return
	}
	fields := parts.([]string)
	if len(fields) >= 3 {
		ip, host := fields[0], fields[2]
		port, err := strconv.Atoi(fields[1])
		if err != nil {
			return i, err
		}
		i1.started(ip, port, host)
	} else if len(fields) > 0 {
		i1.claimed(fields[0])
	} else {
		// TODO
	}
	return
}

func ptyInstancesPath(app, rev, pty string) string {
	return path.Join(appsPath, app, procsPath, pty, instancesPath, rev)
}

func parseInstanceId(idstr string) (int64, error) {
	return strconv.ParseInt(idstr, 10, 64)
}

func (i *Instance) idString() string {
	return fmt.Sprintf("%d", i.Id)
}

func (i *Instance) RefString() string {
	return fmt.Sprintf("%s:%s@%s", i.AppName, i.ProcessName, i.RevisionName)
}

func (i *Instance) ServiceName() string {
	return fmt.Sprintf("%s-%s", i.AppName, i.ProcessName)
}

func (i *Instance) ptyFailedPath() string {
	return path.Join(appsPath, i.AppName, procsPath, i.ProcessName, failedPath, i.idString())
}

func (i *Instance) ptyInstancesPath() string {
	return path.Join(appsPath, i.AppName, procsPath, i.ProcessName, instancesPath, i.RevisionName, i.idString())
}

func (i *Instance) claimPath(host string) string {
	return i.Dir.Prefix("claims", host)
}

func (i *Instance) claimDir() *cp.Dir {
	// TODO move to factory for Dir creation
	return &cp.Dir{i.Dir.Snapshot, i.Dir.Prefix(claimsPath)}
}

func (i *Instance) Fields() string {
	return fmt.Sprintf("%d %s %s %s %s %d", i.Id, i.AppName, i.RevisionName, i.ProcessName, i.Ip, i.Port)
}

func (i *Instance) objectArray() []string {
	return []string{i.AppName, i.RevisionName, i.ProcessName}
}

func (i *Instance) startArray() []string {
	return []string{i.Ip, i.portString(), i.Host}
}

func (i *Instance) portString() string {
	return fmt.Sprintf("%d", i.Port)
}

// String returns the Go-syntax representation of Instance.
func (i *Instance) String() string {
	return fmt.Sprintf("Instance{id=%d, app=%s, rev=%s, proc=%s, addr=%s:%d}", i.Id, i.AppName, i.RevisionName, i.ProcessName, i.Ip, i.Port)
}

// IdString returns a string of the format "INSTANCE[id]"
func (i *Instance) IdString() string {
	return fmt.Sprintf("INSTANCE[%d]", i.Id)
}

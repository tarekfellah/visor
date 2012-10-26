// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"fmt"
	"github.com/soundcloud/doozer"
	"path"
	"strconv"
	"strings"
	"time"
)

const claimsPath = "claims"
const instancesPath = "instances"
const deathsPath = "deaths"
const startPath = "start"
const stopPath = "stop"

const (
	InsStatusInitial     InsStatus = "initial"
	InsStatusClaimed               = "claimed"
	InsStatusStarted               = "started"
	InsStatusFailed                = "failed"
	InsStatusDead                  = "dead"
	InsStatusUnclaimable           = "unclaimable"
	InsStatusExited                = "exited"
)

// Instance represents application instances.
type Instance struct {
	dir
	Id           int64
	AppName      string
	RevisionName string
	ProcessName  string
	Ip           string
	Port         int
	Host         string
	Status       InsStatus
}

// GetInstance returns an Instance from the given id
func GetInstance(s Snapshot, id int64) (ins *Instance, err error) {
	p := instancePath(id)

	var (
		status InsStatus
		ip     string
		port   int
		host   string
	)

	f, err := s.getFile(p+"/start", new(listCodec))
	if IsErrNoEnt(err) {
	} else if err != nil {
		return
	} else {
		fields := f.Value.([]string)

		if len(fields) > 0 { // IP
			ip = fields[0]
		}
		if len(fields) > 1 { // Port
			status = InsStatusStarted
			port, err = strconv.Atoi(fields[1])
			if err != nil {
				panic("invalid port number: " + fields[1])
			}
		}
		if len(fields) > 2 { // Hostname
			host = fields[2]
		}
	}

	statusStr, _, err := s.get(p + "/status")
	if IsErrNoEnt(err) {
		err = nil
		status = InsStatusInitial
	} else if err == nil {
		status = InsStatus(statusStr)
	} else {
		return
	}

	f, err = s.getFile(p+"/object", new(listCodec))
	if err != nil {
		return
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
		dir:          dir{s, instancePath(id)},
	}
	return
}

func getInstanceIds(s Snapshot, app, rev, pty string) (ids []int64, err error) {
	p := ptyInstancesPath(app, rev, pty)
	exists, _, err := s.conn.Exists(p)
	if err != nil || !exists {
		return
	}

	dir, err := s.getdir(p)
	if err != nil {
		return
	}
	ids = []int64{}
	for _, f := range dir {
		id, e := strconv.ParseInt(f, 10, 64)
		if e != nil {
			return nil, e
		}
		ids = append(ids, id)
	}
	return
}

func RegisterInstance(app string, rev string, pty string, s Snapshot) (ins *Instance, err error) {
	//
	//   instances/
	//       6868/
	// +         object = <app> <rev> <proc>
	// +         start  =
	//
	//   apps/<app>/procs/<proc>/instances/<rev>
	// +     6868 = 2012-07-19 16:41 UTC
	//
	id, err := Getuid(s)
	if err != nil {
		return
	}
	ins = &Instance{
		Id:           id,
		AppName:      app,
		RevisionName: rev,
		ProcessName:  pty,
		Status:       InsStatusInitial,
		dir:          dir{s, instancePath(id)},
	}

	_, err = createFile(s, ins.dir.prefix("object"), ins.objectArray(), new(listCodec))
	if err != nil {
		return nil, err
	}
	_, err = createFile(s, ins.dir.prefix(startPath), "", new(stringCodec))
	if err != nil {
		return nil, err
	}
	s1, err := s.set(ins.ptyInstancesPath(), time.Now().UTC().String())
	if err != nil {
		return
	}
	ins = ins.FastForward(s1.Rev)

	return
}

func StopInstance(id int64, s Snapshot) (s1 Snapshot, err error) {
	//
	//   instances/
	//       6868/
	//           ...
	// +         stop =
	//
	// TODO Check that instance is started
	d := dir{s, instancePath(id)}
	rev, err := d.set("stop", "")
	if err != nil {
		return
	}
	s1 = s.FastForward(rev)

	return
}

func instancePath(id int64) string {
	return path.Join(instancesPath, strconv.FormatInt(id, 10))
}

// FastForward advances the instance in time. It returns
// a new instance of Instance with the supplied revision.
func (i *Instance) FastForward(rev int64) *Instance {
	return i.Snapshot.fastForward(i, rev).(*Instance)
}

func (i *Instance) createSnapshot(rev int64) snapshotable {
	tmp := *i
	tmp.Snapshot = Snapshot{rev, i.conn}
	return &tmp
}

// Claims returns the list of claimers.
func (i *Instance) Claims() (claims []string, err error) {
	rev, err := i.conn.Rev()
	if err != nil {
		return
	}
	claims, err = i.conn.Getdir(i.dir.prefix("claims"), rev)
	if err, ok := err.(*doozer.Error); ok && err.Err == doozer.ErrNoEnt {
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
	val, rev, err := i.dir.get(startPath)
	if err != nil {
		return nil, err
	}
	if val != "" {
		return nil, ErrInsClaimed
	}
	d := i.dir.fastForward(rev)

	_, err = d.set(startPath, host)
	if err != nil {
		return i, err
	}

	rev, err = i.claimDir().fastForward(rev).set(host, time.Now().UTC().String())
	if err != nil {
		return i, err
	}
	return i.FastForward(rev), err
}

func (i *Instance) Unregister() (err error) {
	err = i.Snapshot.del(i.ptyInstancesPath())
	if err != nil {
		println(IsErrNoEnt(err))
		println(err.Error())
		if IsErrNoEnt(err) {
			err = nil
		} else {
			return
		}
	}
	err = i.del("/")
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
	err = i.Snapshot.del(i.ptyInstancesPath())

	return
}

// Failed tells the cooridnator that the instance has failed.
func (i *Instance) Failed(host string, reason error) (i1 *Instance, err error) {
	if err = i.verifyClaimer(host); err != nil {
		return
	}
	i1, err = i.updateStatus(InsStatusFailed)
	if err != nil {
		return
	}
	_, err = i1.claimDir().set(host, fmt.Sprintf("%s %s", time.Now().UTC().String(), reason))

	return
}

func (i *Instance) started(ip string, port int, host string) {
	i.Ip = ip
	i.Port = port
	i.Host = host
	i.Status = InsStatusStarted
}

func (i *Instance) claimed(ip string) {
	i.Ip = ip
	i.Status = InsStatusClaimed
}

func (i *Instance) Started(ip string, port int, host string) (i1 *Instance, err error) {
	//
	//   instances/
	//       6868/
	//           object = <app> <rev> <proc>
	// -         start  = 10.0.0.1
	// +         start  = 10.0.0.1 24690 localhost
	//
	if err = i.verifyClaimer(ip); err != nil {
		return
	}
	i.started(ip, port, host)

	f, err := createFile(i.Snapshot, i.dir.prefix(startPath), i.startArray(), new(listCodec))
	if err != nil {
		return
	}
	i1 = i.FastForward(f.FileRev)

	return
}

func (i *Instance) updateStatus(s InsStatus) (i1 *Instance, err error) {
	rev, err := i.set("status", string(s))
	if err != nil {
		return
	}
	i.Status = s

	return i.FastForward(rev), err
}

func (i *Instance) getClaimer() (*string, error) {
	f, err := i.getFile(i.dir.prefix(startPath), new(listCodec))

	if IsErrNoEnt(err) {
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

func (i *Instance) setClaimer(claimer string) (rev int64, err error) {
	return i.set(startPath, claimer)
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

	rev, err := i.setClaimer("")
	if err != nil {
		return
	}
	i1 = i.FastForward(rev)

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

func (i *Instance) Unclaimable(host string, reason error) (i1 *Instance, err error) {
	if err = i.verifyClaimer(host); err != nil {
		return
	}
	i1, err = i.updateStatus(InsStatusUnclaimable)
	if err != nil {
		return
	}
	rev, err := i1.claimDir().set(host, fmt.Sprintf("%s %s", time.Now().UTC().String(), reason))
	if err != nil {
		return
	}
	i1 = i1.FastForward(rev)

	return
}

func (i *Instance) Dead(host string, reason error) (i1 *Instance, err error) {
	if err = i.verifyClaimer(host); err != nil {
		return
	}
	_, err = i.updateStatus(InsStatusDead)
	if err != nil {
		return
	}
	s, err := i.Snapshot.set(i.ptyDeathsPath(), reason.Error())
	if err != nil {
		return
	}
	err = i.Snapshot.del(i.ptyInstancesPath())
	if err != nil {
		return
	}
	i1 = i.FastForward(s.Rev)

	return
}

func WatchInstanceStart(s Snapshot, listener chan *Instance, errors chan error) {
	// instances/*/start =
	rev := s.Rev

	for {
		ev, err := s.conn.Wait(path.Join(instancesPath, "*", startPath), rev+1)
		rev = ev.Rev

		if !ev.IsSet() || string(ev.Body) != "" {
			continue
		}
		idstr := strings.Split(ev.Path, "/")[2]

		id, err := strconv.ParseInt(idstr, 0, 64)
		if err != nil {
			panic(err)
		}
		ins, err := GetInstance(s.FastForward(rev), id)
		if err != nil {
			panic(err)
		}
		listener <- ins
	}
}

func (i *Instance) WaitStop() (i1 *Instance, err error) {
	p := path.Join(instancesPath, strconv.FormatInt(i.Id, 10), stopPath)

	ev, err := i.Snapshot.conn.Wait(p, i.Rev+1)
	if err != nil {
		return
	}
	i1 = i.FastForward(ev.Rev)

	return
}

func (i *Instance) WaitStart() (i1 *Instance, err error) {
	p := path.Join(instancesPath, strconv.FormatInt(i.Id, 10), startPath)

	ev, err := i.Snapshot.conn.Wait(p, i.Rev+1)
	if err != nil {
		return
	}
	i1 = i.FastForward(ev.Rev)
	parts, err := new(listCodec).Decode(ev.Body)
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
	} else {
		i1.claimed(fields[0])
	}
	return
}

func WaitTicketProcessed(s Snapshot, id int64) (status InsStatus, s1 Snapshot, err error) {
	var ev doozer.Event

	rev := s.Rev

	for {
		ev, err = s.conn.Wait(fmt.Sprintf("/%s/%d/status", instancesPath, id), rev+1)
		if err != nil {
			return
		}
		rev = ev.Rev

		// TODO
		//if ev.IsSet() && InsStatus(ev.Body) == InsStatusDone {
		//	status = InsStatusDone
		//	break
		//}
		if ev.IsSet() && InsStatus(ev.Body) == InsStatusDead {
			status = InsStatusDead
			break
		}
	}
	s1 = s.FastForward(rev)

	return
}

func ptyInstancesPath(app, rev, pty string) string {
	return path.Join(appsPath, app, procsPath, pty, instancesPath, rev)
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

func (i *Instance) ptyDeathsPath() string {
	return path.Join(appsPath, i.AppName, procsPath, i.ProcessName, deathsPath, i.idString())
}

func (i *Instance) ptyInstancesPath() string {
	return path.Join(appsPath, i.AppName, procsPath, i.ProcessName, instancesPath, i.RevisionName, i.idString())
}

func (i *Instance) claimPath(host string) string {
	return i.dir.prefix("claims", host)
}

func (i *Instance) claimDir() *dir {
	return &dir{i.Snapshot, i.dir.prefix(claimsPath)}
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

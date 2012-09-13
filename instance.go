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

const (
	InsStatusInitial     InsStatus = "initial"
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

func newInstanceStub(app, rev, pty string, s Snapshot) (ins *Instance) {
	ins = &Instance{
		Id:           -1,
		AppName:      app,
		RevisionName: rev,
		ProcessName:  pty,
		Status:       InsStatusInitial,
		dir:          dir{s, "<invalid-path>"},
	}
	return
}

func newInstance(id int64, fields []string, status string, s Snapshot) (ins *Instance) {
	ins = newInstanceStub(fields[0], fields[1], fields[2], s)

	ins.Id = id
	ins.Status = InsStatus(status)
	ins.Ip = fields[3]
	port, err := strconv.Atoi(fields[4])
	if err != nil {
		panic("invalid port number: " + fields[4])
	}
	ins.Port = port
	ins.Host = fields[5]

	return
}

// GetInstance returns an Instance from the given app, rev, proc and instance ids.
func GetInstance(s Snapshot, id int64) (ins *Instance, err error) {
	p := fmt.Sprintf("%s/%d", instancesPath, id)

	status, _, err := s.get(p + "/status")
	if err != nil {
		return
	}

	f, err := s.getFile(p+"/object", new(listCodec))
	if err != nil {
		return
	}
	fields := f.Value.([]string)

	ins = newInstance(id, fields, status, s)

	return
}

func CreateInstance(app string, rev string, pty string, s Snapshot) (i *Instance, err error) {
	return newInstanceStub(app, rev, pty, s).Create()
}

func StopInstance(id string, s Snapshot) (s1 Snapshot, err error) {
	d := dir{s, path.Join(instancesPath, id)}
	rev, err := d.set("stop", "")
	if err != nil {
		return
	}
	s1 = s.FastForward(rev)

	return
}

// FastForward advances the ticket in time. It returns
// a new instance of Ticket with the supplied revision.
func (i *Instance) FastForward(rev int64) *Instance {
	return i.Snapshot.fastForward(i, rev).(*Instance)
}

func (i *Instance) createSnapshot(rev int64) snapshotable {
	tmp := *i
	tmp.Snapshot = Snapshot{rev, i.conn}
	return &tmp
}

func (i *Instance) Create() (i1 *Instance, err error) {
	i1 = i

	id, err := Getuid(i.Snapshot)
	if err != nil {
		return
	}
	i.Id = id
	i.dir.Name = fmt.Sprintf("%s/%d", instancesPath, i.Id)

	f, err := createFile(i.Snapshot, i.dir.prefix("object"), i.array(), new(listCodec))
	if err != nil {
		return
	}
	f, err = createFile(i.Snapshot, i.dir.prefix("start"), "", new(stringCodec))
	if err == nil {
		i1 = i.FastForward(f.FileRev)
	}
	return
}

// Claims returns the list of claimers
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

// Claim locks the Ticket to the specified host.
func (i *Instance) Claim(host string) (*Instance, error) {
	exists, rev, err := i.Snapshot.exists(i.dir.prefix("start"))
	if !exists {
		return i, fmt.Errorf("instance start field missing")
	}
	d := i.dir.fastForward(rev)

	_, err = d.set("start", host)
	if err != nil {
		return i, err
	}

	rev, err = i.claimDir().fastForward(rev).set(host, time.Now().UTC().String())
	if err != nil {
		return i, err
	}
	return i.FastForward(rev), err
}

func (i *Instance) Exit() (err error) {
	exists, _, err := i.Snapshot.exists(i.dir.prefix("stop"))
	if !exists {
		return fmt.Errorf("instance stop field missing")
	}
	_, err = i.updateStatus(InsStatusExited)
	if err != nil {
		return
	}
	err = i.Snapshot.del(i.ptyInstancesPath())
	if err != nil {
		return
	}
	return
}

func (i *Instance) Fail() (i1 *Instance, err error) {
	return i.updateStatus(InsStatusFailed)
}

func (i *Instance) Start(ip string, port int, host string) (i1 *Instance, err error) {
	_, err = i.updateStatus(InsStatusStarted)
	if err != nil {
		return
	}
	*i1 = *i
	i1.Ip = ip
	i1.Port = port
	i1.Host = host

	_, err = createFile(i.Snapshot, i.dir.prefix("object"), i1.array(), new(listCodec))
	if err != nil {
		return
	}
	s, err := i.Snapshot.set(i.ptyInstancesPath(), time.Now().UTC().String())
	if err != nil {
		return
	}
	i1 = i1.FastForward(s.Rev)

	return
}

func (i *Instance) updateStatus(s InsStatus) (i1 *Instance, err error) {
	rev, err := i.set("state", string(s))
	if err != nil {
		return
	}
	i.Status = s

	return i.FastForward(rev), err
}

// Unclaim removes the lock applied by Claim of the Ticket.
func (i *Instance) Unclaim(host string) (i1 *Instance, err error) {
	rev, err := i.set("start", "")
	if err != nil {
		return i, err
	}
	i1 = i.FastForward(rev)

	return
}

func (i *Instance) Unclaimable(host string, reason error) (i1 *Instance, err error) {
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

func WatchTicket(s Snapshot, listener chan *Instance, errors chan error) {
	rev := s.Rev

	for {
		ev, err := s.conn.Wait(path.Join(instancesPath, "*", "status"), rev+1)
		if err != nil {
			errors <- err
			return
		}
		rev = ev.Rev

		if !ev.IsSet() || string(ev.Body) != "unclaimed" {
			continue
		}

		ticket, err := parseTicket(s.FastForward(rev), &ev, ev.Body)
		if err != nil {
			continue
		}
		listener <- ticket
	}
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

func parseTicket(snapshot Snapshot, ev *doozer.Event, body []byte) (t *Instance, err error) {
	idStr := strings.Split(ev.Path, "/")[2]
	id, err := strconv.ParseInt(idStr, 0, 64)
	if err != nil {
		return nil, fmt.Errorf("ticket id %s can't be parsed as an int64", idStr)
	}

	p := path.Join(instancesPath, idStr)

	f, err := snapshot.getFile(path.Join(p, "op"), new(listCodec))
	if err != nil {
		return t, err
	}
	data := f.Value.([]string)

	t = &Instance{
		Id:           id,
		AppName:      data[0],
		RevisionName: data[1],
		ProcessName:  data[2],
		dir:          dir{snapshot, p},
	}
	return t, err
}

func (i *Instance) idString() string {
	return fmt.Sprintf("%d", i.Id)
}

func (i *Instance) ServiceName() string {
	return fmt.Sprintf("%s-%s", i.AppName, i.ProcessName)
}

func (i *Instance) ptyDeathsPath() string {
	return path.Join(appsPath, i.AppName, procsPath, i.ProcessName, deathsPath, i.idString())
}

func (i *Instance) ptyInstancesPath() string {
	return path.Join(appsPath, i.AppName, procsPath, i.ProcessName, instancesPath, i.idString())
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

func (i *Instance) array() []string {
	return []string{i.AppName, i.RevisionName, i.ProcessName, i.Ip, i.portString(), i.Host}
}

func (i *Instance) portString() string {
	return fmt.Sprintf("%d", i.Port)
}

// String returns the Go-syntax representation of Ticket.
func (i *Instance) String() string {
	return fmt.Sprintf("Instance{id=%d, app=%s, rev=%s, proc=%s, addr=%s:%d}", i.Id, i.AppName, i.RevisionName, i.ProcessName, i.Ip, i.Port)
}

// IdString returns a string of the format "TICKET[$ticket-id]"
func (i *Instance) IdString() string {
	return fmt.Sprintf("INSTANCE[%d]", i.Id)
}

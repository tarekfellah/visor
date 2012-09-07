package test_helper

import (
	"fmt"
	"github.com/soundcloud/visor"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

const PID_DIR = "/tmp/bazooka/pids"

var ticketId int64 = 10
var appNames = []string{"cat", "dog", "bird", "wolf", "bear", "lion", "tiger"}
var revNames = []string{"master", "slave", "e7fa81", "a91748", "f7ea91", "dev", "stable"}
var ptyNames = []string{"web", "db", "worker", "clock", "pusher", "dealer", "handoff"}

func ResetCoordinator(snapshot visor.Snapshot) {
	err := snapshot.ResetCoordinator()
	if err == nil {
		fmt.Println("test: reset coordinator tree.")
	} else {
		fmt.Printf("test: error reseting coordinator @%d: %s\n", snapshot.Rev, err)
	}
}

func WaitForEvent(c chan interface{}, evs string, timeout time.Duration) (ev interface{}, err error) {
	for {
		select {
		case ev = <-c:
			if reflect.TypeOf(ev).String() != evs {
				continue
			}
			return
		case <-time.After(timeout):
			err = fmt.Errorf("waiting for %s timed out", evs)
			return
		}
	}
	return
}

func ExpectEvent(c chan interface{}, evs string, timeout time.Duration, t *testing.T) (ev interface{}, err error) {
	ev, err = WaitForEvent(c, evs, timeout)
	if err != nil {
		t.Error(err)
	}
	return
}

func App(s visor.Snapshot) (app *visor.App) {
	name := randItem(appNames)
	app = visor.NewApp(name, "git://"+name+".git", visor.Stack("my-stack"), s)
	return
}

func Revision(app *visor.App, s visor.Snapshot) (rev *visor.Revision) {
	if app == nil {
		app = App(s)
	}
	name := randItem(revNames)
	rev = visor.NewRevision(app, name, s)
	return
}

func ProcType(app *visor.App, s visor.Snapshot, name string) (pty *visor.ProcType) {
	if app == nil {
		app = App(s)
	}
	pty = visor.NewProcType(app, visor.ProcessName(name), s)
	return
}

func Instance(pty *visor.ProcType, rev *visor.Revision, s visor.Snapshot) (ins *visor.Instance) {
	if pty == nil {
		pty = ProcType(nil, s, randItem(ptyNames))
	}
	if rev == nil {
		rev = Revision(nil, s)
	}
	addr := fmt.Sprintf("127.0.0.1:%d", 8000+rand.Int63n(1000))
	ins, err := visor.NewInstance(string(pty.Name), rev.Ref, rev.App.Name, addr, s)
	if err != nil {
		panic(err)
	}
	return
}

func randItem(items []string) string {
	return items[rand.Int63n(int64(len(items)))]
}

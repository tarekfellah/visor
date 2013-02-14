// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
)

func visorSetup(root string) (s Snapshot) {
	s, err := Dial(DefaultAddr, root)
	if err != nil {
		panic(err)
	}
	s.del("/")
	s = s.FastForward(-1)

	rev, err := Init(s)
	if err != nil {
		panic(err)
	}
	s = s.FastForward(rev)

	return
}

func TestDialWithDefaultAddrAndRoot(t *testing.T) {
	_, err := Dial(DefaultAddr, DefaultRoot)
	if err != nil {
		t.Error(err)
	}
}

func TestDialWithInvalidAddr(t *testing.T) {
	_, err := Dial("foo.bar:123:876", "wrong")
	if err == nil {
		t.Error("Dialed with invalid addr")
	}
}

func TestScaleErrors(t *testing.T) {
	s := visorSetup("/scale-error-test")
	scale := 5

	app := genApp(s)
	rev := genRevision(app)
	pty := genProctype(app, "web")

	s = s.FastForward(-1)

	// Scale up

	_, _, err := Scale("fnord", rev.Ref, pty.Name, scale, s)
	if err == nil {
		t.Error("expected error (bad arguments)")
	}
	_, _, err = Scale(app.Name, "fnord", pty.Name, scale, s)
	if err == nil {
		t.Error("expected error (bad arguments)")
	}
	_, _, err = Scale(app.Name, rev.Ref, "fnord", scale, s)
	if err == nil {
		t.Error("expected error (bad arguments)")
	}
}

func TestScale(t *testing.T) {
	s := visorSetup("/scale-test")
	scale := 5

	app := genApp(s)
	rev := genRevision(app)
	pty := genProctype(app, "web")

	s = s.FastForward(-1)

	// Scale up

	_, current, err := Scale(app.Name, rev.Ref, pty.Name, scale, s)
	if current != 0 {
		t.Fatal("expected current scale to = 0")
	}
	if err != nil {
		t.Fatal(err)
	}
	s = s.FastForward(-1)

	scale1, _, err := s.GetScale(app.Name, rev.Ref, pty.Name)
	if err != nil {
		t.Fatal(err)
	}
	if scale1 != scale {
		t.Fatalf("expected %d instances, got %d", scale, scale1)
	}

	// Scale down

	_, _, err = Scale(app.Name, rev.Ref, pty.Name, -1, s)
	if err == nil {
		t.Error("expected error on a non-positive scaling factor")
	}

	_, current, err = Scale(app.Name, rev.Ref, pty.Name, 1, s)
	if current != 5 {
		t.Fatal("expected current scale to = 5")
	}
	if err != nil {
		t.Fatal(err)
	}
	s = s.FastForward(-1)

	scale1, _, err = s.GetScale(app.Name, rev.Ref, pty.Name)
	if scale1 != scale {
		t.Fatalf("expected %d instances, got %d", scale, scale1)
	}

	_, _, err = Scale(app.Name, rev.Ref, pty.Name, 0, s)
	if err != nil {
		t.Fatal(err)
	}
	s = s.FastForward(-1)

	scale1, _, err = s.GetScale(app.Name, rev.Ref, pty.Name)
	if err != nil {
		t.Fatal(err)
	}
	if scale1 != scale {
		t.Fatalf("expected %d instances, got %d", scale, scale1)
	}
}

func TestGetuid(t *testing.T) {
	s, err := Dial(DefaultAddr, "/scale-test")
	if err != nil {
		panic(err)
	}
	uids := map[int64]bool{}
	ch := make(chan bool)

	for i := 0; i < 30; i++ {
		go func(i int) {
			uid, err := Getuid(s)
			if err != nil {
				t.Error(err)
			}
			if uids[uid] {
				t.Error("duplicate UID")
			}
			uids[uid] = true
			ch <- true
		}(i)
	}
	for i := 0; i < 30; i++ {
		<-ch
	}
}

func TestLock(t *testing.T) {
	s, err := Dial(DefaultAddr, "/lock-test")
	if err != nil {
		panic(err)
	}

	s1, err := Lock("my-lock", "secret", s)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Lock("my-lock", "secret", s)
	if err != ErrLocked {
		t.Fatal("expected lock to be taken")
	}

	err = Unlock("my-lock", s)
	if err == nil {
		t.Fatal("expected lock to be taken")
	}

	err = Unlock("my-lock", s1)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Lock("my-lock", "secret", s1)
	if err != nil {
		t.Fatal(err)
	}
}

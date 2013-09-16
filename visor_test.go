// Copyright (c) 2013, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package visor

import (
	"testing"
)

func visorSetup(root string) *Store {
	s, err := DialUri(DefaultUri, root)
	if err != nil {
		panic(err)
	}
	err = s.reset()
	if err != nil {
		panic(err)
	}
	s, err = s.FastForward()
	if err != nil {
		panic(err)
	}
	s, err = s.Init()
	if err != nil {
		panic(err)
	}

	return s
}

func TestScaleErrors(t *testing.T) {
	s := visorSetup("/scale-error-test")
	scale := 5

	app := genApp(s)
	rev := genRevision(app)
	pty := genProctype(app, "web")

	// Scale up

	_, _, err := s.Scale("fnord", rev.Ref, pty.Name, scale)
	if err == nil {
		t.Error("expected error (bad arguments)")
	}
	_, _, err = s.Scale(app.Name, "fnord", pty.Name, scale)
	if err == nil {
		t.Error("expected error (bad arguments)")
	}
	_, _, err = s.Scale(app.Name, rev.Ref, "fnord", scale)
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

	// Scale up
	ins, current, err := s.Scale(app.Name, rev.Ref, pty.Name, scale)
	if current != 0 {
		t.Fatal("expected current scale to = 0")
	}
	if err != nil {
		t.Fatal(err)
	}

	scale1, _, err := s.GetScale(app.Name, rev.Ref, pty.Name)
	if err != nil {
		t.Fatal(err)
	}
	if scale1 != scale {
		t.Fatalf("expected %d instances, got %d", scale, scale1)
	}

	// Scale down
	err = setInstancesToStarted(ins)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = s.Scale(app.Name, rev.Ref, pty.Name, -1)
	if err == nil {
		t.Error("expected error on a non-positive scaling factor")
	}

	_, _, err = s.Scale(app.Name, rev.Ref, pty.Name, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = s.Scale(app.Name, rev.Ref, pty.Name, 0)
	if err == nil {
		t.Fatal("expected error when scaling down instances twice")
	}

	scale1, _, err = s.GetScale(app.Name, rev.Ref, pty.Name)
	if err != nil {
		t.Fatal(err)
	}
	if scale1 != scale {
		t.Fatalf("expected %d instances, got %d", scale, scale1)
	}
}

func TestGetScale(t *testing.T) {
	s := visorSetup("/getscale-test")
	scale := 5

	app := s.NewApp("scale-app", "git://scale.git", "scale-stack")
	pty := s.NewProcType(app, "scaleproc")
	rev := s.NewRevision(app, "scale-ref")

	_, err := app.Register()
	if err != nil {
		t.Fatal(err)
	}
	_, err = rev.Register()
	if err != nil {
		t.Fatal(err)
	}
	_, err = pty.Register()
	if err != nil {
		t.Fatal(err)
	}
	s, err = s.FastForward()
	if err != nil {
		t.Fatal(err)
	}

	scale, _, err = s.GetScale(app.Name, rev.Ref, string(pty.Name))
	if err != nil {
		t.Error(err)
	}
	if scale != 0 {
		t.Error("expected initial scale of 0")
	}

	_, _, err = s.Scale(app.Name, rev.Ref, pty.Name, 9)
	if err != nil {
		t.Fatal(err)
	}
	s, err = s.FastForward()
	if err != nil {
		t.Fatal(err)
	}

	scale, _, err = s.GetScale(app.Name, rev.Ref, string(pty.Name))
	if err != nil {
		t.Error(err)
	}
	if scale != 9 {
		t.Errorf("expected scale of 9, got %d", scale)
	}

	scale, _, err = s.GetScale("invalid-app", rev.Ref, string(pty.Name))
	if scale != 0 {
		t.Errorf("expected scale to be 0")
	}
}

func TestValidateInput(t *testing.T) {
	err := validateInput("")
	if err == nil || !IsErrInvalidArgument(err) {
		t.Errorf("expected zero length string to not validate")
	}
	err = validateInput(" ")
	if err == nil || !IsErrInvalidArgument(err) {
		t.Errorf("expected empty string to not validate")
	}
	err = validateInput("app name")
	if err == nil || !IsErrInvalidArgument(err) {
		t.Errorf("expected whitespace in input to not validate")
	}
	err = validateInput("underscore_proc")
	if err == nil || !IsErrInvalidArgument(err) {
		t.Errorf("expected underscore in input to not validate")
	}
	err = validateInput("valid-name")
	if err != nil {
		t.Errorf("expected input '%s' to validate: %s", "valid-name", err)
	}
}

func setInstancesToStarted(ins []*Instance) error {
	host := "127.0.0.1"
	hostname := "localhost"
	port := 5000

	for _, i := range ins {
		i, err := i.Claim(host)
		if err != nil {
			return err
		}
		i, err = i.Started(host, port, hostname)
		if err != nil {
			return err
		}
		port += 1
	}
	return nil
}

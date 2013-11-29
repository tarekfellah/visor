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
	proc := genProctype(app, "web")
	env := genEnv(app, "default", map[string]string{})

	// Scale up

	_, _, err := s.Scale("fnord", rev.Ref, proc.Name, env.Ref, scale)
	if err == nil {
		t.Error("expected error (bad arguments)")
	}
	_, _, err = s.Scale(app.Name, "fnord", proc.Name, env.Ref, scale)
	if err == nil {
		t.Error("expected error (bad arguments)")
	}
	_, _, err = s.Scale(app.Name, rev.Ref, "fnord", env.Ref, scale)
	if err == nil {
		t.Error("expected error (bad arguments)")
	}
}

func TestScale(t *testing.T) {
	s := visorSetup("/scale-test")
	scale := 5

	app := genApp(s)
	rev := genRevision(app)
	proc := genProctype(app, "web")
	env := genEnv(app, "default", map[string]string{})

	// Scale up
	ins, current, err := s.Scale(app.Name, rev.Ref, proc.Name, env.Ref, scale)
	if err != nil {
		t.Fatal(err)
	}
	if current != 0 {
		t.Fatalf("expected current scale to = 0, but is %d", current)
	}
	if err != nil {
		t.Fatal(err)
	}

	scale1, _, err := s.GetScale(app.Name, rev.Ref, proc.Name)
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

	_, _, err = s.Scale(app.Name, rev.Ref, proc.Name, "default", -1)
	if err == nil {
		t.Error("expected error on a non-positive scaling factor")
	}

	_, _, err = s.Scale(app.Name, rev.Ref, proc.Name, env.Ref, 0)
	if err != nil {
		t.Fatal(err)
	}

	_, _, err = s.Scale(app.Name, rev.Ref, proc.Name, env.Ref, 0)
	if err == nil {
		t.Fatal("expected error when scaling down instances twice")
	}

	scale1, _, err = s.GetScale(app.Name, rev.Ref, proc.Name)
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

	app := genApp(s)
	rev := genRevision(app)
	proc := genProctype(app, "scaleproc")
	env := genEnv(app, "default", map[string]string{})

	scale, _, err := s.GetScale(app.Name, rev.Ref, proc.Name)
	if err != nil {
		t.Error(err)
	}
	if scale != 0 {
		t.Error("expected initial scale of 0")
	}

	_, _, err = s.Scale(app.Name, rev.Ref, proc.Name, env.Ref, 9)
	if err != nil {
		t.Fatal(err)
	}
	s, err = s.FastForward()
	if err != nil {
		t.Fatal(err)
	}

	scale, _, err = s.GetScale(app.Name, rev.Ref, proc.Name)
	if err != nil {
		t.Error(err)
	}
	if scale != 9 {
		t.Errorf("expected scale of 9, got %d", scale)
	}

	scale, _, err = s.GetScale("invalid-app", rev.Ref, proc.Name)
	if scale != 0 {
		t.Errorf("expected scale to be 0")
	}
}

func TestValidateInput(t *testing.T) {
	invalidInputs := []string{
		"",
		" ",
		"with space",
		"with_underscore",
	}
	for _, input := range invalidInputs {
		err := validateInput(input)
		if err == nil || !IsErrInvalidArgument(err) {
			t.Errorf("expected '%s' to not validate: %s", input, err)
		}
	}

	validInputs := []string{
		"valid",
		"valid-with-scores",
		"valid.with.dots",
		"0123456789",
	}
	for _, input := range validInputs {
		err := validateInput(input)
		if err != nil {
			t.Errorf("expected '%s' to validate: %s", input, err)
		}
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

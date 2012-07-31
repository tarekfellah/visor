// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	"flag"
	"fmt"
	"github.com/soundcloud/visor"
	"log"
	"os"
	"strings"
	"text/template"
)

const VERSION_STRING = "0.4.2"

type Command struct {
	Run       func(cmd *Command, args []string)
	Flag      flag.FlagSet
	Long      string
	Name      string
	UsageLine string
	Short     string
	Snapshot  visor.Snapshot
}

func (c *Command) Usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s\n\n", c.UsageLine)
	fmt.Fprintf(os.Stderr, "%s\n", strings.TrimSpace(c.Long))
	os.Exit(2)
}

var Uri string
var Root string
var Version bool

func init() {
	flag.StringVar(&Uri, "uri", visor.DEFAULT_URI, "doozer uri")
	flag.StringVar(&Root, "root", visor.DEFAULT_ROOT, "doozer root")
	flag.BoolVar(&Version, "version", false, "print version and exit")
}

var commands = []*Command{
	cmdAppDescribe,
	cmdAppEnvDel,
	cmdAppEnvGet,
	cmdAppEnvSet,
	cmdAppInstances,
	cmdAppInstancesPurge,
	cmdAppList,
	cmdAppRegister,
	cmdAppServices,
	cmdAppUnregister,
	cmdInit,
	cmdProcRegister,
	cmdProcUnregister,
	cmdRevDescribe,
	cmdRevExists,
	cmdRevRegister,
	cmdRevUnregister,
	cmdScale,
}

func main() {
	flag.Usage = usage
	flag.Parse()
	log.SetFlags(0)

	args := flag.Args()

	if Version == true {
		fmt.Println(VERSION_STRING)
		os.Exit(0)
	}

	if len(args) < 1 {
		usage()
	}

	for _, cmd := range commands {
		if cmd.Name == args[0] && cmd.Run != nil {
			s, err := visor.DialUri(Uri, Root)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error connection %s\n", err.Error())
				os.Exit(2)
			}

			cmd.Snapshot = s
			cmd.Flag.Usage = func() { cmd.Usage() }
			cmd.Flag.Parse(args[1:])
			cmd.Run(cmd, cmd.Flag.Args())
			return
		}
	}

	fmt.Fprintf(os.Stderr, "Unknown command %#q\n\n", args[0])
	usage()
}

func usage() {
	t := template.New("top")
	t.Funcs(template.FuncMap{"trim": strings.TrimSpace})
	template.Must(t.Parse(usageTmpl))
	if err := t.Execute(os.Stderr, commands); err != nil {
		panic(err)
	}

	os.Exit(2)
}

var usageTmpl = `Usage: visor [globals] command [arguments]

Globals:
  -root     Doozerd tree prefix
  -uri      Doozerd cluster URI
  -version  Print version and exit

Commands:{{range .}}
  {{.Name | printf "%-15s"}} {{.Short}}{{end}}
`

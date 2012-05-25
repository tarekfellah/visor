// Copyright (c) 2012, SoundCloud Ltd.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
// Source code and contact info at http://github.com/soundcloud/visor

package main

import (
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"github.com/soundcloud/visor"
	"os"
)

var snapshot func() (s visor.Snapshot)

const VERSION_STRING = "v0.1.1"

func main() {
	instanceSubCommands := getopt.SubCommands{
		"exists": {
			"returns 0 if instance exists something != 0 otherwise",
			getopt.Definitions{
				{"instanceid", "id of the instance of interest", getopt.IsArg | getopt.Required, ""},
			},
		},
		"describe": {
			"describe instance",
			getopt.Definitions{
				{"instanceid", "id of the instance of interest", getopt.IsArg | getopt.Required, ""},
			},
		},
		"tail": {
			"tail instance stdout / stderr",
			getopt.Definitions{
				{"instanceid", "id of the instance of interest", getopt.IsArg | getopt.Required, ""},
			},
		},
		"kill": {
			"kill an instance",
			getopt.Definitions{
				{"instanceid", "id of the instance of interest", getopt.IsArg | getopt.Required, ""},
				{"signal|s", "signal to send", getopt.Optional, "SIGKILL"},
			},
		},
	}

	if os.Getenv("VISOR_DEBUG") != "" {
		instanceSubCommands["create"] = getopt.Options{
			"[DEBUG] create a new instance entry",
			getopt.Definitions{
				{"app", "app", getopt.IsArg | getopt.Required, ""},
				{"revision", "revision", getopt.IsArg | getopt.Required, ""},
				{"proctype", "proctype", getopt.IsArg | getopt.Required, ""},
				{"addr", "tcp address in the form of <host>:<port>", getopt.IsArg | getopt.Required, "10.20.30.40:20000"},
			},
		}
	}

	ssco := getopt.SubSubCommandOptions{
		getopt.Options{
			"A cli interface to visor (see http://github.com/soundcloud/visor)",
			getopt.Definitions{
				{"config|c|CONFIG", "config file", getopt.IsConfigFile | getopt.ExampleIsDefault, "/etc/visor.conf"},
				{"doozerd|d|DOOZERD_HOST", "doozer server", getopt.Optional | getopt.ExampleIsDefault, "127.0.0.1"},
				{"port|p|DOOZERD_PORT", "doozer server port", getopt.Optional | getopt.ExampleIsDefault, "8046"},
				{"root|r|VISOR_ROOT", "namespacing for visor: all entries to the coordinator will be namespaced to this dir", getopt.Optional | getopt.ExampleIsDefault, visor.DEFAULT_ROOT},
				{"version|v", "print version of the current visor", getopt.Optional | getopt.Flag, false},
				{"scope", "scope to operate on", getopt.IsSubCommand, ""},
			},
		},
		getopt.Scopes{
			"root": {
				getopt.Options{
					"Everything that has to do root and setup",
					getopt.Definitions{
						{"command", "command to execute", getopt.IsSubCommand, ""},
					},
				},
				getopt.SubCommands{
					"init": {
						"Initialize coordinator state",
						getopt.Definitions{},
					},
				},
			},
			"app": {
				getopt.Options{
					"Everything that has to do with apps",
					getopt.Definitions{
						{"command", "command to execute", getopt.IsSubCommand, ""},
					},
				},
				getopt.SubCommands{
					"exists": {
						"returns 0 if app exists something != 0 otherwise",
						getopt.Definitions{
							{"name", "name of the new app", getopt.IsArg | getopt.Required, ""},
						},
					},
					"list": {
						"list available applications",
						getopt.Definitions{},
					},
					"describe": {
						"show information about the app",
						getopt.Definitions{
							{"name", "name of the new app", getopt.IsArg | getopt.Required, ""},
						},
					},
					"setenv": {
						"sets an environment variable for this application",
						getopt.Definitions{
							{"name", "name of the app", getopt.IsArg | getopt.Required, ""},
							{"key", "key (name) of the env variable", getopt.IsArg | getopt.Required, ""},
							{"value", "value of the env variable (omit to delete an env var)", getopt.IsArg | getopt.Optional, ""},
						},
					},
					"getenv": {
						"gets an environment variable for this application",
						getopt.Definitions{
							{"name", "name of the app", getopt.IsArg | getopt.Required, ""},
							{"key", "key (name) of the env variable", getopt.IsArg | getopt.Required, ""},
						},
					},
					"unregister": {
						"unregister / delete a new application, and all it's revisions and stop all running instances",
						getopt.Definitions{
							{"name", "name of the new app", getopt.IsArg | getopt.Required, ""},
						},
					},
					"register": {
						"register a new application with bazooka",
						getopt.Definitions{
							{"type|t", "deploy type of the application (lxc, mount or bazapta)", getopt.Optional | getopt.ExampleIsDefault, "lxc"},
							{"repourl|u", "url to the repository of this app", getopt.Required, "http://github.com/soundcloud/<your_project>"},
							{"stack|s", "stack version ... should usually be HEAD", getopt.Optional | getopt.ExampleIsDefault, "HEAD"},
							{"name", "name of the new app", getopt.IsArg | getopt.Required, ""},
						},
					},
					"env": {
						"show environment of an application",
						getopt.Definitions{
							{"name", "name of the new app", getopt.IsArg | getopt.Required, ""},
						},
					},
					"revisions": {
						"show available revisions of an app",
						getopt.Definitions{
							{"name", "name of the new app", getopt.IsArg | getopt.Required, ""},
						},
					},
				},
			},
			"ticket": {
				getopt.Options{
					"Show and list tickets",
					getopt.Definitions{
						{"command", "command to execute", getopt.IsSubCommand, ""},
					},
				},
				getopt.SubCommands{
					"create": {
						"create new ticket",
						getopt.Definitions{
							{"app", "app", getopt.IsArg | getopt.Required, ""},
							{"revision", "revision", getopt.IsArg | getopt.Required, ""},
							{"proctype", "proctype", getopt.IsArg | getopt.Required, ""},
							{"operation", "operation (start|stop)", getopt.IsArg | getopt.Required, ""},
						},
					},
				},
			},
			"revision": {
				getopt.Options{
					"Everything that has to do with revisions",
					getopt.Definitions{
						{"command", "command to execute", getopt.IsSubCommand, ""},
					},
				},
				getopt.SubCommands{
					"exists": {
						"returns 0 if revision exists something != 0 otherwise",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, ""},
							{"revision", "revision to use", getopt.IsArg | getopt.Required, "HEAD"},
						},
					},
					"describe": {
						"describe revision of an app",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, ""},
							{"revision", "revision to use", getopt.IsArg | getopt.Optional | getopt.ExampleIsDefault, "HEAD"},
						},
					},
					"register": {
						"register an app-revision",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, ""},
							{"revision", "revision to use", getopt.IsArg | getopt.Required, ""},
							{"artifacturl|u", "url to the deployed artifact", getopt.Required, ""},
							{"proctypes|t", "list of proctypes", getopt.Optional, []string{"web", "worker"}},
						},
					},
					"unregister": {
						"unregister an app-revision",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, ""},
							{"revision", "revision to use", getopt.IsArg | getopt.Required, ""},
						},
					},
					"instances": {
						"list all instances of an app revision",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, "myapp"},
							{"revision", "revision to use", getopt.IsArg | getopt.Required, "34f3457"},
						},
					},
					"scale": {
						"scales an app revisions proctype by the given factor",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, "myapp"},
							{"revision", "revision to use", getopt.IsArg | getopt.Required, "34f3457"},
							{"proctype", "proctype to scale", getopt.IsArg | getopt.Required, "web"},
							{"factor", "scaling factor", getopt.IsArg | getopt.Required, 5},
						},
					},
				},
			},
			"instance": {
				getopt.Options{
					"Everything that has to do with instances",
					getopt.Definitions{
						{"command", "command to execute", getopt.IsSubCommand, ""},
					},
				},
				instanceSubCommands,
			},
		},
	}

	scope, subCommand, options, arguments, passThrough, e := ssco.ParseCommandLine()

	help, wantsHelp := options["help"]

	if e != nil || wantsHelp {
		exit_code := 0

		switch {
		case wantsHelp && help.String == "usage":
			fmt.Print(ssco.Usage())
		case wantsHelp && help.String == "help":
			fmt.Print(ssco.Help())
		case options["version"].Bool == true:
			fmt.Println(VERSION_STRING)
		default:
			fmt.Printf("\n**** Error: %s\n\n%s", e.Error(), ssco.Help())
			if e.ErrorCode != getopt.MissingArgument {
				if subCommand != "" && e.ErrorCode != getopt.UnknownSubCommand {
					fmt.Printf("**** See as well the help for the scope command by doing a\n\t%s %s --help\n\n", os.Args[0], scope)
				}
				if scope != "*" && e.ErrorCode != getopt.UnknownScope {
					fmt.Printf("**** See as well the help for the global command by doing a\n\t%s --help\n\n", os.Args[0])
				}
			}
			exit_code = e.ErrorCode
		}
		os.Exit(exit_code)
	}

	snapshot = func() (s visor.Snapshot) {
		root := options["root"].String
		doozerd := options["doozerd"].String + ":" + options["port"].String
		s, err := visor.Dial(doozerd, root)

		if err != nil {
			fmt.Fprintf(os.Stderr, "**** Error: %s\n", err.Error())
			os.Exit(2)
		}

		return s
	}

	var err error
	switch scope {
	case "root":
		err = Root(subCommand, options, arguments, passThrough)
	case "app":
		err = App(subCommand, options, arguments, passThrough)
	case "revision":
		err = Revision(subCommand, options, arguments, passThrough)
	case "instance":
		err = Instance(subCommand, options, arguments, passThrough)
	case "ticket":
		err = Ticket(subCommand, options, arguments, passThrough)
	default:

		fmt.Println("no fucking way did this happen!")
	}

	if err != nil {
		fmt.Printf("**** Error: " + err.Error() + "\n")
		os.Exit(1)
	}

}

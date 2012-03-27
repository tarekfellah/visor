package main

import (
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"os"
)

func main() {
	ssco := getopt.SubSubCommandOptions{
		getopt.Options{
			"A cli interface to visor (see http://github.com/soundcloud/visor)",
			getopt.Definitions{
				{"config|c|CONFIG", "config file", getopt.IsConfigFile | getopt.ExampleIsDefault, "/etc/visor.conf"},
				{"doozer|d|DOOZERD_HOST", "doozer server", getopt.Required, ""},
				{"port|p|DOOZERD_PORT", "doozer server port", getopt.Required | getopt.ExampleIsDefault, "8046"},
				{"root|r|ROOT", "namespacing for visor: all entries to the coordinator will be namespaced to this dir", getopt.Required | getopt.ExampleIsDefault, "/bazooka"},
				{"scope", "scope to operate on", getopt.IsSubCommand, ""},
			},
		},
		getopt.Scopes{
			"app": {
				getopt.Options{
					"Everything that has to do with apps",
					getopt.Definitions{
						{"command", "command to execute", getopt.IsSubCommand, ""},
					},
				},
				getopt.SubCommands{
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
					"register": {
						"register a new application with bazooka",
						getopt.Definitions{
							{"type|t", "deploy type of the application (lxc, mount or bazapta)", getopt.Required | getopt.ExampleIsDefault, "lxc"},
							{"repourl|r", "url to the repository of this app", getopt.Optional, "http://github.com/soundcloud/<your_project>"},
							{"stack|s", "stack version ... should usually be HEAD", getopt.Required | getopt.ExampleIsDefault, "HEAD"},
							{"irc|i|", "comma separated list of irc channels where to announce new deployments", getopt.Optional, []string{"#deploys"}},
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
			"revision": {
				getopt.Options{
					"everything that has to do with revisions",
					getopt.Definitions{},
				},
				getopt.SubCommands{
					"describe": {
						"describe revision of an app",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, ""},
							{"revision", "revision to use", getopt.IsArg | getopt.Required | getopt.ExampleIsDefault, "HEAD"},
						},
					},
					"unregister": {
						"unregister an app-revision",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, ""},
							{"revision", "revision to use", getopt.IsArg | getopt.Required, ""},
						},
					},
					"scale": {
						"scale app-revision-proc_type",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, "myapp"},
							{"revision", "revision to use", getopt.IsArg | getopt.Required, "34f3457"},
							{"proc", "proc type", getopt.IsArg | getopt.Required, "web"},
							{"num", "scaling factor", getopt.IsArg | getopt.Required, ""},
						},
					},
					"instances": {
						"list all instances of an app revision",
						getopt.Definitions{
							{"app", "name of the app", getopt.IsArg | getopt.Required, "myapp"},
							{"revision", "revision to use", getopt.IsArg | getopt.Required, "34f3457"},
						},
					},
				},
			},
			"instance": {
				getopt.Options{
					"everything that has to do with instances",
					getopt.Definitions{},
				},
				getopt.SubCommands{
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
				},
			},
		},
	}

	scope, subCommand, options, arguments, passThrough, e := ssco.ParseCommandLine()

	help, wantsHelp := options["help"]

	fmt.Printf("scope: %s\n", scope)
	fmt.Printf("subCommand: %s\n", subCommand)
	fmt.Printf("options:%#v\n", options)
	fmt.Printf("arguments: %#v\n", arguments)
	fmt.Printf("passThrough: %#v\n", passThrough)

	if e != nil || wantsHelp {
		exit_code := 0

		switch {
		case wantsHelp && help.String == "usage":
			fmt.Print(ssco.Usage())
		case wantsHelp && help.String == "help":
			fmt.Print(ssco.Help())
		default:
			fmt.Println("**** Error: ", e.Error(), "\n", ssco.Help())
			exit_code = e.ErrorCode
		}
		os.Exit(exit_code)
	}

}

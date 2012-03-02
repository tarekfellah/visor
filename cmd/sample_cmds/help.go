package main

import (
	"fmt"
	getopt "github.com/kesselborn/go-getopt"
	"os"
)

func main() {
	//  visor [-c config] -s <server> [-p <port>] [-r <root>]
	optionDefinition := getopt.Options{
		{"scope", "Show commands for one of these scopes: app, revision, instance", getopt.IsArg | getopt.Required, ""},
		{"command", `Show help for a specific command; available scopes and their commands:

App:
  app list       
  app describe   
  app register   
  app unregister 
  app setenv     
  app getenv     
  app env        
    

Instance:
  instance add         
  instance describe    
  instance list        
  instance register    
  instance unregister  

Revision:
  revision add         
  revision describe    
  revision list        
  revision register    
  revision unregister  
  revision scale       
  revision instances   
  revision reginst     
  revision unreginst   

Ticket:
  ticket create
    
    `, getopt.IsArg | getopt.Optional, ""},
	}

	_, _, _, e := optionDefinition.ParseCommandLine()

	os.Args[0] = "visor help"

	if e != nil {
		exit_code := 0
		description := ""

		switch {
		case e.ErrorCode == getopt.WantsUsage:
			fmt.Print(optionDefinition.Usage())
		case e.ErrorCode == getopt.WantsHelp:
			fmt.Print(optionDefinition.Help(description))
		default:
			fmt.Println("**** Error: ", e.Message, "\n", optionDefinition.Help(description))
			exit_code = e.ErrorCode
		}
		os.Exit(exit_code)
	}
}

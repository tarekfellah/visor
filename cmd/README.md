Visor CLI
=========

General
------

The Visor CLI works with subcommands; The general call looks like this:

### General command layout:

    Usage: visor [-c <config>] -s <server> [-p <port>] [-r <root>] <scope> <command> [<...>]
    
    Options:
        -c, --config=<config>     config file (default: /etc/visor.conf)
        -s, --server=<server>     doozer server
        -p, --port=<port>         port the doozer server is listening on (default: 8046)
        -r, --root=<root>         visor namespace within doozer: all entries will be prepended with this path (default: /visor)
        -h, --help                usage (-h) / detailed help text (--help)
    
    Arguments:
        <scope>                   scope a command should be executed on; call 'visor help' for an overview
        <command>                 command to execute
        <...>                     command's arguments and options


### Help

    Usage: visor help <scope> [<command>]
    
    Arguments:
        <scope>                   Show commands for one of these scopes: app, revision, instance
        <command>                 Show help for a specific command; available scopes and their commands:
    
    App:
      app list       
      app describe <name>
      app register <name>
      app unregister <name>
      app setenv <name> <key> [<value>]
      app getenv <name> <key>
      app env <name>       
        
    
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

App Scope
---------

### Register an app:

    Usage: visor [opts] app register <name> [-t <deploytype>] [-r <repourl>] [-s <stack>] [-i <irc>]
    
    Options:
        -t, --deploytype=<deploytype>   deploy type (one of mount, bazapta, lxc) (default: lxc)
        -r, --repourl=<repourl>         repository url of this app (default: http://github.com/soundcloud/<name>)
        -s, --stack=<stack>             stack version this app should be pinned to -- ommit if you always want the latest stack
        -i, --irc=<irc>                 comma separated list of irc channels where a deploy should be announced (default: deploys)
        -h, --help                      usage (-h) / detailed help text (--help)
    
    Arguments:
        <name>                          app's name

### Unregister app:

    Usage: visor [opts] app register <name>
    
    deletes an app from the coordinator -- does not stop currently running instances.
    
    Arguments:
        <name>              app's name
        
### show app details

Usage: visor [opts] app describe <name>

Arguments:
    <name>              app's name

### list 

    Usage: visor [opts] app list
    
    Lists all available applications


### set env variable

    Usage: visor [opts] app setenv <appname> <key> [<value>]
    
    Sets an environment variable that will be set passed to the application when it's started
    
    Arguments:
        <name>                    app's name
        <key>                     environment variable's name
        <value>                   environment variable's value (omit in order to delete variable)

### get env variable
    
    Usage: visor [opts] app getenv <name> <key>
    
    Gets an app environment variable
    
    Arguments:
        <name>              app's name
        <key>               environment variable's name
    
### get env

    Usage: visor [opts] app env <name>
    
    Gets an application's environment
    
    Arguments:
        <name>              app's name


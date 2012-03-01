Visor CLI
=========

A CLI to visor
-------------

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

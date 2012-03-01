Visor CLI
=========

A CLI to visor
-------------

The Visor CLI works with subcommands; The general call looks like this:

### General command layout:

    Usage: visor [-c <config>] -s <server> [-p <port>] [-r <root>] <component> <command> [<...>]
    
    Options:
        -c, --config=<config>         config file (default: /etc/visor.conf)
        -s, --server=<server>         doozer server
        -p, --port=<port>             port the doozer server is listening on (default: 8046)
        -r, --root=<root>             visor namespace within doozer: all entries will be prepended with this path (default: /visor)
        -h, --help                    usage (-h) / detailed help text (--help)
    
    Arguments:
        <component>                   component scope a command should be executed on; call 'visor help' for an overview
        <command>                     command to execute
        <...>                         command's arguments and options

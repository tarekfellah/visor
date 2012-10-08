
the bazooka instance scaling system
===================================

bazooka-cli scales object from 0 to 1, this creates a ticket with an empty
*start* file, and an entry under the app/proc instances path.

        instances/
            5461/
      +         object = <app> <rev> <proc>
      +         start  =

        apps/<app>/procs/<proc>/instances/<rev>/
      +     5461 = 2012-07-19 16:28 UTC

a bazooka-pm claims the instance, by successfully setting the *start* file to
its address. It then adds itself to the *claims* dir.

        instances/
            5461/
                claims/
      +             10.0.1.24 = 2012-07-19 16:22 UTC
                object = <app> <rev> <proc>
      -         start  =
      +         start  = 10.0.1.24

the bazooka-pm fails to deploy the object, it clears the *start* file, which in
turn triggers a new event for the remaining bazooka-pms.

it may also set its claim key to the error message.

        instances/
            5461/
                claims/
      -             10.0.1.24 = 2012-07-19 16:28 UTC
      +             10.0.1.24 = 2012-07-19 16:28 UTC file 'bin/server' not found
                object = <app> <rev> <proc>
      -         start  = 10.0.1.24
      +         start  =

another bazooka-pm claims the ticket and adds itself to the *claims* dir.

        instances/
            5461/
                claims/
                    10.0.1.24 = 2012-07-19 16:28 UTC file 'bin/server' not found
      +             10.0.1.15 = 2012-07-19 16:41 UTC
                object = <app> <rev> <proc>
      -         start  =
      +         start  = 10.0.1.15

this bazooka-pm successfuly deploys the object, sets the *status* file to the
instance ip/port/hostname.

        instances/
            5461/
                claims/
                    10.0.1.24 = 2012-07-19 16:28 UTC file 'bin/server' not found
                    10.0.1.15 = 2012-07-19 16:41 UTC
                object = <app> <rev> <proc>
      -         start  = 10.0.1.15
      +         start  = 10.0.1.15 9090 instance.local
  
bazooka-cli asks for a scale from 1 to 0 this only succeeds for instances which
have already been started. An empty *stop* file is created.

        instances/
            5461/
                claims/
                    10.0.1.24 = 2012-07-19 16:28 UTC file 'bin/server' not found
                    10.0.1.15 = 2012-07-19 16:41 UTC
                object = <app> <rev> <proc>
                start  = 10.0.1.15 9090 instance.local
      +         stop   =

the bazooka-pm which owns the ticket, sets the *stop* file to its address.

        instances/
            5461/
                claims/
                    10.0.1.24 = 2012-07-19 16:28 UTC file 'bin/server' not found
                    10.0.1.15 = 2012-07-19 16:41 UTC
                object = <app> <rev> <proc>
                start  = 10.0.1.15 9090 instance.local
      -         stop   =
      +         stop   = 10.0.1.15 2012-07-19 16:41 UTC

if successful, the status is set to `exited`

        instances/
            5461/
                claims/
                    10.0.1.24 = 2012-07-19 16:28 UTC file 'bin/server' not found
                    10.0.1.15 = 2012-07-19 16:41 UTC
                object = <app> <rev> <proc>
                start  = 10.0.1.15 9090 instance.local
                stop   = 10.0.1.15 2012-07-19 16:41 UTC
      +         status = exited

and the instance entry is deleted from the app/pty tree

        instances/
            5461/
                claims/
                    10.0.1.24 = 2012-07-19 16:28 UTC file 'bin/server' not found
                    10.0.1.15 = 2012-07-19 16:41 UTC
                object = <app> <rev> <proc>
                start  = 10.0.1.15 9090 instance.local
                stop   = 10.0.1.15 2012-07-19 16:41 UTC
                status = exited

        apps/<app>/procs/<proc>/instances/<rev>/
      -     5461 = 2012-07-19 16:28 UTC


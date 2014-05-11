# Procmgr - A Simple Process Manager

`procmgr` is a process manager for managing running processes.
It is loosely based on process management of services like upstart and systemd, but is not meant to be a full init system.

## Config

`procmgr` reads a yaml config file with a description of various processes to run.

Example:

    processes:
    -
      name: hello
      command: echo 'Hello, World'
      streams: [stdout]

Given the above config file, `procmgr` will start the process `echo 'Hello, World'` when run.
The process will have it's stdout piped to the current stdout. Stdin and stderr are discarded.

### Dependencies

Processes can be chained in an event-driven system:

    processes:
    -
      name: database
      command: pg_ctl -D /var/lib/postgres/data/ start
    -
      name: webapp
      command: unicorn
      streams: [stdin, stdout, stderr]
      after: [database.finished]

Given the above config, the `database` process would be started, and only after the process has completed will the `webapp` process be started.


### Restarting

Processes can be restarted when they exit cleanly (0 exit code) by providing a `restart: true` field:

    processes:
    -
      # This process prints 'Hello' to stdout once every 2 seconds.
      # It is restarted by the service manager when it stops cleanly (0 exit code).
      # If it exists with non-zero exit code, it will not be restarted.
      name: greeter
      command: echo 'Hello' ; sleep 2
      streams: [stdout]
      restart: true 

### Example

    processes:
    -
      # Simple task.
      # Attaches stdout so that the echo will be printed to screen.
      name: hello
      command: echo 'Hello World'
      streams: [stdout]
    -
      # Sleep for 2 seconds to demonstrate the dependencies
      name: sleep
      command: sleep 2
      streams: []
      after: [hello.finished]
    -
      # Wait until the sleep finishes and say hello again
      name: helloagain
      command: echo 'Hello Again'
      streams: [stdout]
      after: [sleep.finished]
    -
      # Run at the same time as 'helloagain' and print goodbye to stderr
      # Attaches stderr so that the echo will be printed to the screen.
      name: goodbye
      command: echo 'Goodbye World' 1>&2
      streams: [stderr]
      after: [helloagain.started]

## Planned Features

- Changing the `cwd` of a process.
- Changing the `user`/`group` of a process.
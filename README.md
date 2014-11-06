# pvc-server

A server for handling pvc agents running on puppetmasters to report their statistics and
puppet agents to manage their runs.

# Building
``
$ export GOPATH="/path/to/buildir" ; mkdir -p $GOPATH
$ git clone https://github.com/johnj/pvc-server && cd pvc-server
$ go get ./..
$ go build github.com/johnj/pvc-server
``

# Running
``
$ pvc-server
``

# Database Setup
`pvc-server` relies on PostgreSQL (any version), the schema is located in the `schema.sql`
file.

``
$ createuser -U postgres pvc
$ createdb -U pvc pvc
$ psql -U pvc pvc < schema.sql
``

# Configuration
By default found @ /etc/pvc-server.conf, can be specified on the command line with `-f`.

# Processors (optional)
Processors are invoked for reports (post puppet run) and facts (also post puppet runs).

The executable scripts are specified via the `reportprocessor` and `factsprocessor`
settings. Each processor takes base64 encoded data via stdin (fd/0).

These settings can be left empty if you have no need to do any post-processing.

# Overview

xenvman is an extensible environment manager which can be used to
create environments for testing microservices.

![Overview](docs/img/overview.png)

xenvman can be used to:

* Define environment templates using JavaScript
* Create images on the fly
* Spawn as many containers as needed inside an environment
* Link containers together in a single isolated network
* Expose container ports for external access

For a detailed example take a look at [tutorial](docs/tutorial.md).

# Installation

## Download release

## Compilation from source

## Configuration

## Running API server

# Environments

Environment is an isolated bubble where one or more containers can be run 
together in order to provide a necessary playground for infrastructure testing.

Environments are created, managed and destroyed using HTTP API provided
by running `xenvman` server.

# Templates

An environment is set up by executing one or more templates,
where a template is a a small program written in JavaScript
which defines what images to build/fetch, what and how many containers
to spawn, what files to mount inside containers, what ports to expose etc.

A template script is run by embedded JS interpreter inside `xenvman` server.
One template is just one javascript file located withing a template base directory (defined by `tpl.base-dir` configuration parameter, or `XENVMAN_TPL.BASE_DIR` environment variable).

A template file name must follow the format: `<name>.tpl.js` and can be located
either directly within tpl base dir or in any sub-directory.

A fully qualified template name consists of javascript file name without `.tpl.js` suffix, preceeded by directory names relative to template base dir.

To make it clear, let's consider a simple example.
Let's say our base dir is `/opt/xenvman/base` and it looks like this:

```
/opt/xenvman/base/
   db/
      mysql.tpl.data/
      mongo.tpl.data/
      mysql.tpl.js
      mongo.tpl.js
   custom.tpl.data/
      Dockerfile
      custom.yaml
   custom.tpl.js
```

So here we have three templates with fully qualified names:
`db/mysql`, `db/mongo` and `custom`.

## Data directory

There's usually a bunch of files needed by template like Dockerfile to build
images on the fly, configuration templates, required modules, shared libraries
etc. All those files must be placed in a special directory called
`template data directory` (or just `data dir` for short).
Data dir must be located inside the same dir where template file is
and must be named using the following format: `<name>.tpl.data`, where `<name>`
is the same template name as in main json file.

Template javascript API provides functions to copy files from data dir to image
workspace, mount them inside containers etc.

Please note, that all files in data directory are never changed
by a template, they are always copied when needed.

## Workspace directory

Because `xenvman` allows you to build docker images on the fly,
there are often files you'd want to include in the image.
All those files are collected in a special temporary dir called
`workspace`. A workspace is a temporary directory, separately created for 
any image your template is trying to build during template execution.
The only required file is a Dockerfile itself, which describes what kind of
image you're building.

## Mount directory

A `mount directory` is a temporary dir created for every container
the template wants to run and holds files which will be mounted inside
the container. You can create files in a mount dir by either copying
them from a data dir (using container JS API) or by using data from template runtime parameters.

## Template directories summary

The following picture provides a general view of template directories
and their relations.

![Template directories](docs/img/dirs.png)

## Javascript API

## Interpolation
### Workspace files
### Mounted files

# HTTP API

# Clients
## Golang

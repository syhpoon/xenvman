[![Build Status](https://travis-ci.org/syhpoon/xenvman.svg?branch=master)](https://travis-ci.org/syhpoon/xenvman)
[![codecov](https://codecov.io/gh/syhpoon/xenvman/branch/master/graph/badge.svg)](https://codecov.io/gh/syhpoon/xenvman)
[![Go Report Card](https://goreportcard.com/badge/github.com/syhpoon/xenvman)](https://goreportcard.com/report/github.com/syhpoon/xenvman)

Table of Contents
=================

   * [Overview](#overview)
   * [Installation](#installation)
      * [Download release](#download-release)
      * [Compilation from source](#compilation-from-source)
      * [Configuration](#configuration)
         * [Configuration file](#configuration-file)
         * [Environment](#environment)
         * [api_auth (XENVMAN_API_AUTH) [""]](#api_auth-xenvman_api_auth-)
         * [auth_basic [""]](#auth_basic-)
         * [container_engine (XENVMAN_CONTAINER_ENGINE) ["docker"]](#container_engine-xenvman_container_engine-docker)
         * [export_address (XENVMAN_EXPORT_ADDRESS) ["localhost"]](#export_address-xenvman_export_address-localhost)
         * [keepalive (XENVMAN_KEEPALIVE) ["2m"]](#keepalive-xenvman_keepalive-2m)
         * [listen (XENVMAN_LISTEN) [":9876"]](#listen-xenvman_listen-9876)
         * [ports_range (XENVMAN_PORTS_RANGE) [[20000, 30000]]](#ports_range-xenvman_ports_range-20000-30000)
         * [tpl.base_dir (XENVMAN_TPL_BASE_DIR) [""]](#tplbase_dir-xenvman_tpl_base_dir-)
         * [tpl.ws_dir (XENVMAN_TPL_WS_DIR) [""]](#tplws_dir-xenvman_tpl_ws_dir-)
         * [tpl.mount_dir (XENVMAN_TPL_MOUNT_DIR) [""]](#tplmount_dir-xenvman_tpl_mount_dir-)
         * [tls.cert (XENVMAN_TLS_CERT) [""]](#tlscert-xenvman_tls_cert-)
         * [tls.key (XENVMAN_TLS_key) [""]](#tlskey-xenvman_tls_key-)
      * [Running API server](#running-api-server)
   * [Environments](#environments)
   * [Templates](#templates)
      * [Data directory](#data-directory)
      * [Workspace directory](#workspace-directory)
      * [Mount directory](#mount-directory)
      * [Template directories summary](#template-directories-summary)
      * [Javascript API](#javascript-api)
         * [Template format](#template-format)
         * [Template API](#template-api)
            * [BuildImage(name :: string) -&gt; <a href="#BuildImage-API">BuildImage</a>](#buildimagename--string---buildimage)
            * [FetchImage(name :: string) -&gt; <a href="#FetchImage-API">FetchImage</a>](#fetchimagename--string---fetchimage)
            * [AddReadinessCheck(name :: string, params :: object) -&gt; null](#addreadinesscheckname--string-params--object---null)
         * [BuildImage API](#buildimage-api)
            * [CopyDataToWorkspace(path :: string...) -&gt; null](#copydatatoworkspacepath--string---null)
            * [AddFileToWorkspace(path :: string, data :: string, mode int) -&gt; null](#addfiletoworkspacepath--string-data--string-mode-int---null)
            * [InterpolateWorkspaceFile(file :: string, data :: object) -&gt; null](#interpolateworkspacefilefile--string-data--object---null)
            * [NewContainer(name :: string) -&gt; <a href="#Container-API">Container</a>](#newcontainername--string---container)
         * [FetchImage API](#fetchimage-api)
            * [NewContainer(name :: string) -&gt; <a href="#Container-API">Container</a>](#newcontainername--string---container-1)
         * [Container API](#container-api)
            * [SetEnv(env, val :: string) -&gt; null](#setenvenv-val--string---null)
            * [SetLabel(key :: string, value :: {string, number}) -&gt; null](#setlabelkey--string-value--string-number---null)
            * [SetCmd(cmd :: string) -&gt; null](#setcmdcmd--string---null)
            * [SetPorts(port :: number...) -&gt; null](#setportsport--number---null)
            * [MountString(data, contFile :: string, mode :: int, opts :: object) -&gt; null](#mountstringdata-contfile--string-mode--int-opts--object---null)
            * [MountData(dataFile, contFile :: string, opts :: object) -&gt; null](#mountdatadatafile-contfile--string-opts--object---null)
         * [Readiness checks](#readiness-checks)
            * [http](#http)
         * [Helper JS functions](#helper-js-functions)
            * [fmt(format :: string, args :: any...)](#fmtformat--string-args--any)
            * [type](#type)
               * [type.EnsureString(arg :: any)](#typeensurestringarg--any)
               * [type.EnsureNumber(arg :: any)](#typeensurenumberarg--any)
               * [type.EnsureListOfStrings(arg :: any)](#typeensurelistofstringsarg--any)
               * [type.EnsureListOfNumbers(arg :: any)](#typeensurelistofnumbersarg--any)
               * [type.FromBase64(name :: string, value :: string)](#typefrombase64name--string-value--string)
               * [type.IsArray(arg :: any)](#typeisarrayarg--any)
               * [type.IsDefined(arg :: any)](#typeisdefinedarg--any)
      * [Interpolation](#interpolation)
         * [Workspace files interpolation](#workspace-files-interpolation)
         * [Mounted files, readiness checks &amp; environ interpolation](#mounted-files-readiness-checks--environ-interpolation)
   * [HTTP API](#http-api)
      * [POST /api/v1/env](#post-apiv1env)
         * [Body](#body)
            * [tpl](#tpl)
            * [env_options](#env_options)
         * [Response body](#response-body)
            * [tpl-data](#tpl-data)
            * [container-data](#container-data)
      * [DELETE /api/v1/env/{id}](#delete-apiv1envid)
         * [Query parameters](#query-parameters)
      * [POST /api/v1/env/{id}/keepalive](#post-apiv1envidkeepalive)
   * [Dynamic discovery](#dynamic-discovery)
   * [Clients](#clients)
      * [Golang](#golang)
      
# Overview

`xenvman` is an extensible environment manager which is used to
create environments for testing microservices.

![Overview](docs/img/overview.png)

xenvman can be used to:

* Define environment templates using JavaScript
* Create images on the fly
* Spawn as many containers as needed inside an environment
* Link containers together in a single isolated network
* Expose container ports for external access

For a detailed example take a look at [tutorial](https://medium.com/@syhpoon/xenvman-tutorial-c9967ddefaae).

# Installation

Please note, that even though `xenvman` binaries are provided for both
Linux and MacOS, at the moment only Linux is officially supported.

## Download release

Simply download the latest available binary for your OS/platform
from [here](https://github.com/syhpoon/xenvman/releases),
rename the binary to `xenvman` and place anywhere in your `$PATH`.

## Compilation from source

In order to compile `xenvman` from source you must have installed
[Golang](https://golang.org/) with the minimum version of `1.11`.

`xenvman` uses new feature introduced in Go version `1.11` - 
[Modules](https://github.com/golang/go/wiki/Modules) and so you can
clone the sources anywhere, no need to do it into `$GOPATH`.

The build process is super simple:
```bash
$ cd ~ && git clone https://github.com/syhpoon/xenvman.git && cd xenvman
$ make test && make build
```

If everything is good, there will be a `xenvman` executable in
the project root, which you can copy anywhere in your `$PATH`
and that would be it for the installation.

## Configuration

There are two ways to provide configuration: configuration file or
environment variables.

### Configuration file

`xenvman` uses [toml](https://github.com/toml-lang/toml) as a configuration
format. Most of the configuration parameters have reasonable default
values so you can run the program even without supplying any configuration
at all. An example file with all the available options can
be found [here](https://github.com/syhpoon/xenvman/blob/master/etc/xenvman.toml).

In order to provide custom configuration, create a `xenvman.toml` file
anywhere you like and run the server with `-c` option:

`xenvman run -c <path-to-xenvman.toml>`

### Environment

Configuration parameters can also be provided using environment variables.
The variable must be a capitalized version of config param with a special
`XENVMAN_` prefix.

For example, setting server listen address port can be done
using these both ways:

`listen = ":9876"` using configuration file, or

`XENVMAN_LISTEN=":9876"` - using env.

`Please note`: use underscore (`_`) to separate nested fields when using env,
not dots.

### api_auth (XENVMAN_API_AUTH) [""]

Type of authentication backend to use.
Available types include:

* `basic` - HTTP basic auth

### auth_basic [""]

Section specifying mapping from usernames to passwords for http basic auth.

### container_engine (XENVMAN_CONTAINER_ENGINE) ["docker"]

Type of container engine to use.
Currently only `docker` is supported.

### export_address (XENVMAN_EXPORT_ADDRESS) ["localhost"]

The external address to expose to clients.

### keepalive (XENVMAN_KEEPALIVE) ["2m"]

Default environment keepalive

### listen (XENVMAN_LISTEN) [":9876"]

IP:port to listen on.
If `IP` is ommitted, `localhost` will be used.

### ports_range (XENVMAN_PORTS_RANGE) [[20000, 30000]]

A port range from which to take exposed ports,
specified as a list of two [min, max] numbers.

### tpl.base_dir (XENVMAN_TPL_BASE_DIR) [""]

Base directory where to search for [templates](#Templates).

### tpl.ws_dir (XENVMAN_TPL_WS_DIR) [""]

Base directory where temporary image [workspaces](#Workspace-directory)
will be created.

### tpl.mount_dir (XENVMAN_TPL_MOUNT_DIR) [""]

Base directory where temporary container [mount dirs](#Mount-directory)
will be created.

### tls.cert (XENVMAN_TLS_CERT) [""]

Path to TLS certificate file. If not set, TLS mode will not be used.

### tls.key (XENVMAN_TLS_key) [""]

Path to TLS privatet key file. If not set, TLS mode will not be used.

## Running API server

Running `xenvman` server is very simple:

1. When using configuration file: `xenvman run -c <path-to-xenvman.toml>`
2. When using env variables: `XENVMAN_<PARAM>=<VALUE> xenvman run`

# Environments

Environment is an isolated bubble where one or more containers can be run 
together in order to provide a necessary playground for integration testing.

Environments are created, managed and destroyed using HTTP API provided
by running `xenvman` server.

`Please note`: here environment is `NOT` the usual shell one.

# Templates

An environment is set up by executing one or more templates,
where a template is a a small program written in JavaScript
which defines what images to build/fetch, what and how many containers
to spawn, what files to mount inside containers, what ports to expose etc.

A template script is run by embedded JS interpreter inside `xenvman` server.
One template is just one javascript file located within a template base directory (defined by `tpl.base-dir` configuration parameter, or `XENVMAN_TPL.BASE_DIR` environment variable).

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

As mentioned above, a template is a JavaScript program which
uses special API to configure required environment.
Let's take a closer look at template shape and form.

`Please note`: `xenvman` uses an embedded [JS interpreter](https://github.com/robertkrimen/otto), which implies certain limitations as compared
to running JS in a browser or in node.js ecosystem:

* No DOM-related functions
* `"use strict"` will parse, but does nothing
* The regular expression engine (re2/regexp) is not fully compatible with the ECMA5 specification
* Only ES5 is supported. ES6 features (eg: Typed Arrays) are not available

### Template format

A template must define an entry point function:

`function execute(tpl, params) {}`

This function is expected to provide necessary instructions in order
to configure an environment.

First parameter, `tpl`, is a [template instance](#Template-API), while
`params` is an arbitrary key-value object which is used to 
configure template by the caller.

`Please note`: calling tpl instance functions, such as `BuildImage`,
`FetchImage` etc. does not cause these actions to occur immediately,
instead they are scheduled and performed at later stages, after
JS execution phase.

### Template API

Template instance, which is passed as a first argument has the following methods:

#### BuildImage(name :: string) -> [BuildImage](#BuildImage-API)

Instucts `xenvman` to build a new image with the given name.
`name` parameter is the resulting Docker image name.

Return value is a [BuildImage](#BuildImage-API) instance.

#### FetchImage(name :: string) -> [FetchImage](#FetchImage-API)

Instructs `xenvman` to fetch an existing image from public or 
private image repository.
 
The `name` is a fully-qualified docker image name, including
repository address and tag, that is the same format is expected as
for regular `docker pull` invocation.

For private repos, existing credentials (acquired by `docker login`)
are used by the user who started `xenvman` server.

#### AddReadinessCheck(name :: string, params :: object) -> null

Adds a new [readiness check](#Readiness-checks) for the current template.

### BuildImage API

BuildImage instance represents an image which `xenvman` is going to build
on the fly. Files included in the image can be either copied from a
[template data dir](#Data-directory) or by supplying data for files in
template HTTP parameters.

#### CopyDataToWorkspace(path :: string...) -> null

This function takes a variable list of FS object names from data dir
and copies them into [image workspace](#Workspace-directory).
Object names must be relative to the data dir.
For example, if data dir contained the following files:

```
<data-dir>/
   subdir/
      subfile.png
   file1.json
```

then the paths would be: `subdir/subfile.png` and `file1.json`.

A special value `*` can be provided in order to copy every object from
data dir.

#### AddFileToWorkspace(path :: string, data :: string, mode int) -> null

Sometimes you want to dynamically include some file into the image
which is different every time you build it. So it cannot be simply
placed into data dir. Imagine you've patched some microservice
and want to test it, you can simply include the binary itself
(assuming your microservice is written in a compiled language)
in the HTTP request as a template parameter and by calling
`AddFileToWorkspace` it will be copied to image workspace.

* `path` argument is a path inside an image where to save the data.
* `data` is the data itself as a binary/string. Usually it is base64-encoded
  during HTTP transfer and then decoded back using `type.FromBase64()`
  js function.
* `mode` is a standard Unix file mode as an octal number.  

#### InterpolateWorkspaceFile(file :: string, data :: object) -> null

Instructs `xenvman` to interpolate a file in a workspace dir (that is
it must already be copied there before).

* `file` is a file path relative to workspace dir.
* `data` is an object providing values for interpolation.

[More details about interpolation](#Interpolation).

#### NewContainer(name :: string) -> [Container](#Container-API)

Create a new container with a given name from the image instance.

### FetchImage API

FetchImage instance represents an image which will be fetched by
`xenvman` (using Docker). Because in this case the image is already built
the amount of possible actions is limited as compared to building a new
image from scratch. Basically the only possible modification is mounting
files into the container from the host ([Mount dir](#Mount-directory)).

#### NewContainer(name :: string) -> [Container](#Container-API)

Create a new container with a given name from the image instance.

### Container API

#### SetEnv(env, val :: string) -> null

Sets a shell environment variable inside a container.

#### SetLabel(key :: string, value :: {string, number}) -> null

This function sets a container label. Labels here are `xenvman` entity
and are used later during [interpolation](#Interpolation) in order
to filter containers.

#### SetCmd(cmd :: string) -> null

Sets a [`CMD`](https://docs.docker.com/engine/reference/builder/#cmd) for the container.

#### SetPorts(port :: number...) -> null

Instructs `xenvman` to expose certain ports from the container.
Ports here are internal container ones, `xenvman` will select different
external ports for every exposed one.

#### MountString(data, contFile :: string, mode :: int, opts :: object) -> null

Instructs `xenvman` to mount the `data` string into a container
under the `contFile` name.
`mode` is a regular Linux file mode expressed as an octal int.
`opts` is an object, representing additional mounting parameters:

* `readonly` :: bool - If mounted file should be read only.
* `interpolate` :: bool - If the contents of a mounted file needs to be
                          interpolated.

#### MountData(dataFile, contFile :: string, opts :: object) -> null

Instructs `xenvman` to copy a `dataFile` from the [data dir](#Data directory) 
and mount it inside a container under `contFile` name.

In addition to `opts` from `MountString` above, `MountData` can take the
following:

* `skip-if-nonexistent` :: bool - If set to `true`, an error will not be
                                  raised if specified `dataFile` does not exist.

### Readiness checks

`xenvman` was primarily designed to create environments for
integration testing. Because of that, it needs to make sure
an environment is `ready` before returning the access data to the caller.
This is what readiness checks are for.

An environment can define any number of readiness checks and
`xenvman` will only return back to the caller after all the checks for
all the used templates are completed.

Readiness checks are added by calling [AddReadinessCheck()](#addreadinesscheckname--string-params--object---null)
function of `tpl` instance.

Please note, that every value in check parameters is
[interpolated](#Readiness-checks-interpolation).

Currently available readiness checks include:

#### http

As the name suggests this readiness check is used to ensure
readiness of a http service[s].

Availalable parameters include:

* `url` :: string - A HTTP URL to try fetching.
* `codes` :: [int] - A list of successful HTTP response codes.
                     At lest one must match in order for a check to
                     be considered successful.
* `headers` :: [object] - A list of header objects to match.
	                      Values within the same objects are matched
	                      in a conjuctive way (AND).
	                      Values from different objects are matched in
	                      a disjunctive way (OR).
* `body` :: string - A regexp to match response body against.
* `retry_limit` :: int - How many times to retry a check before giving up.
* `retry_interval` :: string - How long to wait between retrying.
                               String must follow Golang [`fmt.Duration`](https://golang.org/pkg/time/#ParseDuration) format.

### Helper JS functions

In addition to image and container specific APIs there are also some
additional helper modules which can be used directly anywhere in the template
script.

#### fmt(format :: string, args :: any...)

A useful shorthand for printing formatted messages.
It's nothing more than an exported Golang function [`fmt.Printf`](https://golang.org/pkg/fmt/#Printf).

#### type

`type` module contains functions related to managing types.

All `Ensure*` functions take a value and panic if the value is not
of correspdonging type. It passes otherwise (including value not
being defined).

##### type.EnsureString(arg :: any)
##### type.EnsureNumber(arg :: any)
##### type.EnsureListOfStrings(arg :: any)
##### type.EnsureListOfNumbers(arg :: any)
##### type.FromBase64(name :: string, value :: string)

Decodes a value from base64 string to a byte array.
`name` argument is only used for logging in case of errors.

##### type.IsArray(arg :: any)

Returns true if given argument is of array type.

##### type.IsDefined(arg :: any)

Returns true if given argument is neither `null` nor `undefined`.

## Interpolation
TODO

### Workspace files interpolation
TODO

### Mounted files, readiness checks & environ interpolation
TODO

# HTTP API

`xenvman` exposes all its functionality using HTTP API.

## POST /api/v1/env

Create a new environment.

### Body

```
{
  // Environment name
  "name" :: string,
 
  // Environment description
  "description" :: string,
 
  // Templates to use
  "templates" :: [tpl]

  // Additional env options
  "options" :: env_options
}
```

#### tpl
```
{
  // Template name (a path relative to xenvman base template dir)
  "tpl" :: string,
  
  // Template parameters as arbitrary JSON object
  "parameters" :: object 
}
```

#### env_options
```
{
  // Environment keep alive setting
  "keep_alive" :: string,
  
  // Whether to disable dynamic discovery DNS agent and revert back to static
  // hostnames
  "disable_discovery" :: bool
}
```

### Response body

```
{
  // Environment id
  "id" :: string,
  
  // Templates data
  "templates" :: {name :: string -> [tpl-data]}
}
```

#### tpl-data
```
{
  // Template containers
  "containers" :: {name :: string -> [container-data]}
}
```

#### container-data
```
{
  // Unique container id
  "id"    :: string,
  // container port to exposed address:port
  // Exposed address contains public xenvman ip plus exposed container port
  "ports" :: {port :: int -> string}
}
```

## PATCH /api/v1/env/{id}

Update existing environment.

### Body

```
{
  // A list of fully-qualified container names to stop
  "stop_containers" :: [string],
  
  // A list of fully-qualified container names to restart
  "start_containers" :: [string],
 
  // New templates to execute
  "templates" :: [tpl]
}
```

## DELETE /api/v1/env/{id}

Delete an environment.

### Query parameters

* id - Environment id

## POST /api/v1/env/{id}/keepalive

Keep alive an environment.

A client needs to periodicall call this endpoint in order to keep
the environemnt running.

# Dynamic discovery
TODO

# Clients

Because `xenvman` uses plain HTTP API, any language/tool capable of
talking HTTP can be used as a client. But it's arguably easier to have
native and idiomatic libraries for a language of choice, especially
to embed managing environments directly into integration tests themselves.

Currently `xenvman` only has support for `Go` language client.

## Golang
TODO

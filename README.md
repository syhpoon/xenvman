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
            * [SetCmd(cmd :: string...) -&gt; null](#setcmdcmd--string---null)
            * [SetEntrypoint(cmd :: string...) -&gt; null](#setentrypointcmd--string---null)
            * [SetPorts(port :: number...) -&gt; null](#setportsport--number---null)
            * [MountString(data, contFile :: string, mode :: int, opts :: object) -&gt; null](#mountstringdata-contfile--string-mode--int-opts--object---null)
            * [MountData(dataFile, contFile :: string, opts :: object) -&gt; null](#mountdatadatafile-contfile--string-opts--object---null)
         * [Readiness checks](#readiness-checks)
            * [http](#http)
            * [net](#net)
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
            * [.Self -&gt; Container](#self---container)
            * [.Extra -&gt; Any](#extra---any)
            * [.ExternalAddress -&gt; string](#externaladdress---string)
            * [.ContainersWithLabels(label : string, value : string) -&gt; [Container]](#containerswithlabelslabel--string-value--string---container)
            * [.ContainerWithLabel(label : string, value : string) -&gt; Container](#containerwithlabellabel--string-value--string---container)
            * [.AllContainers() -&gt; [Container]](#allcontainers---container)
            * [Container instance methods](#container-instance-methods)
               * [.IP -&gt; string](#ip---string)
               * [.Hostname -&gt; string](#hostname---string)
               * [.Name -&gt; string](#name---string)
               * [.GetLabel(label : string) -&gt; string](#getlabellabel--string---string)
               * [.ExposedPort(iport : int) -&gt; int](#exposedportiport--int---int)
   * [HTTP API](#http-api)
      * [GET /api/v1/env](#get-apiv1env)
         * [Response body](#response-body)
      * [POST /api/v1/env](#post-apiv1env)
         * [Body](#body)
         * [Response body](#response-body-1)
      * [GET /api/v1/env/{id}](#get-apiv1envid)
         * [Response body](#response-body-2)
      * [PATCH /api/v1/env/{id}](#patch-apiv1envid)
         * [Body](#body-1)
         * [Response body](#response-body-3)
      * [DELETE /api/v1/env/{id}](#delete-apiv1envid)
         * [Query parameters](#query-parameters)
      * [POST /api/v1/env/{id}/keepalive](#post-apiv1envidkeepalive)
      * [GET /api/v1/tpl](#get-apiv1tpl)
         * [Response body](#response-body-4)
      * [Types](#types)
         * [InputEnv](#inputenv)
         * [InputEnvOptions](#inputenvoptions)
         * [OutputEnv](#outputenv)
         * [PatchEnv](#patchenv)
         * [InputTpl](#inputtpl)
         * [TplData](#tpldata)
         * [ContainerData](#containerdata)
         * [TplInfo](#tplinfo)
         * [TplInfoParam](#tplinfoparam)
   * [Dynamic discovery](#dynamic-discovery)
   * [Dynamic environment reconfiguration](#dynamic-environment-reconfiguration)
   * [Web UI](#web-ui)
   * [Clients](#clients)
      * [Golang](#golang)
      * [Python](#python)

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
* Dynamically change environment composition (add, stop, restart containers) on the fly

For a detailed example take a look at [tutorial](http://syhpoon.ca/posts/xenvman-tutorial/).

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

#### SetCmd(cmd :: string...) -> null

Sets a [`CMD`](https://docs.docker.com/engine/reference/builder/#cmd) for the container.

#### SetEntrypoint(cmd :: string...) -> null

Sets an [`ENTRYPOINT`](https://docs.docker.com/engine/reference/builder/#entrypoint) for the container.

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
* `extra-interpolate-data` :: object - Additional interpolation data.
                                       The data is accessible under
                                       `.Extra` key of a container
                                       instance inside templates.

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

#### net

A simple low-level network readiness check.
`protocol` and `address` parametes must be formatted according to
Golang [`net.Dial`](https://golang.org/pkg/net/#Dial) function.

Availalable parameters include:

* `protocol` :: string - Network protocol 
* `address` :: string - Address string
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

Sometimes a static file, either embedded into an image or mounted
into a container in runtime is not enough, we need to be able to
include some dynamic parts, parts which can take different values
from environment to environment. For example you may need to
specify a database address in a config for your service.
But the hostname will always be different, you cannot just hardcode it.

This is where interpolation kicks in. Interpolation is just a fancy name
for variable substitution. Basically you just reserve certain placeholders
in your configs and they will be filled with needed values in due time.

There are two main types of interpolation in `xenvman`:

1. Workspace files
2. Mounted files, readiness checks and environ interpolation

The main difference between them is the available data.

For workspace files the only data you can substitute is the
one you supply yourself. The reason for this is that workspace files
are baked into an image and thus cannot be modified in any way
during container lifetime.

All the rest, on the other hand, do have access to some runtime info
like containers, ports etc.

All the interpolation is done using [Golang template language](https://golang.org/pkg/text/template/).

### Workspace files interpolation

Workspace files interpolation is very simple, you just call
[InterpolateWorkspaceFile()](#interpolateworkspacefilefile--string-data--object---null) function on
a file you want to interpolate. You must copy the file using
[CopyDataToWorkspace()](#copydatatoworkspacepath--string---null) or 
[AddFileToWorkspace()](#addfiletoworkspacepath--string-data--string-mode-int---null) first. 

You can supply arbitrary object and its fields in your interpolation
placeholders.

For example, let's say we have a `Dockerfile` in our data dir.
It will be included into an image every time we build it.
And we allow clients to supply their own executable binary.
Thus we don't know what the binary will be so we cannot hardcode
the executable name and so we'll use interpolation for it.

Let's examine template code first:

```javascript
function execute(tpl, params) {
    var img = tpl.BuildImage("service-%s", params.service);
    
    img.CopyDataToWorkspace("Dockerfile");
    img.InterpolateWorkspaceFile("Dockerfile", {"service": params.service});
}
```

And the Dockerfile itself:

```dockerfile
FROM ubuntu

COPY {{.service}} /
CMD ["/{{.service}}", "run"]
```

Here `{{.service}}` will be substituted with whaterver was provided in

```javascript
    img.InterpolateWorkspaceFile("Dockerfile", {"service": params.service});
```

in our template.

### Mounted files, readiness checks & environ interpolation

For this type of interpolation, in addition to providing your own
placeholder data, there's also some internal environment-specific data
available for you.

Let's take a look at what's available:

#### .Self -> Container

Return a current container instance.

#### .Extra -> Any

Returns a user-provided data, if any.

#### .ExternalAddress -> string

Returns an external address.

#### .ContainersWithLabels(label : string, value : string) -> [Container]

Find all containers with a given label name and value.
Empty `value` matches any label.

#### .ContainerWithLabel(label : string, value : string) -> Container

Find a container with a given label name and value.
Empty `value` matches any label.

#### .AllContainers() -> [Container]

Return a list of all containers in the environment.

#### Container instance methods

##### .IP -> string

Returns internal container IP address.

##### .Hostname -> string

Returns container hostname.

##### .Name -> string

Returns container name.

##### .GetLabel(label : string) -> string

Returns label value. Empty string is returned if there's no such label
on the container.

##### .ExposedPort(iport : int) -> int

Returns an external (exposed) port for the given internal one.
It's an error if there's no such port exposed on the container.

# HTTP API

`xenvman` exposes all its functionality using HTTP API.

## GET /api/v1/env

List active environments.

### Response body

[[OutputEnv]](#outputenv)

## POST /api/v1/env

Create a new environment.

### Body

[InputEnv](#inputenv)

### Response body

[OutputEnv](#outputenv)

## GET /api/v1/env/{id}

Get environment info.

### Response body

[OutputEnv](#outputenv)

## PATCH /api/v1/env/{id}

Update existing environment.

### Body

[PatchEnv](#patchenv)

### Response body

[OutputEnv](#outputenv)

## DELETE /api/v1/env/{id}

Delete an environment.

### Query parameters

* id - Environment id

## POST /api/v1/env/{id}/keepalive

Keep alive an environment.

A client can periodicall call this endpoint in order to keep
the environment running. Otherwise an environment will be terminated
after the configured keepalive interval.

## GET /api/v1/tpl

Get templates info.

### Response body
```{name: string -> TplInfo}```

## Types

### InputEnv
```
{
   // Environment name
   name: string,
 
   // Environment description
   description: string,
 
   // Templates to use
   templates: [InputTpl]

   // Additional env options
   options: InputEnvOptions
}
```

### InputEnvOptions
```
{
  // Environment keep alive setting
  keep_alive: string,
  
  // Whether to disable dynamic discovery DNS agent and revert back to static
  // hostnames
  disable_discovery: bool
}
```

### OutputEnv
```
{
    // Environment id
    id: string,
	
    // Environment name
    name: string,
    
    // Environment description
    description: string,
    
    // Workspace directory
    ws_dir: string,
    
    // Mount directory
    mount_dir: string,
    
    // Container engine network id for the environment
    net_id: string,
    
    // Creation time
    created: string,
    
    // Environment keep alive setting
    keep_alive: string,
    
    // External address (hostname or IP) of the xenvman server
    external_address: string,
    
    // Templates data
    templates: {name: string -> [TplData]}
}
```

### PatchEnv

```
{
   // A list of fully-qualified container names to stop
   stop_containers: [string],
  
   // A list of fully-qualified container names to restart
   start_containers: [string],
 
   // New templates to execute
   templates: [InputTpl]
}
```

### InputTpl
```
{
  // Template name (a path relative to xenvman base template dir)
  tpl: string,
  
  // Template parameters as arbitrary JSON object
  parameters: object 
}
```

### TplData
```
{
   // Template containers
   containers: {name: string -> [ContainerData]}
}
```

### ContainerData
```
{
   // Unique container id
   id: string,
   // Internal container hostname
   hostname: string,
   // Mapping between internal container port and corresponding external one
   ports: {port: string -> int}
}
```

### TplInfo
```
   // Template description
   description: string,
   
   // Template parameters
   parameters: {name: string -> TplInfoParam},
   
   // List of files in template data directory
   data_dir: [string]
```

### TplInfoParam
```
   // Parameter description
   description: string,
   
   // Parameter type
   type: string,
   
   // Whether a parameter is mandatory
   mandatory: bool,
   
   // Default value
   default: any,
```

# Dynamic discovery

In the version `v2.0.0` a new dynamic discovery agent has been introduced.
It is basically a simple DNS proxy configurable over HTTP which
all the containers running inside an environment are configured to use.
This allows us to dynamicall re-configure environment on-the-fly,
add/update/stop containers and make sure newly added containers can 
be discovered by the old ones.

The discovery agent is injected as just any another template and
a [tiny image from docker hub](https://hub.docker.com/r/syhpoon/xenvman)
is fetched with size of about 8 mb. 
This template creates a single container with a running agent.
The hostname of the container is `discovery.0.discovery.xenv` and
it exposes port `8080` over which it is configured later by
`xenvman` server.

You can disable this feature and opt in for static DNS config by
setting `disable_discovery=true` [env option](#env_options).
If you do this though, you will not be able to dynamically
change the environment composition (add/stop containers)
after it has started.

By default discovery agent is enabled.

The following picture illustrates these two different approaches:

![Static vs Dynamic DNS configuration](docs/img/discovery_agent.png)

# Dynamic environment reconfiguration

Starting from version `v2.0.0` it is possible to change environment
composition while it is running. One can stop existing containers 
(perhaps to emulate network split) or inject new templates and
introduce new containers (peers arrival). A new API endpoint
have been added for this purpose: `PATCH /api/v1/env/{id}`.

`Please note:` the environment reconfiguration is only available if
[dynamic agent](#dynamic-discovery) has not been disabled.

# Web UI

Starting from version `v2.0.0` xenman has a simple embedded web application.
It can be used to:

* Inspect currently running environments
* Terminate an environment
* Browse through all available templates
* Inspect invidual templates and its parameters

Once `xenvman` is running simply point your browser at
`http://<HOST>:<PORT>/`, where `<HOST>` is the hostname/ip where `xenvman`
is running and `<PORT>` is the post which `xenvman` is listening on.

List of active environments:
![List of environments](docs/img/webapp-1.png)

Environment info:
![Environment info](docs/img/webapp-2.png)

Templates browser:
![Templates browser](docs/img/webapp-3.png)

# Clients

Because `xenvman` uses plain HTTP API, any language/tool capable of
talking HTTP can be used as a client. But it's arguably easier to have
native and idiomatic libraries for a language of choice, especially
to embed managing environments directly into integration tests themselves.

Currently `xenvman` supports the following language clients:

## Golang

Go documentation for client package is available
[here](https://godoc.org/github.com/syhpoon/xenvman/pkg/client).

An example of how to use the client API is available
in [xenvman-tutorial](https://github.com/syhpoon/xenvman-tutorial/blob/master/bro_xenv_test.go).

## Python

Python client is available [here](https://github.com/syhpoon/xenvman-python).

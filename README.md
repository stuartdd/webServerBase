# webServerBase

* [Packages](#Packages)
* [Directories](#Directories)
* [Root Files](#Root%20Files)
* [Logger Configuration](#Logger%20Configuration)
  * [File Name substitutions](#File%20Name%20substitutions)
* [Web Server Modes](#Web%20Server%20Modes)
  * [Static files](#Static%20files)
  * [Redirection](#Redirection)
  * [Template files](#Template%20files)
* [Templates](#Templates)
  * [What is a Template](#What%20is%20a%20Template)
  * [Single Templates](#Single%20Templates)
  * [Group Templates](#Group%20Templates)
  * [Template Data](#Template%20Data)
* [Server start](#Server%20start)
* [Docker](#Docker)

[comment]: <> (Use Crtl+Shift+v | Crtl+k+v to view in Visual Code MD pluggin markdownlint)
[comment]: <> (Use Crtl+Shift+v | Crtl+k+v to view in Visual Code MD pluggin markdownlint)

GoLang based ReST based Server

For a full annotated example refer to **webServerExample.go** in the root directory

## Packages

[Top](#webServerBase)

* **servermain** - This is where all the server code resides
  * Dependends on **config** and **logging**
  * Does NOT depend on any other package (except standard go)
* **config** - This is where the configuration data is read and accessed
* **logging** - This is an enhanced logging framework for the server logs
* **test** - Tools used during testing. Asserts and the like!

## Directories

[Top](#webServerBase)

* **site** - Directory without code (TEST DATA ONLY).
  * This is for testing only and has test data files such as:
    * icons, html, json and even some xml. See webServerTest.json for paths to this diractory
* **templates** - Directory without code (TEST DATA ONLY).
  * This is for testing only contain templates and templat configuration files
* **.vscode** - Configuration files for Visual Studio Code IDE.
  * launch.json - for build and run (stand alone)
  * tasks.json for debugging (stand alone)

## Root Files

[Top](#webServerBase)

* **webServerBase.code-workspace** - File used by Visual Studio Code IDE
* **webServerExample.go** - Web server uses as an **Example** and **TEST** server
* **webServerExample_test.go** - Tests for **webServerExample.go**:
  * Start server.
  * Do loads of tests.
  * Stop server!
* **webServerExample_test.json** - Specific configuration file for **webServerExample_test.go**.
* **go.mod** and **go.sum** - These are generated by the go modules system and define dependencies for the project. Currently the project has no external deps so the definition in here '**replaces**' the dependency with the local packages. For example **config** and **logging**. If these are missing use ```go mod init <mainmodule>```. To updete the dependencies use ```go mod tidy```.
* README.md - This file!
* **LICENSE** - The open source lisence for this application

## Configuration

[Top](#webServerBase)

The server '**servermain**' and '**logging**' packages are NOT configured using the 'config' package. Creating a server and configuring the logs is done using standard Go types such as map and array. This separation is necessary so that '**servermain**' and '**logging**' can be used without constraining it to predefined configuration types and simplifies dependencies.

## Server Configuration

[Top](#webServerBase)

### See webServerExample.go in the root directory for a complete example!

The sturcture **ServerInstanceData** in the **servermain** package defines the configuration of a server instance. The code in th e**servermain** package uses this data to create an instance of a server. Multiple servers can be created by creating multiple **ServerInstanceData** objects.

A factory method **NewServerInstanceData** is called to create and validate a **ServerInstanceData** object. It returns a pointer to the **ServerInstanceData**.

**NewServerInstanceData** has the following parameters:

* baseHandlerNameIn string - Is the 'Name of the server'
  * This is the prefered name of the server for loggin pusposes. For example:
  * ```2019/09/25 14:52:26.350161 <baseHandlerNameIn>:8080 ServerMain [-]   INFO Server will start on port 8080```
* contentTypeCharsetIn string - Is the default response character set (utf-8).
  * So this is the default value for the ContentType header. For example a JSON response header would be:
    * application/json; charset=utf-8
    * The application/json is derived from the code. See later for details
    * The charset=utf-8 is derive from "charset=" + contentTypeCharsetIn.

## Web Server Modes

[Top](#webServerBase)

A web server can function in two modes.

* A Web application, serving html pages, images icons etc..
* A ReST server returning pure data as JSON or XML to a browser running JavaScript (or any application for that matter).
* OR BOTH. There is no reason why not other than design.

This web server (WebServerBase) can do both. It uses URL Mappings for ReST style responses and Static files (or Templates) for html and other web resources.

Note BOTH modes requests require URL Mappings, they just change where teh data comes from.

## Static files

[Top](#webServerBase)

This feature is supportted by the **staticFileDataManager.go** code in the servermain package.

Static files are 'files' in the file system that are returned unchanged. The server just needs to know where they are held. To set the root of the static file to a specific directory, use the following:

``` go
m := make(map[string]string)
m["/static/"] = "my/Static/files"
m["data"] = "site/"
servermain.SetStaticFileServerData("myfiles/path")
```

Add a mapping to the **DefaultStaticFileHandler** and your done.

``` go
serverInstance.AddMappedHandler("/static/*", http.MethodGet, servermain.DefaultStaticFileHandler)
```

* Note - The '*' in the above mapping indicates match ANY text after */static/*

* Note - **servermain.DefaultStaticFileHandler** is an existing handler for this purpose defined in **requestHandlers.go**. It is basic but sufficient for this function. Please feel free to use it as a basis for your own method.

* Note - the directory paths given in the example will be relative to the server directory.

The following request Will return the contents of **my/Static/files/index.html**

``` http
http://localhost:8080/static/index.html
```

If you add a [Redirection](#Redirection) (See below) as follows:

``` go
m := make(map[string]string)
m["/"] = "/static/index.html"
serverInstance.SetRedirections(m)
```

Then ```http://localhost:8080/``` and ```http://localhost:8080``` will also return the contents of **my/Static/files/index.html**

This is really usefull fo setting a home page.

## Redirection

Redirection allows you to recognise a url, substitute another and send it.

So as in the above example the URL */* is redirected to */static/index.html*. 

Note *Redirections* are always processes first and might cause an infinite loop. For example if you redirect '/' to '/'

Create a map of redirections and call *SetRedirections*

``` go
m := make(map[string]string)
m["/"] = "/static/index.html"
serverInstance.SetRedirections(m)
```

Query parameters are copied over to the redirected URL. For example with the following redirect:

``` go
m["/a/b"] = "/d/XYZ"
```

So ```/a/b?q1=A&Q2=B``` will be mapped to ```/a/XYZ?redirect=true&q1=A&Q2=B```

Note the additional query parameter that indicates a redirect has taken place. This should not impact any requests unless the handler look for the query value 'redirect'.

## Template files

[Top](#webServerBase)

Template files are 'files' in the file system that are returned after they have been updated by the required data. Again the server just needs to know where they are held.

``` go
serverInstance.SetPathToTemplates("my/templates)
```

In this example 'my/templates' is a relative directory path.

* This Loads all templates from that path and caches them in memory!

Please refer to the section on [Templates](#Templates) for further details.

## Config Package

[Top](#webServerBase)

This reads a JSON file in to a structure that is then used to configure the server and the logger. However it is not a dependency of either.

If you design the structure to contain the exact parameters required to configure the server and logger then your 'main' can load the configuration data from a JSON file (or any other type of file) and pass the componentes (values) contained in that structure in to the server and logger methods.

See **webServerExample.go** for an example.

## Logger Configuration

[Top](#webServerBase)

The logger (package "logger") is a wrapper for the Go 'log' package. This package manages writes to all logs. The logger wrapper adds functionality to the standard go package.

The function **CreateLogWithFilenameAndAppID** is used to create the logger data structure. It does NOT return a reference to this data structure as the logger is a singleton and simply requires initialiasation.

The parameters to **CreateLogWithFilenameAndAppID** initialise the logger and the various logger levels.

* **defaultLogFileNameIn string** - Is the default file name for loggers that write to files.
  * If this is an empty string then the default output will  be system out (console)
  * If this is a file name then further logging the writes to the default file will append to this log.
* **applicationID string** - This is the name associated with the log:
  * ```2019/09/25 14:52:26.350161 <applicationID>:8080 ServerMain [-]   INFO Server will start on port 8080```
* **fatalRCIn int** - This is the application return code for 'Fatal' errors that cause the server to exit.
  * The **Fatal** log function logs a message and stack trace. If fatalRCIn is NON zero it will finally terminate the server application returning the value of fatalRCIn to the operating system.
* **logLevelActivationData map[string]string** - Defines the output and active properties for each log level.
  * The log levels are currently: 
    * **DEBUG** - Used by the developers for diagnostic (non error) log entries
    * **INFO** - General information log entries
    * **ACCESS**
      * Used to log requests to the server (including request headers if **DEBUG** is active)
      * Used to log responses from the server (including response headers if **DEBUG** is active)
    * **WARN** - Used to indicate non-critical errors that do not impact functionality but should be logged
    * **ERROR** - Used to indicate critical errors that will impact functionality but do not impact the overall server function
    * **FATAL** - The server cannot continue to function. If **fatalRCIn** is non zero then Exit and return **fatalRCIn** as the response code

Unlike most loggers there is no implied hierarchy to the log levels dispite being called levels. In this logger all levels are independent and each one is separatly configurable.

Each log level can be configured as follows:

* **OFF** (or inactive) - No logging will take place. A call to the function Is* will return false. For example **IsDebug()** will return false if **DEBUG** is OFF, IsWarn() will return true if **WARN** is not OFF.
* **DEFAULT** (active) - Log lines will be written to  **defaultLogFileNameIn** if it is defined. If not then log lines will be written to the console (SYSOUT).
* **SYSOUT** - Log lines will be written to the console (system out)
* **SYSERR** - Log lines will be written to the console (system error out)
* **FileName** - If a file name is given then the log lines will be appended to that file. This is independent of **defaultLogFileNameIn**.

The default state for each log level is as follows:

* **DEBUG** - OFF
* **INFO** - OFF
* **ACCESS** - OFF
* **WARN** - OFF
* **ERROR** - SYSERR
* **FATAL** - SYSERR

The **CreateTestLogger()** function sets ALL log levels to **SYSOUT** except for **ERROR** and **FATAL** which remain as **SYSERR**

This example sets DEBUG to active and all log lines will be written to the console.

``` Go
levels := make(map[string]string)
levels["DEBUG"] = "SYSOUT"
CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
```

This example sets DEBUG to active and all log lines will be written to the file.

``` Go
levels := make(map[string]string)
levels["DEBUG"] = "/logs/%ID-%YYYY-%MM-%DD-%HH-%MM-%SS.log"
levels["WARN"] = "SYSOUT"
CreateLogWithFilenameAndAppID("", "AppID", -1, levels)
```

### File Name substitutions

* **%YYYY** - 4 digit year
* **%MM** - 2 Digit month
* **%DD** - 2 Digit dat of month
* **%HH** - 2 Digit hour of the day (24 hour format)
* **%mm** - 2 Digit minute of the hour
* **%SS** - 2 Digit seconds of the minute
* **%ID** - The application ID (**applicationID**)
* **%PID** - The application process id
* **%ENV.[name].ENV%** - The value of a named environment variable. E.g. %ENV.HOSTNAME.ENV% will return the host name of the system

## Templates

[Top](#webServerBase)

Support for the GO templating engine is included in the server. You can implement this yourself or use the Default template processing provided.

This can be used for example to return a html document that has data values substituted in to it. Templating in GO is quite comples so I would recommend reading th documentation. 

The process in this server:

* Loads ALL the templates in the template directory in to memory.
* Maps HTTP requests to them.
* Provided an oportunity to derive template data based on the template name.

Template support is in **templateManager.go** with examples in **webServerExample.go**

First we need to tell the server where ALL the templates are held in the file system:

``` go
serverInstance.SetPathToTemplates("templates/")
```

This will Load ALL the templates in to a cache in memory and perform some validation of the syntax.

We can then get a list of available template name and log it.

``` go
serverInstance.ListTemplateNames(", ")
```

Will list the templates with a ', ' between each name.

## What is a Template

[Top](#webServerBase)

You will need to read the documentation (I had to!).

You can define *single* templates or *groups* of templates that can generate a single document.

The template directory is parsed and all qualifying templates are loaded.

## Single Templates

[Top](#webServerBase)

These are files in the templates directory that have the following file naming convention:

\<**templateNme**\>.template.\<**templateExtension**\>

All files in the directory that have that naming convention will be picked up. For example:

* myTemplate.template.html - Template name will be myTemplate.html
* another.file.template.txt - Template name will be another.file.txt

## Group Templates

[Top](#webServerBase)

A special JSON file is used to define a group of templates as a single entity. It can define multiple entities and there can be multiple files. 

It should have the following file naming convention:

\<**prefix**\>.template.groups.json

For example:

* super.template.groups.json - Note the prefix is ignored. It is used to enable multiple group definition files to be read

The template name is defined in the JSON file not the file name of the group file. The format is as follows:

``` json
[
    {
        "name": "composite1.html",
        "templates": [
            "import1.html",
            "part1.html",
            "part2.html"
        ]
    }, {
        "name": "composite2.html",
        "templates": [
            "import2.html",
            "part3.html",
            "part4.html"
        ]
    }
]
```

This defines two templates composite1.html and composite2.html.

* composite1.html - is made up of  import1.html, part1.html and part2.html
* composite2.html - is made up of  import2.html, part3.html and part4.html

## Template Mapping

[Top](#webServerBase)

A Mapping Handler is required to process templates. There is an example in **webServerExample.go**:

``` go
serverInstance.AddMappedHandler("/site/?", http.MethodGet, servermain.DefaultTemplateFileHandler)
```

A default template handler is defined in **requestHandlers.go**. You can use this as it is or use it to base your own version on.

## Template Data

[Top](#webServerBase)

To derive data for your templates you need a data object (or data set). This is an interface{} and is often a map of type **map[string]string**.

However if you read the GO documentation for templates it can be all sorts of things that contain data for substitution in templates.

You can delegate to a function to create a data set for any template. Create a function like this one:

``` go
func templateDataProvider(r *http.Request, templateName string, data interface{})
```

There is an example defined in **webServerExample.go**.

Then add that function to the server:

``` go
serverInstance.AddTemplateDataProvider(templateDataProvider)
```

Any template that is executed will call that function before it is resolved. This is your opportunity to define some data for the template

The example **templateDataProvider** in **webServerExample.go** reads a map from the config data and based on the template name adds the values in the map to the data.

Based on the template name you could call a data base or read a file and add the data that way.

The reasonable template handler defined in **templateManager.go** adds the URL Query data to a map before calling **templateDataProvider** so the map already has some data in it.

## Server start

[Top](#webServerBase)

Once you have done all the setup the last thing to do is Start the server.

To start the server use the **ListenAndServeOnPort(port int)** method bound to **ServerInstanceData**

For example:

``` go
serverInstance = servermain.NewServerInstanceData("MyServer", "utf-8)
/*
Do all the setup!
*/
serverInstance.ListenAndServeOnPort(8080)
```

This command does not return until the server terminates!

## Docker

Ref: [Docker Example](https://www.callicoder.com/docker-golang-image-container-example/)
# webServerBase

GoLang based ReST based Server

## Packages

* **config** - Contains server configuration data read from JSON file
* **logging** - Contains a wrapper for the go logging framework that is configurable and has more functionality
* **servermain** - This is where all the server code resides
* **test** - Contains test tools for doing Asserts etc...

## Directories

* **site** - Directory without code (TEST DATA ONLY).
  * This is for testing only and has test data files such as:
    * icons, html, json and even some xml. See webServerTest.json for paths to this diractory
* **templates** - Directory without code (TEST DATA ONLY).
  * This is for testing only contain templates and templat configuration files
* **.vscode** - Configuration files for Visual Studio Code IDE.
  * launch.json - for build and run (stand alone)
  * tasks.json for debugging (stand alone) 

## Files (in root dir)

* **webServerBase.code-workspace** - File used by Visual Studio Code IDE
* **webServerBase.go** - Web server for **TEST** purposes and as a model for other servers
* **webServerBase_test.go** - Tests for **webServerBase.go**:
  * Start server.
  * Do loads of tests. 
  * Stop server!
* **webServerTest.json** - Specific configuration file for **webServerBase_test.go**.
* **webServerBase.json** - Specific configuration file for **webServerBase.go** when running as a standalone server.
* README.md - This file!
* **LICENSE** - The open source lisence for this application

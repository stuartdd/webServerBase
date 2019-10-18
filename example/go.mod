module webServerExample.go

go 1.13

require github.com/stuartdd/webServerBase v0.0.0-20191018161238-f7126594efcb

replace github.com/stuartdd/webServerBase/config => ../webServerBase/config

replace github.com/stuartdd/webServerBase/logging => ../webServerBase/logging

replace github.com/stuartdd/webServerBase/panicapi => ../webServerBase/panicapi

replace github.com/stuartdd/webServerBase/servermain => ../webServerBase/servermain

replace github.com/stuartdd/webServerBase/exec => ../webServerBase/exec

replace github.com/stuartdd/webServerBase/largefile => ../webServerBase/largefile

replace github.com/stuartdd/webServerBase/substitution => ../webServerBase/substitution

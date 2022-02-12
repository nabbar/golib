# Cobra pakcage
Help build integration of spf13/Cobra lib. 
This package will simplify call and use of flags / command for a CLI Tools.

## Exmaple of implement
Make some folder / file like this: 
```
/api
|- root.go
|- one_command.go
/pkg
|- common.go
/main.go
```

### Main init :

This `commong.go` file will make the init of version, logger, cobra, viper and config packages.

```go
import (
	"strings"

	libcbr "github.com/nabbar/golib/cobra"
	libcfg "github.com/nabbar/golib/config"
	liblog "github.com/nabbar/golib/logger"
	libver "github.com/nabbar/golib/version"
	libvpr "github.com/nabbar/golib/viper"
)
```

Some variables :
```go
var(
	// Golib Version Package
    vrs libver.Version
	
	// Main config / context Golib package
	cfg libcfg.Config
	
	// Root logger before config 
	log liblog.Logger
	
    // Main Cobra / Viper resource
    cbr libcbr.Cobra
	vpr libvpr.Viper
)
```

This function will init the libs if not set and return the lib resource :
```go
func GetViper() libvpr.Viper {
    if vpr == nil {
        vpr = libvpr.New()
        vpr.SetHomeBaseName(strings.ToLower(GetVersion().GetPackage()))
        vpr.SetEnvVarsPrefix(GetVersion().GetPrefix())
        vpr.SetDefaultConfig(GetConfig().DefaultConfig)
    }

    return vpr
}

func GetCobra() libcbr.Cobra {
    if cbr == nil {
        cbr = libcbr.New()
        cbr.SetVersion(GetVersion())
        cbr.SetLogger(GetLogger)
        cbr.SetViper(GetViper)
    }

    return cbr
}

func GetConfig() libcfg.Config {
    if cfg == nil {
        cfg = libcfg.New()
        cfg.RegisterFuncViper(GetViper)
    }

    return cfg
}

func GetVersion() libver.Version {
    if vrs == nil {
        //vrs = libver.NewVersion(...)
    }

    return vrs
}

func GetLogger() liblog.Logger {
    if log == nil {
        log = liblog.New(cfg.Context())
        _ = log.SetOptions(&liblog.Options{
            DisableStandard:  false,
            DisableStack:     false,
            DisableTimestamp: false,
            EnableTrace:      true,
            TraceFilter:      GetVersion().GetRootPackagePath(),
            DisableColor:     false,
            LogFile:          nil, // TODO
            LogSyslog:        nil, // TODO
        })
        log.SetLevel(liblog.InfoLevel)
		log.SetSPF13Level(liblog.InfoLevel, nil)
    }

    return log
}
```

### The root file in command folder
This file will be the global file for cobra. It will config and init the cobra for yours command.
This file will import your `pkg/common.go` file : 
```go
import(
    "fmt"
    "path"
    
    libcns "github.com/nabbar/golib/console"
    liberr "github.com/nabbar/golib/errors"
    liblog "github.com/nabbar/golib/logger"
    
    compkg "../pkg"
    apipkg "../api"
)
```

And will define some flag vars : 
```go
var (
	// var for config file path to be load by viper
	cfgFile    string
	
	// var for log verbose
	flgVerbose int
	
	// flag for init cobra has error
	iniErr     liberr.Error
)
```

The `InitCommand` function will initialize the cobra / viper package and will be call by the main init function into you `main/init()` function:
```go
func InitCommand() {
	//the initFunc is a function call on init cobra before calling command but after parsing flag / command ...
	compkg.GetCobra().SetFuncInit(initConfig)
	
	// must init after the SetFuncinit function
	compkg.GetCobra().Init()
	
	// add the Config file flag
	compkg.GetLogger().CheckError(liblog.FatalLevel, liblog.DebugLevel, "Add Flag Config File", compkg.GetCobra().SetFlagConfig(true, &cfgFile))
	
	// add the verbose flas
	compkg.GetCobra().SetFlagVerbose(true, &flgVerbose)

	// Add some generic command
	compkg.GetCobra().AddCommandCompletion()
	compkg.GetCobra().AddCommandConfigure("", compkg.GetConfig().DefaultConfig)
	compkg.GetCobra().AddCommandPrintErrorCode(cmdPrintError)
	
	// Add one custom command. The best is to isolate each command into a specific file
	// Each command file, will having a main function to create the cobra command. 
	// Here we will only call this main function to add each command into the main cobra command like this
	compkg.GetCobra().AddCommand(initCustomCommand1())
	
	// Carefully with the `init` function, because you will not manage the order of running each init function.
	// To prevent it, the best is to not having init function, but custom init call in the awaiting order into the `main/init` function
}

// this function will be use in the PrintError command to print first ErrorCode => Package
func cmdPrintError(item, value string) {
	println(fmt.Sprintf("%s : %s", libcns.PadLeft(item, 15, " "), path.Dir(value)))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Redefine the root logger level with the verbose flag
	switch flgVerbose {
	case 0:
		compkg.GetLogger().SetLevel(liblog.ErrorLevel)
	case 1:
		compkg.GetLogger().SetLevel(liblog.WarnLevel)
	case 2:
		compkg.GetLogger().SetLevel(liblog.InfoLevel)
	default:
		compkg.GetLogger().SetLevel(liblog.DebugLevel)
	}

	// Define the config file into viper to load it if is set
	iniErr = compkg.GetViper().SetConfigFile(cfgFile)
	compkg.GetLogger().CheckError(liblog.FatalLevel, liblog.DebugLevel, "define config file", iniErr)

	// try to init viper with config (local file or remote)
	if iniErr = compkg.GetViper().Config(liblog.NilLevel, liblog.NilLevel); iniErr == nil {
		// register the reload config function with the watch FS function 
		compkg.GetViper().SetRemoteReloadFunc(func() {
			_ = compkg.GetLogger().CheckError(liblog.ErrorLevel, liblog.DebugLevel, "config reload", compkg.GetConfig().Reload())
		})
		compkg.GetViper().WatchFS(liblog.InfoLevel)
	}

	// Make the config for this app (cf config package)
	apipkg.SetConfig(compkg.GetLogger().GetLevel(), apirtr.GetHandlers())
}

// This is called by main.main() to parse and run the app.
func Execute() {
    compkg.GetLogger().CheckError(liblog.FatalLevel, liblog.DebugLevel, "RootCmd Executed", compkg.GetCobra().Execute())
}
```

### The main file of your app 
The main file of your app, wil implement the `main/init()` function and the `main/main()` function.

```go
import(
    liberr "github.com/nabbar/golib/errors"
    liblog "github.com/nabbar/golib/logger"

    compkg "./pkg"
    cmdpkg "./cmd"
)

func init() {
	// Configure the golib error package
	liberr.SetModeReturnError(liberr.ErrorReturnStringErrorFull)
	
	// Check the go runtime use to build
	compkg.GetLogger().CheckError(liblog.FatalLevel, liblog.DebugLevel, "Checking go version", compkg.GetVersion().CheckGo("1.16", ">="))
	
	// Call the `cmd/InitCommand` function
	cmdpkg.InitCommand()
}

func main() {
	// run the command Execute function
	cmdpkg.Execute()
}

```
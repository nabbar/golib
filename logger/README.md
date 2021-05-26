# Logger pakcage
Help manage logger. This package does not implement a logger but user `logrus` as logger behind.
This package will simplify call of logger and allow more features like `*log.Logger` wrapper.

## Exmaple of implement

In your file, first add the import of `golib/logger` :
```go
	import liblog "github.com/nabbar/golib/logger"
```

Initialize the logger like this
```go
	log := liblog.New()
	log.SetLevel(liblog.InfoLevel)

	if err := l.SetOptions(context.TODO(), &liblog.Options{
		DisableStandard:  false,
		DisableStack:     false,
		DisableTimestamp: false,
		EnableTrace:      false,
		TraceFilter:      "",
		DisableColor:     false,
		LogFile: []liblog.OptionsFile{
			{
				LogLevel: []string{
					"panic",
					"fatal",
					"error",
					"warning",
					"info",
					"debug",
				},
				Filepath:         "/path/to/my/logfile-with-trace",
				Create:           true,
				CreatePath:       true,
				FileMode:         0644,
				PathMode:         0755,
				DisableStack:     false,
				DisableTimestamp: false,
				EnableTrace:      true,
			},
		},
	}); err != nil {
		panic(err)
	}
```

Calling log like this :
```go
	log.Info("Example log", nil, nil)
    
	// example with a struct name o that you want to expose in log
	// and an list of error : err1, err2 and err3
	log.LogDetails(liblog.InfoLevel, "example of detail log message with simple call", o, []error{err1, err2, err3}, nil, nil)
    
```

Having new log based on last logger but with some pre-defined information
```go
    l := log.Clone(context.TODO())    
    l.SetFields(l.GetFields().Add("one-key", "one-value").Add("lib", "myLib").Add("pkg", "some-package"))
    l.Info("Example log with pre-define information", nil, nil)
    // will print line like : level=info fields.level=Info fields.time="2021-05-25T13:10:02.8033944+02:00" lib=myLib message="Example log with pre-define information" pkg=some-package stack=924 one-key=one-value
    
    // Override the field value on one log like this 
    l.LogDetails(liblog.InfoLevel, "example of detail log message with simple call", o, []error{err1, err2, err3}, liblog.NewFields().Add("lib", "another lib"), nil)
    // will print line like : level=info fields.level=Info fields.time="2021-05-25T13:10:02.8033944+02:00" lib="another lib" message="Example log with pre-define information" pkg=some-package stack=924 one-key=one-value
```


## Implement other logger to this logger

Plug the SPF13 (Cobra / Viper) logger to this logger like this
```go
   log.SetSPF13Level(liblog.InfoLevel, logSpf13)
```

Plug the Hashicorp logger hclog with the logger like this
```go
   log.SetHashicorpHCLog()
```

Or get a hclog logger from the current logger like this
```go
   hlog := log.NewHashicorpHCLog()
```

This call, return a go *log.Logger interface
```go
   l := log.Clone(context.TODO())
   l.SetFields(l.GetFields().Add("one-key", "one-value").Add("lib", "myLib").Add("pkg", "some-package"))
   glog := l.GetStdLogger(liblog.ErrorLevel, log.LstdFlags|log.Lmicroseconds)
```

This call, will connect the default go *log.Logger 
```go
   log.SetStdLogger(liblog.ErrorLevel, log.LstdFlags|log.Lmicroseconds)
```

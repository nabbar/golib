# Logger pakcage
Help manage logger. This package does not implement a logger but user `logrus` as logger behind.
This package will simplify call of logger and allow more features like `*log.Logger` wrapper.

## Exmaple of implement

In your file, first add the import of `golib/logger` :
```go
import . "github.com/nabbar/golib/logger"
```

```go
// Check if the function call will return an error, and if so, will log a fatal (log and os.exit) message
FatalLevel.LogError(GetVersion().CheckGo("1.12", ">="))
```

This call, will disable color, trace, 
```go
FileTrace(false)
DisableColor()
EnableViperLog(true)
```

This call, return a go *log.Logger interface. This example can be found in the golib/httpserver package :
```go
log := GetLogger(ErrorLevel, log.LstdFlags|log.Lmicroseconds, "[http/http2 server '%s']", host)
```

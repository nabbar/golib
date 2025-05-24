# mailPooler Package Documentation

The `mailPooler` package provides a rate-limited SMTP client pooler for sending emails in Go applications. It wraps an SMTP client with configurable limits on the number of emails sent within a given time window, ensuring compliance with provider restrictions and preventing overload.

---

## Features

- Rate-limiting for SMTP email sending (max emails per duration)
- Thread-safe and context-aware
- Custom callback on pool reset
- Cloning and resetting of poolers
- Full SMTP client interface support (send, check, update config, close)
- Monitoring integration

---

## Main Types

### Pooler Interface

Defines the main pooler object, combining rate-limiting and SMTP client operations:

- `Reset() error` — Resets the pooler and underlying SMTP client
- `NewPooler() Pooler` — Clones the pooler with the same configuration
- All methods from the SMTP client interface (send, check, update config, close)

### Config Struct

Configuration for the pooler:

- `Max int` — Maximum number of emails allowed per window
- `Wait time.Duration` — Time window for the rate limit
- `SetFuncCaller(fct FuncCaller)` — Sets a callback function called on pool reset

### Counter

Internal rate-limiting logic:

- `Pool(ctx context.Context) error` — Checks and updates the rate limit before sending
- `Reset() error` — Resets the counter and triggers the callback
- `Clone() Counter` — Clones the counter

---

## Error Handling

- Custom error codes for empty parameters, generic pooler errors, and context cancellation
- Errors are returned with descriptive messages

---

## Example Usage

```go
import (
    "github.com/nabbar/golib/mailPooler"
    "github.com/nabbar/golib/smtp"
    "context"
    "time"
)

cfg := &mailPooler.Config{
    Max:  10,              // max 10 emails
    Wait: 1 * time.Minute, // per minute
}
cfg.SetFuncCaller(func() error {
    // Custom logic on reset (optional)
    return nil
})

smtpClient, _ := smtp.New(/* ... */)
pooler := mailPooler.New(cfg, smtpClient)

err := pooler.Send(context.Background(), "from@example.com", []string{"to@example.com"}, /* data */)
if err != nil {
    // handle error
}
```

---

## Monitoring

- The pooler supports monitoring integration, delegating to the underlying SMTP client.

---

## Notes

- Designed for Go 1.18+.
- All operations are thread-safe.
- Suitable for production and high-concurrency environments.
- The pooler can be cloned and reset as needed.


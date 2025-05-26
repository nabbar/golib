## `ldap` Package Documentation

> **Note:**  
> This package uses an older design and would benefit from a refactor to modern Go idioms and best practices.

---

### Overview

The `ldap` package provides helpers for connecting to, authenticating with, and querying LDAP servers in Go. It supports both plain and TLS/StartTLS connections, user and group lookups, and flexible configuration.

---

### Features

- Connect to LDAP servers with or without TLS/StartTLS
- Bind and authenticate users
- Retrieve user and group information
- Check group membership and list group members
- Customizable search filters and attributes
- Integrated error handling with custom codes
- Logging support for debugging and tracing

---

### Main Types

#### `Config`

Represents the LDAP server configuration.

- `Uri`: Server hostname (FQDN, required)
- `PortLdap`: LDAP port (required, integer)
- `Portldaps`: LDAPS port (optional, integer)
- `Basedn`: Base DN for searches
- `FilterGroup`: Pattern for group search (e.g., `(&(objectClass=groupOfNames)(%s=%s))`)
- `FilterUser`: Pattern for user search (e.g., `(%s=%s)`)

**Validation:**  
Use `Validate()` to check config correctness.

#### `TLSMode`

Enum for connection mode:

- `TLSModeNone`: No TLS
- `TLSModeTLS`: Strict TLS
- `TLSModeStarttls`: StartTLS
- `_TLSModeInit`: Not defined

#### `HelperLDAP`

Main struct for managing LDAP connections and queries.

- `NewLDAP(ctx, config, attributes)`: Create a new helper
- `SetLogger(fct)`: Set a logger function
- `SetCredentials(user, pass)`: Set bind DN and password
- `ForceTLSMode(mode, tlsConfig)`: Force a specific TLS mode and config

---

### Main Methods

- `Check()`: Test connection (no bind)
- `Connect()`: Connect and bind using credentials
- `AuthUser(username, password)`: Test user bind
- `UserInfo(username)`: Get user attributes as a map
- `UserInfoByField(username, field)`: Get user info by a specific field
- `GroupInfo(groupname)`: Get group attributes as a map
- `GroupInfoByField(groupname, field)`: Get group info by a specific field
- `UserMemberOf(username)`: List groups a user belongs to
- `UserIsInGroup(username, groupnames)`: Check if user is in any of the given groups
- `UsersOfGroup(groupname)`: List users in a group
- `ParseEntries(entry)`: Parse DN or attribute string into a map

---

### Error Handling

All errors are wrapped with custom codes for diagnostics, such as:

- `ErrorParamEmpty`
- `ErrorLDAPContext`
- `ErrorLDAPServerConfig`
- `ErrorLDAPServerConnection`
- `ErrorLDAPBind`
- `ErrorLDAPSearch`
- `ErrorLDAPUserNotFound`
- `ErrorLDAPGroupNotFound`
- ...and more

Use `err.Error()` for user-friendly messages and check error codes for diagnostics.

---

### Example Usage

```go
import (
    "context"
    "github.com/nabbar/golib/ldap"
)

cfg := ldap.Config{
    Uri:         "ldap.example.com",
    PortLdap:    389,
    Portldaps:   636,
    Basedn:      "dc=example,dc=com",
    FilterUser:  "(uid=%s)",
    FilterGroup: "(&(objectClass=groupOfNames)(cn=%s))",
}

if err := cfg.Validate(); err != nil {
    // handle config error
}

helper, err := ldap.NewLDAP(context.Background(), &cfg, ldap.GetDefaultAttributes())
if err != nil {
    // handle error
}

helper.SetCredentials("cn=admin,dc=example,dc=com", "password")

if err := helper.Connect(); err != nil {
    // handle connection/bind error
}

userInfo, err := helper.UserInfo("jdoe")
if err != nil {
    // handle user lookup error
}

// ... use userInfo map
helper.Close()
```

---

### Notes

- The package is thread-safe for most operations.
- Designed for Go 1.18+.
- Logging is optional but recommended for debugging.
- The API and code structure are legacy and may not follow modern Go conventions.


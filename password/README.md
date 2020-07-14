# Password pakcage
Help generate random string usable in password....

## Example of implement
In your file, import this lib :
```go
import "github.com/nabbar/golib/password"
```

```go
// generate a password of 24 chars len
newPass := Generate(24)
```

The generated password can having this char :
    - lower letter :    abcdefghijklmnopqrstuvwxyz
    - upper letter :    ABCDEFGHIJKLMNOPQRSTUVWXYZ
    - number :          0123456789
    - special char :    ,;:!?./*%^$&"'(-_)=+~#{[|`\^@]}

An example is available in test/test-password

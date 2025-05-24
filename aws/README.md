# golib/aws

A modular and extensible Go library for AWS S3 and IAM, providing high-level helpers for buckets, objects, users, groups, roles, policies, and advanced uploaders.

---

## Table of Contents

- [Overview](#overview)
- [Installation](#installation)
- [Basic Usage](#basic-usage)
- [Detailed AWS Usage](#detailed-aws-usage)
  - [Bucket](#bucket)
  - [Object](#object)
  - [User](#user)
  - [Group](#group)
  - [Role](#role)
  - [Policy](#policy)
- [Advanced Uploads: Pusher & Multipart](#advanced-uploads-pusher--multipart)
  - [Pusher (Recommended)](#pusher-recommended)
  - [Multipart (Deprecated)](#multipart-deprecated)
- [Error Handling](#error-handling)
- [License](#license)

---

## Overview

This library provides a set of high-level interfaces and helpers to interact with AWS S3 and IAM resources, focusing on ease of use, extensibility, and robust error handling.  
It abstracts the complexity of the AWS SDK for Go, while allowing advanced configuration and custom workflows.

---

## Installation

```sh
go get github.com/nabbar/golib/aws
```

---

## Basic Usage

### Simple S3 Object Upload

```go
package main

import (
    "context"
    "os"
    "github.com/nabbar/golib/aws/pusher"
    "github.com/aws/aws-sdk-go-v2/service/s3"
)

func main() {
    cfg := &pusher.Config{
        FuncGetClientS3: func() *s3.Client { /* return your S3 client */ },
        ObjectS3Options: pusher.ConfigObjectOptions{
            Bucket: aws.String("my-bucket"),
            Key:    aws.String("my-object.txt"),
        },
        PartSize:   5 * 1024 * 1024, // 5MB
        BufferSize: 4096,
        CheckSum:   true,
    }

    ctx := context.Background()
    p, err := pusher.New(ctx, cfg)
    if err != nil { panic(err) }
    defer p.Close()

    file, _ := os.Open("localfile.txt")
    defer file.Close()

    _, err = p.ReadFrom(file)
    if err != nil { panic(err) }

    err = p.Complete()
    if err != nil { panic(err) }
}
```

### Simple Configuration & Usage

```go
package main

import (
    "context"
    "fmt"
	"io"
    "net/http"
	"os"

    "github.com/nabbar/golib/aws"
    "github.com/nabbar/golib/aws/configAws"
)

func main() {
    // Create a new AWS configuration
    cfg := configAws.NewConfig("my-new-bucket", "us-west-2", "your-access-key-id", "your-secret-access-key")

    // Create a new AWS client with default configuration
    cli, err := aws.New(context.Background(), cfg, &http.Client{})
    if err != nil {
        panic(err)
    }

    // Use the client to access S3 buckets, objects, IAM users, etc.
    err := cli.Bucket().Create("my-new-bucket")
    if err != nil {
        panic(err)
    }

    // List all buckets		
    bck, err := cli.Bucket().List()
    if err != nil {
        panic(err)
    }
    for _, bucket := range bck {
        fmt.Println("Bucket:", bucket)
    }

    // Upload an object to the bucket
    file, err := os.Open("path/to/your/file.txt")
    if err != nil {
        panic(err)
    }
    defer file.Close()

    err = cli.Object().Put("file.txt", file)
    if err != nil {
        panic(err)
    }

    fmt.Println("File uploaded successfully!")

    // Download the object
    obj, err := cli.Object().Get("file.txt")
    if err != nil {
        panic(err)
    } else if obj == nil {
        panic("Object not found")
    } else if obj.ContentLength == nil || *obj.ContentLength == 0 {
        panic("Object is empty")
    } else if obj.Body == nil {
        panic("Object is a directory")
    }
  
    defer obj.Body.Close()

    outFile, err := os.Create("downloaded_file.txt")
    if err != nil {
        panic(err)
    }
    defer outFile.Close()

    _, err = io.Copy(outFile, obj.Body)
    if err != nil {
        panic(err)
    }

    fmt.Println("File downloaded successfully!")

    // Clean up: delete the bucket
    err = cli.Bucket().Delete()
    if err != nil {
        panic(err)
    }

    fmt.Println("Bucket deleted successfully!")
}
```
---

## Detailed AWS Usage

### Bucket

Manage S3 buckets: create, delete, list, and set policies.

**Interface:**
```go
type Bucket interface {
    Create(name string, opts ...Option) error
    Delete(name string) error
    List() ([]string, error)
    SetPolicy(name string, policy Policy) error
    // ...
}
```

**Example:**
```go
bck := cli.Bucket().Create("")
lst, err := cli.Bucket().List()
err = cli.Bucket().Delete()
```

#### List Buckets
List all S3 buckets in the account, optionally filtering by prefix:

```go
bck, err := cli.Bucket().List("prefix-")
if err != nil {
    // handle error
}
for _, b := range bck {
    fmt.Println(b)
}
```

#### Walk through S3 buckets
Walk through all S3 buckets, allowing custom processing for each bucket:

```go
err := cli.Bucket().Walk(func(bucket BucketInfo) bool {
    fmt.Printf("Bucket: %s, Created: %s\n", bucket.Name, bucket.CreationDate)
    return true // true to continue walking, false to stop
})
if err != nil {
    // handle error
}
```

---

### Object

Manage S3 objects: upload, download, delete, copy, and metadata.

**Interface:**
```go
type Object interface {
    Put(bucket, key string, data io.Reader, opts ...Option) error
    Get(bucket, key string) (io.ReadCloser, error)
    Delete(bucket, key string) error
    Copy(srcBucket, srcKey, dstBucket, dstKey string) error
    // ...
}
```

**Example:**
```go
err := cli.Object().Put("file.txt", fileReader)
reader, err := cli.Object().Get("file.txt")
err = cli.Object().Delete("file.txt")
```

#### List Object Versions
List all versions of an object in a bucket, or walk through all objects in a bucket.
```go
// Lister toutes les versions d’un objet dans un bucket
versions, err := cli.Object().VersionList("file.txt", "", "")
if err != nil {
    // gérer l’erreur
}
for _, v := range versions {
    fmt.Printf("Key: %s, VersionId: %s, IsLatest: %v\n", v.Key, v.VersionId, v.IsLatest)
}
```

#### Walk through S3 objects
Walk through all objects in a bucket, optionally filtering by prefix and delimiter:

```go
// Walk sur tous les objets d’un bucket
err := cli.Object().WalkPrefix("prefix/", func(info ObjectInfo) bool {
    fmt.Printf("Object: %s, Size: %d\n", info.Key, info.Size)
    return true // true to continue walking, false to stop
})
if err != nil {
    // gérer l’erreur
}
```

---

### User

Manage IAM users: create, delete, credentials, and group membership.

**Interface:**
```go
type User interface {
    Create(name string) error
    Delete(name string) error
    AddToGroup(user, group string) error
    AttachPolicy(user string, policy Policy) error
    // ...
}
```

**Example:**
```go
usr := awsClient.User()
err := usr.Create("alice")
err = usr.AddToGroup("alice", "admins")
```

#### List IAM Users
List all IAM users in the account, optionally filtering by prefix:

```go
usr, err := usr.List("prefix-")
if err != nil {
    // handle error
}
```

#### Walk through IAM users

Walk through all IAM users, allowing custom processing for each user:

```go
err := usr.Walk(func(user UserInfo) bool {
    fmt.Printf("User: %s, ARN: %s\n", user.Name, user.Arn)
    return true // true to continue walking, false to stop
})
if err != nil {
    // handle error
}
```

---

### Group

Manage IAM groups: create, delete, add/remove users, attach policies.

**Interface:**
```go
type Group interface {
    Create(name string) error
    Delete(name string) error
    AddUser(group, user string) error
    RemoveUser(group, user string) error
    AttachPolicy(group string, policy Policy) error
    // ...
}
```

**Example:**
```go
grp := awsClient.Group()
err := grp.Create("admins")
err = grp.AddUser("admins", "alice")
```

#### List IAM Groups
List all IAM groups in the account, optionally filtering by prefix:

```go
grp, err := grp.List("prefix-")
if err != nil {
    // handle error
}
```

#### Walk through IAM groups
Walk through all IAM groups, allowing custom processing for each group:

```go
err := grp.Walk(func(group GroupInfo) bool {
    fmt.Printf("Group: %s, ARN: %s\n", group.Name, group.Arn)
    return true // true to continue walking, false to stop
})
if err != nil {
    // handle error
}
```

---

### Role

Manage IAM roles: create, delete, attach policies.

**Interface:**
```go
type Role interface {
    Create(name string, assumePolicy Policy) error
    Delete(name string) error
    AttachPolicy(role string, policy Policy) error
    // ...
}
```

**Example:**
```go
role := awsClient.Role()
err := role.Create("my-role", assumePolicy)
err = role.AttachPolicy("my-role", policy)
```

#### List IAM Roles
List all IAM roles in the account, optionally filtering by prefix:

```go
rol, err := role.List("prefix-")
if err != nil {
    // handle error
}
```

#### Walk through IAM roles
Walk through all IAM roles, allowing custom processing for each role:

```go
err := role.Walk(func(role RoleInfo) bool {
    fmt.Printf("Role: %s, ARN: %s\n", role.Name, role.Arn)
    return true // true to continue walking, false to stop
})
if err != nil {
    // handle error
}
```

---

### Policy

Manage IAM policies: create, delete, attach/detach.

**Interface:**
```go
type Policy interface {
    Create(name string, document string) error
    Delete(name string) error
    Attach(entity string) error
    Detach(entity string) error
    // ...
}
```

**Example:**
```go
pol := awsClient.Policy()
err := pol.Create("readonly", policyJSON)
err = pol.Attach("my-role")
```

#### List IAM Policies
List all IAM policies in the account, optionally filtering by prefix:

```go
pol, err := pol.List("prefix-")
if err != nil {
    // handle error
}
```

#### Walk through IAM policies
Walk through all IAM policies, allowing custom processing for each policy:

```go
err := pol.Walk(func(policy PolicyInfo) bool {
    fmt.Printf("Policy: %s, ARN: %s\n", policy.Name, policy.Arn)
    return true // true to continue walking, false to stop
})
if err != nil {
    // handle error
}
```

---

## Advanced Uploads: Pusher & Multipart

### Pusher (Recommended)

Modern, high-level API for uploading objects to S3, supporting both single and multipart uploads, with advanced configuration and callback support.

**Interface:**
```go
type Pusher interface {
    io.WriteCloser
    io.ReaderFrom
    Abort() error
    Complete() error
    CopyFromS3(bucket, object, versionId string) error
    // ... (see pusher package for full details)
}
```

**Features:**
- Automatic switch between single and multipart upload
- Custom part size, buffer size, working file path
- Optional SHA256 checksum
- Callbacks for upload progress, completion, and abort
- S3 object copy support (server-side, multipart)
- Thread-safe and context-aware

**Advanced Example:**
```go
cfg := &pusher.Config{
    FuncGetClientS3: myS3ClientFunc,
    FuncCallOnUpload: func(upd pusher.UploadInfo, obj pusher.ObjectInfo, e error) {
        fmt.Printf("Uploaded part %d, etag=%s, error=%v\n", upd.PartNumber, upd.Etag, e)
    },
    FuncCallOnComplete: func(obj pusher.ObjectInfo, e error) {
        fmt.Printf("Upload complete: %+v, error=%v\n", obj, e)
    },
    WorkingPath: "/var/tmp",
    PartSize:    50 * 1024 * 1024,
    BufferSize:  256 * 1024,
    CheckSum:    true,
    ObjectS3Options: pusher.ConfigObjectOptions{
        Bucket: aws.String("my-bucket"),
        Key:    aws.String("my-object"),
        Metadata: map[string]string{"x-amz-meta-custom": "value"},
        ContentType: aws.String("application/octet-stream"),
    },
}
```

---

### Multipart (Deprecated)

**Warning:** The `multipart` package is deprecated.  
Use `pusher` for all new developments.

Legacy API for multipart uploads. Maintained for backward compatibility.

---

## Error Handling

All packages return Go `error` values.  
Common error variables are exported (e.g., `ErrInvalidInstance`, `ErrInvalidClient`, `ErrInvalidResponse`, `ErrInvalidUploadID`, `ErrEmptyContents`, `ErrInvalidChecksum` in `pusher`).  
Always check returned errors and handle them appropriately.

**Example:**
```go
if err := p.Complete(); err != nil {
    if errors.Is(err, pusher.ErrInvalidChecksum) {
        // handle checksum error
    } else {
        // handle other errors
    }
}
```

---

## License

MIT © Nicolas JUHEL

Generated With © Github Copilot.

/*
 *  MIT License
 *
 *  Copyright (c) 2020 Nicolas JUHEL
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in all
 *  copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 *  SOFTWARE.
 *
 */

package aws_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
	"time"

	lbuuid "github.com/hashicorp/go-uuid"
	libaws "github.com/nabbar/golib/aws"
	awscfg "github.com/nabbar/golib/aws/configCustom"
	libhtc "github.com/nabbar/golib/httpcli"
	libpwd "github.com/nabbar/golib/password"
	libsiz "github.com/nabbar/golib/size"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var (
	cli       libaws.AWS
	cfg       libaws.Config
	ctx       context.Context
	cnl       context.CancelFunc
	filename  = "./config.json"
	minioMode = false
	accessKey string
	secretKey string
)

/*
	Using https://onsi.github.io/ginkgo/
	Running with $> ginkgo -cover .
*/

func TestGolibAwsHelper(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Aws Helper Suite")
}

var _ = BeforeSuite(func() {
	var (
		err  error
		name string
		htp  *http.Client
	)

	ctx, cnl = context.WithCancel(context.Background())

	if err = loadConfig(); err != nil {
		var (
			uri = &url.URL{
				Scheme: "http",
				Host:   "localhost:" + strconv.Itoa(GetFreePort()),
			}

			accessKey = libpwd.Generate(20)
			secretKey = libpwd.Generate(64)
		)

		htp = libhtc.GetClient()
		Expect(htp).NotTo(BeNil())

		cfg = awscfg.NewConfig("", accessKey, secretKey, uri, "us-east-1")
		Expect(cfg).NotTo(BeNil())

		cfg.SetRegion("us-east-1")
		err = cfg.RegisterRegionAws(nil)
		Expect(err).NotTo(HaveOccurred())

		minioMode = true

		go LaunchMinio(uri.Host, accessKey, secretKey)

		for WaitMinio(uri.Host) {
			time.Sleep(time.Second)
		}

		time.Sleep(5 * time.Second)
		//println("Minio is waiting on : " + uri.Host)
	}

	cli, err = libaws.New(ctx, cfg, htp)
	Expect(err).NotTo(HaveOccurred())
	Expect(cli).NotTo(BeNil())

	err = cli.ForcePathStyle(ctx, true)
	Expect(err).NotTo(HaveOccurred())

	name, err = lbuuid.GenerateUUID()
	Expect(err).ToNot(HaveOccurred())
	Expect(name).ToNot(BeEmpty())
	// S3 bucket names must be lowercase
	name = "test-" + name
	cli.Config().SetBucketName(name)

	err = cli.Bucket().Check()
	Expect(err).To(HaveOccurred())

	// Create the bucket for tests
	err = cli.Bucket().Create("")
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	cnl()
})

func loadConfig() error {
	var (
		cnfByt []byte
		err    error
	)

	if _, err = os.Stat(filename); err != nil {
		return err
	}

	if cnfByt, err = os.ReadFile(filename); err != nil {
		return err
	}

	if cfg, err = awscfg.NewConfigJsonUnmashal(cnfByt); err != nil {
		return err
	}

	if err = cfg.Validate(); err != nil {
		return err
	}

	return nil
}

func BuildPolicy() string {
	return `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:Get*"],"Resource":["arn:aws:s3:::*/*"]}]}`
}

func BuildRole() string {
	return `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":"sts:AssumeRole","Principal":{"Service":"replication"}}]}`
}

func GetFreePort() int {
	var (
		addr *net.TCPAddr
		lstn *net.TCPListener
		err  error
	)

	if addr, err = net.ResolveTCPAddr("tcp", "localhost:0"); err != nil {
		panic(err)
	}

	if lstn, err = net.ListenTCP("tcp", addr); err != nil {
		panic(err)
	}

	defer func() {
		_ = lstn.Close()
	}()

	return lstn.Addr().(*net.TCPAddr).Port
}

func GetTempFolder() string {
	if tmp, err := os.MkdirTemp("", "minio-data-*"); err != nil {
		panic(err)
	} else {
		return tmp
	}
}

func DelTempFolder(folder string) {
	if err := os.RemoveAll(folder); err != nil {
		panic(err)
	}
}

func LaunchMinio(host, accessKey, secretKey string) {
	os.Setenv("MINIO_ACCESS_KEY", accessKey)
	os.Setenv("MINIO_SECRET_KEY", secretKey)

	tmp := GetTempFolder()
	defer DelTempFolder(tmp)

	if _, minio, _, ok := runtime.Caller(0); ok {
		if err := exec.CommandContext(ctx, filepath.Join(filepath.Dir(minio), "minio"), "server", "--address", host, tmp).Run(); err != nil {
			if ctx.Err() != nil {
				return
			}
			panic(err)
		}
	} else {
		//nolint #goerr113
		panic(fmt.Errorf("minio execution file not found"))
	}

	//minio.Main([]string{"minio", "server", "--address", host, tmp})
}

func WaitMinio(host string) bool {
	conn, err := net.DialTimeout("tcp", host, 10*time.Second)

	defer func() {
		if conn != nil {
			_ = conn.Close()
		}
	}()

	time.Sleep(5 * time.Second)

	return err == nil
}

type randReader struct {
	data []byte
	pos  int64
}

func (r *randReader) Read(p []byte) (n int, err error) {
	if r.pos >= int64(len(r.data)) {
		return 0, io.EOF
	}

	n = copy(p, r.data[r.pos:])
	r.pos += int64(n)
	return n, nil
}

func (r *randReader) Seek(offset int64, whence int) (int64, error) {
	var newPos int64

	switch whence {
	case io.SeekStart:
		newPos = offset
	case io.SeekCurrent:
		newPos = r.pos + offset
	case io.SeekEnd:
		newPos = int64(len(r.data)) + offset
	default:
		return 0, fmt.Errorf("invalid whence")
	}

	if newPos < 0 {
		return 0, fmt.Errorf("negative position")
	}

	r.pos = newPos
	return newPos, nil
}

func randContent(size libsiz.Size) io.ReadSeeker {
	data := make([]byte, size.Int64())
	_, _ = rand.Read(data)

	return &randReader{
		data: data,
		pos:  0,
	}
}

// GenerateUniqueName generates a unique name for test resources
func GenerateUniqueName(prefix string) string {
	name, err := lbuuid.GenerateUUID()
	if err != nil {
		panic(err)
	}
	if prefix != "" {
		return prefix + "-" + name
	}
	return name
}

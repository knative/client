package commands_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCommands(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Commands Suite")
}

var OldStdout *os.File
var Stdout *os.File
var Output string
var ReadFile, WriteFile *os.File

var _ = BeforeSuite(func() {
	captureStdout()
})

var _ = AfterSuite(func() {
	releaseStdout()
})

// ReadStdout collects the current content of os.Stdout
// into Output global
func ReadStdout() string {
	outC := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, ReadFile)
		outC <- buf.String()
	}()
	WriteFile.Close()
	Output = <-outC
	return Output
}

// Private

func captureStdout() {
	OldStdout = os.Stdout
	var err error
	ReadFile, WriteFile, err = os.Pipe()
	Expect(err).NotTo(HaveOccurred())
	Stdout = WriteFile
	os.Stdout = WriteFile
}

func releaseStdout() {
	Output = ReadStdout()
	os.Stdout = OldStdout
}

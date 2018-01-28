package rconfig_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestRconfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Rconfig Suite")
}

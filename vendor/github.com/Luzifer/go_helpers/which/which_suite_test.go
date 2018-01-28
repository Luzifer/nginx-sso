package which_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestWhich(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Which Suite")
}

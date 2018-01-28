package position_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestPosition(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Position Suite")
}

package rconfig_test

import (
	"os"

	. "github.com/Luzifer/rconfig"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing os.Args", func() {
	type t struct {
		A string `default:"a" flag:"a"`
	}

	var (
		err error
		cfg t
	)

	JustBeforeEach(func() {
		err = Parse(&cfg)
	})

	Context("With only valid arguments", func() {

		BeforeEach(func() {
			cfg = t{}
			os.Args = []string{"--a=bar"}
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have the expected values", func() {
			Expect(cfg.A).To(Equal("bar"))
		})

	})

})

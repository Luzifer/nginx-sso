package rconfig

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing sub-structs", func() {
	type t struct {
		Test string `default:"blubb"`
		Sub  struct {
			Test string `default:"Hallo"`
		}
	}

	var (
		err  error
		args []string
		cfg  t
	)

	BeforeEach(func() {
		cfg = t{}
		args = []string{}
	})

	JustBeforeEach(func() {
		err = parse(&cfg, args)
	})

	It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
	It("should have the expected values", func() {
		Expect(cfg.Test).To(Equal("blubb"))
		Expect(cfg.Sub.Test).To(Equal("Hallo"))
	})
})

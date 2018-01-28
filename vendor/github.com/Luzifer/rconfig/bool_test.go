package rconfig

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing bool parsing", func() {
	type t struct {
		Test1 bool `default:"true"`
		Test2 bool `default:"false" flag:"test2"`
		Test3 bool `default:"true" flag:"test3,t"`
		Test4 bool `flag:"test4"`
	}

	var (
		err  error
		args []string
		cfg  t
	)

	BeforeEach(func() {
		cfg = t{}
		args = []string{
			"--test2",
			"-t",
		}
	})

	JustBeforeEach(func() {
		err = parse(&cfg, args)
	})

	It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
	It("should have the expected values", func() {
		Expect(cfg.Test1).To(Equal(true))
		Expect(cfg.Test2).To(Equal(true))
		Expect(cfg.Test3).To(Equal(true))
		Expect(cfg.Test4).To(Equal(false))
	})
})

var _ = Describe("Testing to set bool from ENV with default", func() {
	type t struct {
		Test1 bool `default:"true" env:"TEST1"`
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
		os.Unsetenv("TEST1")
		err = parse(&cfg, args)
	})

	It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
	It("should have the expected values", func() {
		Expect(cfg.Test1).To(Equal(true))
	})
})

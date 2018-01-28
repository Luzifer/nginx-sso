package rconfig

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing float parsing", func() {
	type t struct {
		Test32  float32 `flag:"float32"`
		Test32P float32 `flag:"float32p,3"`
		Test64  float64 `flag:"float64"`
		Test64P float64 `flag:"float64p,6"`
		TestDef float32 `default:"66.256"`
	}

	var (
		err  error
		args []string
		cfg  t
	)

	BeforeEach(func() {
		cfg = t{}
		args = []string{
			"--float32=5.5", "-3", "6.6",
			"--float64=7.7", "-6", "8.8",
		}
	})

	JustBeforeEach(func() {
		err = parse(&cfg, args)
	})

	It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
	It("should have the expected values", func() {
		Expect(cfg.Test32).To(Equal(float32(5.5)))
		Expect(cfg.Test32P).To(Equal(float32(6.6)))
		Expect(cfg.Test64).To(Equal(float64(7.7)))
		Expect(cfg.Test64P).To(Equal(float64(8.8)))

		Expect(cfg.TestDef).To(Equal(float32(66.256)))
	})
})

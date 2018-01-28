package rconfig

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing int parsing", func() {
	type t struct {
		Test    int   `flag:"int"`
		TestP   int   `flag:"intp,i"`
		Test8   int8  `flag:"int8"`
		Test8P  int8  `flag:"int8p,8"`
		Test32  int32 `flag:"int32"`
		Test32P int32 `flag:"int32p,3"`
		Test64  int64 `flag:"int64"`
		Test64P int64 `flag:"int64p,6"`
		TestDef int8  `default:"66"`
	}

	var (
		err  error
		args []string
		cfg  t
	)

	BeforeEach(func() {
		cfg = t{}
		args = []string{
			"--int=1", "-i", "2",
			"--int8=3", "-8", "4",
			"--int32=5", "-3", "6",
			"--int64=7", "-6", "8",
		}
	})

	JustBeforeEach(func() {
		err = parse(&cfg, args)
	})

	It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
	It("should have the expected values", func() {
		Expect(cfg.Test).To(Equal(1))
		Expect(cfg.TestP).To(Equal(2))
		Expect(cfg.Test8).To(Equal(int8(3)))
		Expect(cfg.Test8P).To(Equal(int8(4)))
		Expect(cfg.Test32).To(Equal(int32(5)))
		Expect(cfg.Test32P).To(Equal(int32(6)))
		Expect(cfg.Test64).To(Equal(int64(7)))
		Expect(cfg.Test64P).To(Equal(int64(8)))

		Expect(cfg.TestDef).To(Equal(int8(66)))
	})
})

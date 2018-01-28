package rconfig

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing uint parsing", func() {
	type t struct {
		Test    uint   `flag:"int"`
		TestP   uint   `flag:"intp,i"`
		Test8   uint8  `flag:"int8"`
		Test8P  uint8  `flag:"int8p,8"`
		Test16  uint16 `flag:"int16"`
		Test16P uint16 `flag:"int16p,1"`
		Test32  uint32 `flag:"int32"`
		Test32P uint32 `flag:"int32p,3"`
		Test64  uint64 `flag:"int64"`
		Test64P uint64 `flag:"int64p,6"`
		TestDef uint8  `default:"66"`
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
			"--int16=9", "-1", "10",
		}
	})

	JustBeforeEach(func() {
		err = parse(&cfg, args)
	})

	It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
	It("should have the expected values", func() {
		Expect(cfg.Test).To(Equal(uint(1)))
		Expect(cfg.TestP).To(Equal(uint(2)))
		Expect(cfg.Test8).To(Equal(uint8(3)))
		Expect(cfg.Test8P).To(Equal(uint8(4)))
		Expect(cfg.Test32).To(Equal(uint32(5)))
		Expect(cfg.Test32P).To(Equal(uint32(6)))
		Expect(cfg.Test64).To(Equal(uint64(7)))
		Expect(cfg.Test64P).To(Equal(uint64(8)))
		Expect(cfg.Test16).To(Equal(uint16(9)))
		Expect(cfg.Test16P).To(Equal(uint16(10)))

		Expect(cfg.TestDef).To(Equal(uint8(66)))
	})
})

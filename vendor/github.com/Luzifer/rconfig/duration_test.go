package rconfig

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Duration", func() {
	type t struct {
		Test    time.Duration `flag:"duration"`
		TestS   time.Duration `flag:"other-duration,o"`
		TestDef time.Duration `default:"30h"`
	}

	var (
		err  error
		args []string
		cfg  t
	)

	BeforeEach(func() {
		cfg = t{}
		args = []string{
			"--duration=23s", "-o", "45m",
		}
	})

	JustBeforeEach(func() {
		err = parse(&cfg, args)
	})

	It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
	It("should have the expected values", func() {
		Expect(cfg.Test).To(Equal(23 * time.Second))
		Expect(cfg.TestS).To(Equal(45 * time.Minute))

		Expect(cfg.TestDef).To(Equal(30 * time.Hour))
	})
})

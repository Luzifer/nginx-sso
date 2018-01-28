package rconfig

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing slices", func() {
	type t struct {
		Int     []int    `default:"1,2,3" flag:"int"`
		String  []string `default:"a,b,c" flag:"string"`
		IntP    []int    `default:"1,2,3" flag:"intp,i"`
		StringP []string `default:"a,b,c" flag:"stringp,s"`
	}

	var (
		err  error
		args []string
		cfg  t
	)

	BeforeEach(func() {
		cfg = t{}
		args = []string{
			"--int=4,5", "-s", "hallo,welt",
		}
	})

	JustBeforeEach(func() {
		err = parse(&cfg, args)
	})

	It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
	It("should have the expected values for int-slice", func() {
		Expect(len(cfg.Int)).To(Equal(2))
		Expect(cfg.Int).To(Equal([]int{4, 5}))
		Expect(cfg.Int).NotTo(Equal([]int{5, 4}))
	})
	It("should have the expected values for int-shorthand-slice", func() {
		Expect(len(cfg.IntP)).To(Equal(3))
		Expect(cfg.IntP).To(Equal([]int{1, 2, 3}))
	})
	It("should have the expected values for string-slice", func() {
		Expect(len(cfg.String)).To(Equal(3))
		Expect(cfg.String).To(Equal([]string{"a", "b", "c"}))
	})
	It("should have the expected values for string-shorthand-slice", func() {
		Expect(len(cfg.StringP)).To(Equal(2))
		Expect(cfg.StringP).To(Equal([]string{"hallo", "welt"}))
	})
})

package rconfig

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing errors", func() {

	It("should not accept string as int", func() {
		Expect(parse(&struct {
			A int `default:"a"`
		}{}, []string{})).To(HaveOccurred())
	})

	It("should not accept string as float", func() {
		Expect(parse(&struct {
			A float32 `default:"a"`
		}{}, []string{})).To(HaveOccurred())
	})

	It("should not accept string as uint", func() {
		Expect(parse(&struct {
			A uint `default:"a"`
		}{}, []string{})).To(HaveOccurred())
	})

	It("should not accept string as uint in sub-struct", func() {
		Expect(parse(&struct {
			B struct {
				A uint `default:"a"`
			}
		}{}, []string{})).To(HaveOccurred())
	})

	It("should not accept string slice as int slice", func() {
		Expect(parse(&struct {
			A []int `default:"a,bn"`
		}{}, []string{})).To(HaveOccurred())
	})

	It("should not accept variables not being pointers", func() {
		cfg := struct {
			A string `default:"a"`
		}{}

		Expect(parse(cfg, []string{})).To(HaveOccurred())
	})

	It("should not accept variables not being pointers to structs", func() {
		cfg := "test"

		Expect(parse(cfg, []string{})).To(HaveOccurred())
	})

})

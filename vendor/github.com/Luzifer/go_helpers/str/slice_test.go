package str_test

import (
	. "github.com/Luzifer/go_helpers/str"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Slice", func() {

	Context("AppendIfMissing", func() {
		var sl = []string{
			"test1",
			"test2",
			"test3",
		}

		It("should not append existing elements", func() {
			Expect(len(AppendIfMissing(sl, "test1"))).To(Equal(3))
			Expect(len(AppendIfMissing(sl, "test2"))).To(Equal(3))
			Expect(len(AppendIfMissing(sl, "test3"))).To(Equal(3))
		})

		It("should append not existing elements", func() {
			Expect(len(AppendIfMissing(sl, "test4"))).To(Equal(4))
			Expect(len(AppendIfMissing(sl, "test5"))).To(Equal(4))
			Expect(len(AppendIfMissing(sl, "test6"))).To(Equal(4))
		})
	})

	Context("StringInSlice", func() {
		var sl = []string{
			"test1",
			"test2",
			"test3",
		}

		It("should find elements of slice", func() {
			Expect(StringInSlice("test1", sl)).To(Equal(true))
			Expect(StringInSlice("test2", sl)).To(Equal(true))
			Expect(StringInSlice("test3", sl)).To(Equal(true))
		})

		It("should not find elements not in slice", func() {
			Expect(StringInSlice("test4", sl)).To(Equal(false))
			Expect(StringInSlice("test5", sl)).To(Equal(false))
			Expect(StringInSlice("test6", sl)).To(Equal(false))
		})
	})

})

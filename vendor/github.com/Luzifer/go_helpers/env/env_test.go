package env_test

import (
	"sort"

	. "github.com/Luzifer/go_helpers/env"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Env", func() {

	Context("ListToMap", func() {
		var (
			list = []string{
				"FIRST_KEY=firstvalue",
				"SECOND_KEY=secondvalue",
				"WEIRD=",
				"",
			}
			emap = map[string]string{
				"FIRST_KEY":  "firstvalue",
				"SECOND_KEY": "secondvalue",
				"WEIRD":      "",
			}
		)

		It("should convert the list in the expected way", func() {
			Expect(ListToMap(list)).To(Equal(emap))
		})
	})

	Context("MapToList", func() {
		var (
			list = []string{
				"FIRST_KEY=firstvalue",
				"SECOND_KEY=secondvalue",
				"WEIRD=",
			}
			emap = map[string]string{
				"FIRST_KEY":  "firstvalue",
				"SECOND_KEY": "secondvalue",
				"WEIRD":      "",
			}
		)

		It("should convert the map in the expected way", func() {
			l := MapToList(emap)
			sort.Strings(l) // Workaround: The test needs the elements to be in same order
			Expect(l).To(Equal(list))
		})
	})

})

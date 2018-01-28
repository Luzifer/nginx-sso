package float_test

import (
	"math"

	. "github.com/Luzifer/go_helpers/float"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Round", func() {

	It("should match the example table of IEEE 754 rules", func() {
		Expect(Round(11.5)).To(Equal(12.0))
		Expect(Round(12.5)).To(Equal(13.0))
		Expect(Round(-11.5)).To(Equal(-12.0))
		Expect(Round(-12.5)).To(Equal(-13.0))
	})

	It("should have correct rounding for numbers near 0.5", func() {
		Expect(Round(0.499999999997)).To(Equal(0.0))
		Expect(Round(-0.499999999997)).To(Equal(0.0))
	})

	It("should be able to handle +/-Inf", func() {
		Expect(Round(math.Inf(1))).To(Equal(math.Inf(1)))
		Expect(Round(math.Inf(-1))).To(Equal(math.Inf(-1)))
	})

	It("should be able to handle NaN", func() {
		Expect(math.IsNaN(Round(math.NaN()))).To(Equal(true))
	})

})

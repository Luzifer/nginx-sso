package which_test

import (
	. "github.com/Luzifer/go_helpers/which"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Which", func() {
	var (
		result string
		err    error
		found  bool
	)

	Context("With a file available on linux systems", func() {
		BeforeEach(func() {
			found, err = FindInDirectory("bash", "/bin")
		})

		It("should not have errored", func() {
			Expect(err).NotTo(HaveOccurred())
		})
		It("should have found the binary at /bin/bash", func() {
			Expect(found).To(BeTrue())
		})
	})

	Context("Searching bash on the system", func() {
		BeforeEach(func() {
			result, err = FindInPath("bash")
		})

		It("should not have errored", func() {
			Expect(err).NotTo(HaveOccurred())
		})
		It("should have a result", func() {
			Expect(len(result)).NotTo(Equal(0))
		})
	})

	Context("Searching a non existent file", func() {
		BeforeEach(func() {
			result, err = FindInPath("dfqoiwurgtqi3uegrds")
		})

		It("should have errored", func() {
			Expect(err).To(Equal(ErrBinaryNotFound))
		})
	})

	Context("Searching an empty file", func() {
		BeforeEach(func() {
			result, err = FindInPath("")
		})

		It("should have errored", func() {
			Expect(err).To(Equal(ErrNoSearchSpecified))
		})
	})

})

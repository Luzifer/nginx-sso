package rconfig

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Precedence", func() {

	type t struct {
		A int `default:"1" vardefault:"a" env:"a" flag:"avar,a" description:"a"`
	}

	var (
		err         error
		cfg         t
		args        []string
		vardefaults map[string]string
	)

	JustBeforeEach(func() {
		cfg = t{}
		SetVariableDefaults(vardefaults)
		err = parse(&cfg, args)
	})

	Context("Provided: Flag, Env, Default, VarDefault", func() {
		BeforeEach(func() {
			args = []string{"-a", "5"}
			os.Setenv("a", "8")
			vardefaults = map[string]string{
				"a": "3",
			}
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have used the flag value", func() {
			Expect(cfg.A).To(Equal(5))
		})
	})

	Context("Provided: Env, Default, VarDefault", func() {
		BeforeEach(func() {
			args = []string{}
			os.Setenv("a", "8")
			vardefaults = map[string]string{
				"a": "3",
			}
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have used the env value", func() {
			Expect(cfg.A).To(Equal(8))
		})
	})

	Context("Provided: Default, VarDefault", func() {
		BeforeEach(func() {
			args = []string{}
			os.Unsetenv("a")
			vardefaults = map[string]string{
				"a": "3",
			}
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have used the vardefault value", func() {
			Expect(cfg.A).To(Equal(3))
		})
	})

	Context("Provided: Default", func() {
		BeforeEach(func() {
			args = []string{}
			os.Unsetenv("a")
			vardefaults = map[string]string{}
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have used the default value", func() {
			Expect(cfg.A).To(Equal(1))
		})
	})

})

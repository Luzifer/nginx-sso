package rconfig

import (
	"io/ioutil"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Testing variable defaults", func() {

	type t struct {
		MySecretValue string `default:"secret" env:"foo" vardefault:"my_secret_value"`
		MyUsername    string `default:"luzifer" vardefault:"username"`
		SomeVar       string `flag:"var" description:"some variable"`
		IntVar        int64  `vardefault:"int_var" default:"23"`
	}

	var (
		err         error
		cfg         t
		args        = []string{}
		vardefaults = map[string]string{
			"my_secret_value": "veryverysecretkey",
			"unkownkey":       "hi there",
			"int_var":         "42",
		}
	)

	BeforeEach(func() {
		cfg = t{}
	})

	JustBeforeEach(func() {
		err = parse(&cfg, args)
	})

	Context("With manually provided variables", func() {
		BeforeEach(func() {
			SetVariableDefaults(vardefaults)
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have the expected values", func() {
			Expect(cfg.IntVar).To(Equal(int64(42)))
			Expect(cfg.MySecretValue).To(Equal("veryverysecretkey"))
			Expect(cfg.MyUsername).To(Equal("luzifer"))
			Expect(cfg.SomeVar).To(Equal(""))
		})
	})

	Context("With defaults from YAML data", func() {
		BeforeEach(func() {
			yamlData := []byte("---\nmy_secret_value: veryverysecretkey\nunknownkey: hi there\nint_var: 42\n")
			SetVariableDefaults(VarDefaultsFromYAML(yamlData))
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have the expected values", func() {
			Expect(cfg.IntVar).To(Equal(int64(42)))
			Expect(cfg.MySecretValue).To(Equal("veryverysecretkey"))
			Expect(cfg.MyUsername).To(Equal("luzifer"))
			Expect(cfg.SomeVar).To(Equal(""))
		})
	})

	Context("With defaults from YAML file", func() {
		var tmp *os.File

		BeforeEach(func() {
			tmp, _ = ioutil.TempFile("", "")
			yamlData := "---\nmy_secret_value: veryverysecretkey\nunknownkey: hi there\nint_var: 42\n"
			tmp.WriteString(yamlData)
			SetVariableDefaults(VarDefaultsFromYAMLFile(tmp.Name()))
		})

		AfterEach(func() {
			tmp.Close()
			os.Remove(tmp.Name())
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have the expected values", func() {
			Expect(cfg.IntVar).To(Equal(int64(42)))
			Expect(cfg.MySecretValue).To(Equal("veryverysecretkey"))
			Expect(cfg.MyUsername).To(Equal("luzifer"))
			Expect(cfg.SomeVar).To(Equal(""))
		})
	})

	Context("With defaults from invalid YAML data", func() {
		BeforeEach(func() {
			yamlData := []byte("---\nmy_secret_value = veryverysecretkey\nunknownkey = hi there\nint_var = 42\n")
			SetVariableDefaults(VarDefaultsFromYAML(yamlData))
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have the expected values", func() {
			Expect(cfg.IntVar).To(Equal(int64(23)))
			Expect(cfg.MySecretValue).To(Equal("secret"))
			Expect(cfg.MyUsername).To(Equal("luzifer"))
			Expect(cfg.SomeVar).To(Equal(""))
		})
	})

	Context("With defaults from non existent YAML file", func() {
		BeforeEach(func() {
			file := "/tmp/this_file_should_not_exist_146e26723r"
			SetVariableDefaults(VarDefaultsFromYAMLFile(file))
		})

		It("should not have errored", func() { Expect(err).NotTo(HaveOccurred()) })
		It("should have the expected values", func() {
			Expect(cfg.IntVar).To(Equal(int64(23)))
			Expect(cfg.MySecretValue).To(Equal("secret"))
			Expect(cfg.MyUsername).To(Equal("luzifer"))
			Expect(cfg.SomeVar).To(Equal(""))
		})
	})

})

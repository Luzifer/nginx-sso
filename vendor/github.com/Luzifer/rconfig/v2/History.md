# 2.2.1 / 2019-02-04

  * Add go module information

# 2.2.0 / 2018-09-18

  * Add support for time.Time flags

# 2.1.0 / 2018-08-02

  * Add AutoEnv feature

# 2.0.0 / 2018-08-02

  * Breaking: Ensure an empty default string does not yield a slice with 1 element  
    Though this is a just a tiny change it does change the default behaviour, so I'm marking this as a breaking change. You should ensure your code is fine with the changes.

# 1.2.0 / 2017-06-19

  * Add ParseAndValidate method

# 1.1.0 / 2016-06-28

  * Support time.Duration config parameters
  * Added goreportcard badge
  * Added testcase for using bool with ENV and default

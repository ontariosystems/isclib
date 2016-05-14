package log

// StringArrayFlags is a type that can be used to allow for an array of args to be passed
// into a program.
// @example:
//   var globalLogFields log.StringArrayFlags
//   flag.Var(&initValues.globalLogFields, "global.log.field", "[]key:value")
//   flag.Parse()
//   log.MoreGlobalFlags(initValues.globalLogFields)
type StringArrayFlags []string

func (i *StringArrayFlags) String() string {
	return ""
}

// Set will append the string to the list
func (i *StringArrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

package cmdline

type orderedCommandLineMap struct {
	values map[string]*command
	order  []string
}

func newOrderedCommandLineMap() *orderedCommandLineMap {
	return &orderedCommandLineMap{
		values: make(map[string]*command),
		order:  make([]string, 0),
	}
}

func (m *orderedCommandLineMap) add(name string, cmd *command) {
	m.values[name] = cmd
	m.order = append(m.order, name)
}

type orderedGlobalOptionMap struct {
	values map[string]*globalOption
	order  []string
}

func newOrderedGlobalOptionMap() *orderedGlobalOptionMap {
	return &orderedGlobalOptionMap{
		values: make(map[string]*globalOption),
		order:  make([]string, 0),
	}
}

func (m *orderedGlobalOptionMap) add(name string, opt *globalOption) {
	m.values[name] = opt
	m.order = append(m.order, name)
}

type orderedArgSpecMap struct {
	values map[string]*argSpec
	order  []string
}

func newOrderedArgSpecMap() *orderedArgSpecMap {
	return &orderedArgSpecMap{
		values: make(map[string]*argSpec),
		order:  make([]string, 0),
	}
}

func (m *orderedArgSpecMap) add(name string, as *argSpec) {
	m.values[name] = as
	m.order = append(m.order, name)
}

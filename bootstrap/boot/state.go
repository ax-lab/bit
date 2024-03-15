package boot

type State struct {
	sourceMap
	nodeMap
	errorList
	bindingMap
}

func (st *State) CheckDone() {
	if err := st.nodeMap.CheckDone(); err != nil {
		st.AddError(err)
	}
}

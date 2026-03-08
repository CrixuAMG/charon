package tui

type viewState int

const (
	stateList viewState = iota
	stateAdd
	stateEdit
	stateDelete
	stateTasksetList
	stateTasksetAdd
	stateTasksetEdit
	stateTasksetDelete
	stateInput       // prompting for task template parameters
	stateBulkConfirm // confirming a bulk operation
)

type bulkAction int

const (
	bulkDelete  bulkAction = iota
	bulkArchive            // archives or unarchives all selected
)

type sortMode int

const (
	sortByName sortMode = iota
	sortByRecent
	sortByFrequent
	sortByCustom
)

type filterMode int

const (
	filterNone filterMode = iota
	filterDocker
	filterLocal
)

func (s sortMode) String() string {
	switch s {
	case sortByName:
		return "Name"
	case sortByRecent:
		return "Recent"
	case sortByFrequent:
		return "Frequent"
	case sortByCustom:
		return "Custom"
	default:
		return "Unknown"
	}
}

func (f filterMode) String() string {
	switch f {
	case filterNone:
		return "All"
	case filterDocker:
		return "Docker"
	case filterLocal:
		return "Local"
	default:
		return "Unknown"
	}
}

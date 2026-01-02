package execution

type Terminal interface {
	LaunchLocalTab(
		socket string,
		title string,
		path string,
	) (int, error)

	LaunchDockerTab(
		socket string,
		title string,
		container string,
		workdir string,
	) (int, error)

	SendText(
		socket string,
		windowID int,
		text string,
	) error
}

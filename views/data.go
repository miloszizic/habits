package views

const (
	AlertLvlError   = "red"
	AlertLvlSuccess = "green"
	AlertLvlNeutral = "blue"

	// AlertMsgGeneric is displayed when any random error
	// is encountered by our backend.
	AlertMsgGeneric = "Something went wrong. Please try again, and contact us if the problem persists."
)

// Alert is used to render Bootstrap Alert messages in templates
type Alert struct {
	Color   string
	Message string
}

// Data is the top level structure that views expect data
// to come in.
type Data struct {
	Alert *Alert
}

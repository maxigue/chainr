package notifier

import "fmt"

type LogNotifier struct{}

func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

type LogFormat string

const (
	StartFormat   = "\033[1;34m%s\033[0m\n"
	SuccessFormat = "\033[1;32m%s\033[0m\n"
	FailureFormat = "\033[1;31m%s\033[0m\n"
	DefaultFormat = "%s\n"
)

func (n *LogNotifier) Dispatch(event Event) error {
	format := getLogFormat(event.Type)
	text := fmt.Sprintf("%s\n%s", event.Title, event.Message)
	fmt.Printf(format, text)
	return nil
}

func getLogFormat(t string) string {
	switch t {
	case "START":
		return StartFormat
	case "SUCCESS":
		return SuccessFormat
	case "FAILURE":
		return FailureFormat
	}

	return DefaultFormat
}

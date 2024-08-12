package shortener

import "fmt"

type InvalidURLError struct {
	Reason string
}

func (e InvalidURLError) Error() string {
	return fmt.Sprintf("invalid URL: %s", e.Reason)
}

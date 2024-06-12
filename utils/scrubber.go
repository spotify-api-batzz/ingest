package utils

import (
	"net/url"
	"strings"
)

type Scrubber interface {
	Scrub(string) string
	Is(string) bool
}

type QueryParamScrubber struct {
	keysToScrub []string
}

func (q *QueryParamScrubber) Is(value string) bool {
	chunks := strings.Split(value, "&")
	if len(chunks) == 1 {
		return false
	}

	for _, chunk := range chunks {
		if !strings.Contains(chunk, "=") {
			return false
		}
	}

	return true
}

func (q *QueryParamScrubber) Scrub(value string) string {
	query, err := url.ParseQuery(value)
	if err != nil {
		return value
	}

	scrubbedValue := value
	for _, key := range q.keysToScrub {
		if !query.Has(key) {
			continue
		}
		toBeScrubbed := query.Get(key)
		scrubbed := convertToAsterisks(toBeScrubbed)
		scrubbedValue = strings.ReplaceAll(scrubbedValue, toBeScrubbed, scrubbed)
	}

	return scrubbedValue
}

func ScrubSensitiveData(val string, keysToScrub []string) string {
	scrubbers := []Scrubber{&QueryParamScrubber{keysToScrub: keysToScrub}}
	for _, scrubber := range scrubbers {
		if scrubber.Is(val) {
			return scrubber.Scrub(val)
		}
	}

	return val
}

func convertToAsterisks(input string) string {
	length := len(input)
	output := make([]byte, length)
	for i := range output {
		output[i] = '*'
	}
	return string(output)
}

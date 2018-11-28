package handlers

import (
	"regexp"
	"testing"
)

func TestCookieGenerator(t *testing.T) {
	twentyCookieCharacters := regexp.MustCompile("^[\\w+_]{20}$")
	token := randomToken()
	if !twentyCookieCharacters.MatchString(token) {
		t.Error("Expected /^[\\w+-]{20}$/, got: " + token)
	}
}

package docker

import (
	"log"
	"regexp"
	"strings"
)

type AuthenticationInfo struct {
	Realm   string
	Service string
	Scope   string
}

// ParseWwwAuthentication parse the authentication data received in a WWW-Authenticate header.
// The expected format input is: realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
//
// See also: [WWW-Authenticate doc].
//
// [WWW-Authenticate doc]: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/WWW-Authenticate
func ParseWwwAuthentication(wwwAuthHeader string) (authInfo AuthenticationInfo) {

	headerValue := getLastElement(strings.Split(wwwAuthHeader, " "))
	challenges := challengesExtractor(strings.Split(headerValue, ","))

	return AuthenticationInfo{
		Realm:   challenges["realm"],
		Service: challenges["service"],
		Scope:   challenges["scope"],
	}
}

func getLastElement(stringArray []string) string {
	return stringArray[len(stringArray)-1]
}

// expected input: realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
func challengesExtractor(headerValue []string) map[string]string {
	challenges := make(map[string]string)
	var challengeRegex = regexp.MustCompile(`(.+)="(.+)"`)

	for _, value := range headerValue {
		match := challengeRegex.FindStringSubmatch(value)
		log.Printf("Challenge added: %s\n", match[0])
		challenges[match[1]] = match[2]
	}

	return challenges
}

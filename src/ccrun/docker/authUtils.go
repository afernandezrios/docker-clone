package docker

import (
	"fmt"
	"regexp"
	"strings"
)

type AuthenticationInfo struct {
	Realm   string
	Service string
	Scope   string
}

// realm="https://auth.docker.io/token",service="registry.docker.io",scope="repository:alpine/git:pull"
func ParseWwwAuthentication(wwwAuthHeader string) (authInfo AuthenticationInfo) {

	headerParams := getLastElement(strings.Split(wwwAuthHeader, " "))
	params := paramsExtractor(strings.Split(headerParams, ","))

	return AuthenticationInfo{
		Realm:   params["realm"],
		Service: params["service"],
		Scope:   params["scope"],
	}
}

func getLastElement(stringArray []string) string {
	return stringArray[len(stringArray)-1]
}

func paramsExtractor(headerParams []string) map[string]string {
	params := make(map[string]string)
	var paramRegex = regexp.MustCompile(`(.+)="(.+)"`)

	for _, value := range headerParams {
		match := paramRegex.FindStringSubmatch(value)
		if len(match) != 3 {
			fmt.Printf("Header params not expected: %s\n", value)
		} else {
			params[match[1]] = match[2]
			fmt.Printf("params added: %s = %s\n", match[0], match[1])
		}
	}

	return params
}

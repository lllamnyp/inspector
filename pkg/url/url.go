package url

import (
	"fmt"
	"net/url"
)

type URL struct {
	Scheme   string
	Hostname string
	Port     string
	Path     string
}

func Parse(s string) (URL, error) {
	u, err := url.Parse(s)
	if err != nil {
		fmt.Println("cannot parse url ", s)
		return URL{}, err
	}
	parsedURL := URL{}
	if u.Scheme != "http" && u.Scheme != "https" {
		return URL{}, fmt.Errorf("bad scheme: %s", u.Scheme)
	}
	parsedURL.Scheme = u.Scheme
	if u.Hostname() == "" {
		return URL{}, fmt.Errorf("empty hostname, aborting")
	}
	parsedURL.Hostname = u.Hostname()
	switch parsedURL.Scheme {
	case "http":
		parsedURL.Port = "80"
	case "https":
		parsedURL.Port = "443"
	}
	if u.Port() != "" {
		parsedURL.Port = u.Port()
	}
	parsedURL.Path = u.EscapedPath()
	if u.User != nil {
		fmt.Println("warning: basic auth in url will be ignored")
	}
	if u.RawQuery != "" {
		fmt.Printf("found query parameters %s in url, they will be ignored\n", u.RawQuery)
	}
	if u.Fragment != "" {
		fmt.Printf("found fragment %s in url, it will be ignored\n", u.Fragment)
	}
	return parsedURL, nil
}

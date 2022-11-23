package utils

import (
	"fmt"
	"net/url"
	"strings"
)

type Repository struct {
	Host  string
	Owner string
	Name  string
}

func (r Repository) Url() string {
	return fmt.Sprintf("https://%s/%s/%s", r.Host, r.Owner, r.Name)
}

// Parse extracts the repository information from the following
// string formats: "OWNER/REPO", "HOST/OWNER/REPO", and a full URL.
// If the format does not specify a host, use the host provided.
func ParseWithHost(s, host string) (repo Repository, err error) {
	if IsURL(s) {
		u, err := ParseURL(s)
		if err != nil {
			return repo, err
		}

		host, owner, name, err := RepoInfoFromURL(u)
		if err != nil {
			return repo, err
		}

		return Repository{
			Host:  host,
			Owner: owner,
			Name:  name,
		}, nil
	}

	parts := strings.SplitN(s, "/", 4)
	for _, p := range parts {
		if len(p) == 0 {
			return repo, fmt.Errorf(`expected the "[HOST/]OWNER/REPO" format, got %q`, s)
		}
	}

	switch len(parts) {
	case 3:
		return Repository{
			Host:  parts[0],
			Owner: parts[1],
			Name:  parts[2],
		}, nil
	case 2:
		return Repository{
			Host:  host,
			Owner: parts[0],
			Name:  parts[1],
		}, nil
	default:
		return repo, fmt.Errorf(`expected the "[HOST/]OWNER/REPO" format, got %q`, s)
	}
}

func IsURL(u string) bool {
	return strings.HasPrefix(u, "git@") || isSupportedProtocol(u)
}

func isSupportedProtocol(u string) bool {
	return strings.HasPrefix(u, "ssh:") ||
		strings.HasPrefix(u, "git+ssh:") ||
		strings.HasPrefix(u, "git:") ||
		strings.HasPrefix(u, "http:") ||
		strings.HasPrefix(u, "git+https:") ||
		strings.HasPrefix(u, "https:")
}

// ParseURL normalizes git remote urls.
func ParseURL(rawURL string) (u *url.URL, err error) {
	u, err = url.Parse(rawURL)
	if err != nil {
		return
	}

	if u.Scheme == "git+ssh" {
		u.Scheme = "ssh"
	}

	if u.Scheme == "git+https" {
		u.Scheme = "https"
	}

	if u.Scheme != "ssh" {
		return
	}

	if strings.HasPrefix(u.Path, "//") {
		u.Path = strings.TrimPrefix(u.Path, "/")
	}

	if idx := strings.Index(u.Host, ":"); idx >= 0 {
		u.Host = u.Host[0:idx]
	}

	return
}

// Extract GitHub repository information from a git remote URL.
func RepoInfoFromURL(u *url.URL) (host string, owner string, name string, err error) {
	if u.Hostname() == "" {
		return "", "", "", fmt.Errorf("no hostname detected")
	}

	parts := strings.SplitN(strings.Trim(u.Path, "/"), "/", 3)
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("invalid path: %s", u.Path)
	}

	return normalizeHostname(u.Hostname()), parts[0], strings.TrimSuffix(parts[1], ".git"), nil
}

func normalizeHostname(h string) string {
	return strings.ToLower(strings.TrimPrefix(h, "www."))
}

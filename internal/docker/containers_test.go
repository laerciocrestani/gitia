package docker

import "testing"

func TestFormatPublishers_hostOnlyURL(t *testing.T) {
	got := formatPublishers([]publisherRow{
		{URL: "0.0.0.0", PublishedPort: 8080, TargetPort: 80, Protocol: "tcp"},
	})
	if got != "8080:80" {
		t.Fatalf("got %q want 8080:80", got)
	}
}

func TestFormatPublishers_multiplePorts(t *testing.T) {
	got := formatPublishers([]publisherRow{
		{URL: "0.0.0.0", PublishedPort: 8443, TargetPort: 443, Protocol: "tcp"},
		{URL: "::", PublishedPort: 8443, TargetPort: 443, Protocol: "tcp"},
		{URL: "0.0.0.0", PublishedPort: 8080, TargetPort: 80, Protocol: "tcp"},
		{URL: "::", PublishedPort: 8080, TargetPort: 80, Protocol: "tcp"},
	})
	if got != "8443:443, 8080:80" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatPublishers_unpublished(t *testing.T) {
	got := formatPublishers([]publisherRow{
		{URL: "", PublishedPort: 0, TargetPort: 3306, Protocol: "tcp"},
	})
	if got != "" {
		t.Fatalf("got %q want empty", got)
	}
}

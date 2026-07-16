package collector

import (
	"context"
	"testing"
)

type fakeDocker struct {
	list []dockerContainer
	err  error
}

func (f *fakeDocker) ContainerList(ctx context.Context) ([]dockerContainer, error) {
	return f.list, f.err
}

func (f *fakeDocker) Close() error { return nil }

func TestDockerCollectorCollect(t *testing.T) {
	fake := &fakeDocker{
		list: []dockerContainer{
			{
				Names:  []string{"/peon-proxy"},
				State:  "running",
				Status: "Up 2 hours (healthy)",
				Image:  "traefik:v3.6.6",
			},
			{
				Names:  []string{"/webapp"},
				State:  "exited",
				Status: "Exited (0) 1 hour ago",
				Image:  "peon/webapp:latest",
			},
		},
	}

	c := NewDockerCollectorWithClient(fake)
	out, err := c.Collect(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 {
		t.Fatalf("got %d containers", len(out))
	}
	if out[0].Name != "peon-proxy" {
		t.Errorf("name = %q", out[0].Name)
	}
	if out[0].Health != "healthy" {
		t.Errorf("health = %q", out[0].Health)
	}
	if out[1].State != "exited" {
		t.Errorf("state = %q", out[1].State)
	}
}

func TestFirstName(t *testing.T) {
	if got := firstName([]string{"/foo"}); got != "foo" {
		t.Errorf("got %q", got)
	}
	if got := firstName(nil); got != "" {
		t.Errorf("got %q", got)
	}
}

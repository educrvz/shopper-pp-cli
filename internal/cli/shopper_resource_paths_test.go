// Copyright 2026 educrvz and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import "testing"

func TestTailKnownResourcesHaveReadPaths(t *testing.T) {
	resources := tailKnownResources()
	if len(resources) == 0 {
		t.Fatal("tail must expose at least one readable resource")
	}
	for _, resource := range resources {
		path, err := resourceReadPath(resource)
		if err != nil {
			t.Fatalf("resourceReadPath(%q): %v", resource, err)
		}
		if path == "" {
			t.Fatalf("resourceReadPath(%q) returned an empty path", resource)
		}
	}
}

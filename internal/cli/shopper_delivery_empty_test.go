// Copyright 2026 educrvz and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import (
	"errors"
	"testing"

	"shopper-pp-cli/internal/client"
)

func TestIsNoScheduledDelivery(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "Shopper empty delivery code",
			err:  &client.APIError{StatusCode: 400, Body: `{"error":{"code":"SHPPR019","message":"No scheduled delivery"}}`},
			want: true,
		},
		{
			name: "other bad request",
			err:  &client.APIError{StatusCode: 400, Body: `{"error":{"code":"OTHER"}}`},
		},
		{
			name: "non API error",
			err:  errors.New("network error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isNoScheduledDelivery(tt.err); got != tt.want {
				t.Fatalf("isNoScheduledDelivery() = %v, want %v", got, tt.want)
			}
		})
	}
}

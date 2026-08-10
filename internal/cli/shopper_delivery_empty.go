// Copyright 2026 educrvz and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-written: Shopper returns SHPPR019 when an authenticated store has no
// delivery scheduled. Read commands expose that valid empty state as success.
// pp:data-source live

package cli

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
	"shopper-pp-cli/internal/client"
)

func isNoScheduledDelivery(err error) bool {
	var apiErr *client.APIError
	return errors.As(err, &apiErr) &&
		apiErr.StatusCode == 400 &&
		strings.Contains(apiErr.Body, `"code":"SHPPR019"`)
}

func writeNoScheduledDelivery(cmd *cobra.Command, flags *rootFlags, resource string) error {
	return printJSONFiltered(cmd.OutOrStdout(), map[string]any{
		"resource": resource,
		"status":   "no_scheduled_delivery",
		"items":    []any{},
		"note":     "This store has no delivery scheduled.",
	}, flags)
}

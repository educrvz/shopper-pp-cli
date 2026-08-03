// Copyright 2026 educrvz and contributors. Licensed under Apache-2.0. See LICENSE.
// Hand-written: payment-aware checkout preparation for one-time and recurring
// storefronts. The CLI validates capabilities but never handles card secrets,
// generates payment instruments, or confirms an order.
// pp:data-source live

package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type checkoutPaymentMethod struct {
	Name            string `json:"name"`
	Available       bool   `json:"available"`
	MinimumLeadDays int    `json:"minimum_lead_days,omitempty"`
	Completion      string `json:"completion"`
	Note            string `json:"note"`
}

func paymentMethods(params *checkoutPayment) []checkoutPaymentMethod {
	if params == nil {
		return nil
	}
	return []checkoutPaymentMethod{
		{
			Name:            "card",
			Available:       params.MinDaysCardCredit > 0,
			MinimumLeadDays: params.MinDaysCardCredit,
			Completion:      "browser",
			Note:            "Choose a saved card at checkout, or add a card in Shopper's browser form. The CLI never reads card numbers or security codes.",
		},
		{
			Name:            "boleto",
			Available:       params.MinDaysBoleto > 0,
			MinimumLeadDays: params.MinDaysBoleto,
			Completion:      "browser",
			Note:            "Choose boleto at checkout. Shopper generates the bank slip only after browser confirmation.",
		},
		{
			Name:       "pix",
			Available:  params.AllowPix,
			Completion: "browser",
			Note:       "Choose PIX only when the live store capability enables it. Shopper generates the QR code or copy-and-paste code in the browser.",
		},
	}
}

func normalizePaymentMethod(value string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "card", "credit-card", "credit_card", "saved-card", "saved_card", "cartao", "cartão":
		return "card", nil
	case "boleto", "bank-slip", "bank_slip", "bankslip":
		return "boleto", nil
	case "pix":
		return "pix", nil
	default:
		return "", fmt.Errorf("unknown payment method %q; use card, boleto, or pix", value)
	}
}

func selectPaymentMethod(params *checkoutPayment, requested string) (checkoutPaymentMethod, error) {
	method, err := normalizePaymentMethod(requested)
	if err != nil {
		return checkoutPaymentMethod{}, err
	}
	for _, candidate := range paymentMethods(params) {
		if candidate.Name != method {
			continue
		}
		if !candidate.Available {
			return candidate, fmt.Errorf("payment method %q is not enabled for this store according to live Shopper capabilities", method)
		}
		return candidate, nil
	}
	return checkoutPaymentMethod{}, fmt.Errorf("payment capabilities are unavailable for this store")
}

func newPaymentMethodsCmd(flags *rootFlags) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "methods",
		Short: "List live payment methods and minimum lead times for the active store",
		Long: `Reads the active storefront's live payment parameters and reports whether
card, boleto, and PIX are currently enabled. Payment completion always happens
in Shopper's authenticated browser checkout; this command is read-only.`,
		Example: "  shopper-pp-cli payment methods --store unica --agent\n  shopper-pp-cli payment methods --store now --agent",
		Annotations: map[string]string{
			"mcp:read-only":          "true",
			"pp:no-error-path-probe": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if dryRunOK(flags) {
				return writeDryRun(cmd.OutOrStdout(), flags, "list-payment-methods")
			}
			storeName := flags.store
			if storeName == "" {
				storeName = "programada"
			}
			c, err := flags.newClient()
			if err != nil {
				return err
			}
			storesData, err := c.Get(cmd.Context(), "/features/stores", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			params := extractPaymentParams(storesData, storeName)
			if params == nil {
				return apiErr(fmt.Errorf("Shopper returned no payment capabilities for store %q", storeName))
			}
			return printJSONFiltered(cmd.OutOrStdout(), map[string]any{
				"store":               storeName,
				"methods":             paymentMethods(params),
				"completion":          "browser",
				"checkout_url":        "https://" + resolveSubdomain(storeName) + ".shopper.com.br/shop/checkout",
				"handles_card_data":   false,
				"capability_source":   "GET /features/stores",
				"availability_notice": "Availability can change; run this immediately before checkout.",
			}, flags)
		},
	}
	return cmd
}

func newCheckoutPrepareCmd(flags *rootFlags) *cobra.Command {
	var requestedPayment string
	var openBrowserFlag bool

	cmd := &cobra.Command{
		Use:   "prepare",
		Short: "Validate a payment method and prepare a safe browser checkout handoff",
		Long: `Prepares a recurring or one-time purchase for final browser checkout.

The command validates the selected payment method against live store capabilities,
checks the current cart and minimum order value, and prints the exact checkout URL
and next steps. Use --open to launch that page after validation.

The CLI does not submit the order, charge a saved card, handle raw card data,
generate a PIX code, or generate a boleto. Those consequential actions remain in
Shopper's authenticated browser session behind its CSRF protection.`,
		Example: "  shopper-pp-cli checkout prepare --store unica --payment card --agent\n  shopper-pp-cli checkout prepare --store unica --payment boleto --open\n  shopper-pp-cli checkout prepare --store now --payment pix --agent",
		Annotations: map[string]string{
			"pp:no-error-path-probe": "true",
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			methodName, err := normalizePaymentMethod(requestedPayment)
			if err != nil {
				return usageErr(err)
			}
			storeName := flags.store
			if storeName == "" {
				storeName = "programada"
			}
			checkoutURL := "https://" + resolveSubdomain(storeName) + ".shopper.com.br/shop/checkout"
			if dryRunOK(flags) {
				return printJSONFiltered(cmd.OutOrStdout(), map[string]any{
					"dry_run":        true,
					"store":          storeName,
					"payment_method": methodName,
					"would_validate": []string{"GET /features/stores", "GET /cart/summary"},
					"would_open":     openBrowserFlag,
					"checkout_url":   checkoutURL,
				}, flags)
			}

			c, err := flags.newClient()
			if err != nil {
				return err
			}
			storesData, err := c.Get(cmd.Context(), "/features/stores", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			selected, err := selectPaymentMethod(extractPaymentParams(storesData, storeName), methodName)
			if err != nil {
				return usageErr(err)
			}

			cartData, err := c.Get(cmd.Context(), "/cart/summary", nil)
			if err != nil {
				return classifyAPIError(err, flags)
			}
			cart := extractCheckoutCart(cartData)
			blockers := []string{}
			if cart == nil || cart.ItemCount == 0 {
				blockers = append(blockers, "cart is empty")
			}
			if cart != nil && cart.MinValue > 0 && !cart.MinValueMet {
				blockers = append(blockers, fmt.Sprintf("cart total is below the store minimum of %.2f", cart.MinValue))
			}

			steps := []string{
				"Review the delivery address and delivery slot in Shopper checkout.",
			}
			switch selected.Name {
			case "card":
				steps = append(steps, "Choose a saved card. If no suitable card is available, add one in Shopper's browser card form.")
			case "boleto":
				steps = append(steps, "Choose boleto and confirm in the browser; Shopper generates the bank slip after confirmation.")
			case "pix":
				steps = append(steps, "Choose PIX and confirm in the browser; Shopper then displays the QR code or copy-and-paste code.")
			}
			steps = append(steps, "Review the final total and explicitly confirm the order in the browser.")

			ready := len(blockers) == 0
			result := map[string]any{
				"store":                    storeName,
				"purchase_type":            map[bool]string{true: "recurring", false: "one-time"}[isSubscriptionStore(storeName)],
				"payment_method":           selected,
				"cart":                     cart,
				"ready_to_open":            ready,
				"blockers":                 blockers,
				"checkout_url":             checkoutURL,
				"steps":                    steps,
				"browser_confirmation":     true,
				"cli_submits_order":        false,
				"cli_handles_card_secrets": false,
			}
			if err := printJSONFiltered(cmd.OutOrStdout(), result, flags); err != nil {
				return err
			}
			if openBrowserFlag {
				if !ready {
					return usageErr(fmt.Errorf("checkout has blockers; resolve them before using --open"))
				}
				return openBrowser(checkoutURL)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&requestedPayment, "payment", "", "Payment method: card, boleto, or pix")
	_ = cmd.MarkFlagRequired("payment")
	cmd.Flags().BoolVar(&openBrowserFlag, "open", false, "Open Shopper checkout after validation")
	return cmd
}

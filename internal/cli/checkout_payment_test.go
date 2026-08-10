// Copyright 2026 educrvz and contributors. Licensed under Apache-2.0. See LICENSE.

package cli

import "testing"

func TestNormalizePaymentMethod(t *testing.T) {
	tests := map[string]string{
		"card":       "card",
		"saved-card": "card",
		"cartão":     "card",
		"boleto":     "boleto",
		"bank-slip":  "boleto",
		"PIX":        "pix",
	}
	for input, want := range tests {
		got, err := normalizePaymentMethod(input)
		if err != nil || got != want {
			t.Fatalf("normalizePaymentMethod(%q) = %q, %v; want %q", input, got, err, want)
		}
	}
	if _, err := normalizePaymentMethod("cash"); err == nil {
		t.Fatal("normalizePaymentMethod(cash) should fail")
	}
}

func TestSelectPaymentMethodHonorsLiveCapabilities(t *testing.T) {
	params := &checkoutPayment{MinDaysCardCredit: 2, MinDaysBoleto: 5, AllowPix: false}
	card, err := selectPaymentMethod(params, "saved-card")
	if err != nil || card.Name != "card" || card.MinimumLeadDays != 2 {
		t.Fatalf("card selection = %#v, %v", card, err)
	}
	if _, err := selectPaymentMethod(params, "pix"); err == nil {
		t.Fatal("disabled PIX should fail validation")
	}
}

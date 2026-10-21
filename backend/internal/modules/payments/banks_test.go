package payments

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// A local stand-in for Paystack's /bank and /bank/resolve (shapes from Paystack's API reference).
func fakePaystack(t *testing.T) *Paystack {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk_test" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/bank":
			if r.URL.Query().Get("use_cursor") != "true" || r.URL.Query().Get("country") != "nigeria" {
				t.Errorf("bank list query: %s", r.URL.RawQuery)
			}
			if r.URL.Query().Get("next") == "" {
				_, _ = w.Write([]byte(`{"status":true,"data":[{"name":"Access Bank","code":"044","active":true},{"name":"Old Bank","code":"999","active":false}],"meta":{"next":"cursor-2"}}`))
				return
			}
			_, _ = w.Write([]byte(`{"status":true,"data":[{"name":"Guaranty Trust Bank","code":"058","active":true}],"meta":{"next":null}}`))
		case "/bank/resolve":
			if r.URL.Query().Get("account_number") == "0000000000" {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = w.Write([]byte(`{"status":false,"message":"Could not resolve account name. Check parameters or try again."}`))
				return
			}
			_, _ = w.Write([]byte(`{"status":true,"message":"Account number resolved","data":{"account_number":"0123456789","account_name":"ADAEZE OKAFOR","bank_id":9}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return &Paystack{secret: "sk_test", base: srv.URL, client: &http.Client{Timeout: 5 * time.Second}}
}

func TestPaystackBankListFollowsCursorAndSkipsInactive(t *testing.T) {
	banks, err := fakePaystack(t).banks(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(banks) != 2 || banks[0].Code != "044" || banks[1].Code != "058" {
		t.Fatalf("banks = %+v", banks)
	}
}

func TestPaystackResolveAccount(t *testing.T) {
	p := fakePaystack(t)
	name, err := p.resolveAccount(context.Background(), "058", "0123456789")
	if err != nil || name != "ADAEZE OKAFOR" {
		t.Fatalf("resolve = %q, %v", name, err)
	}
	if _, err := p.resolveAccount(context.Background(), "058", "0000000000"); !errors.Is(err, ErrAccountNotFound) {
		t.Fatalf("unknown account should be ErrAccountNotFound, got %v", err)
	}
}

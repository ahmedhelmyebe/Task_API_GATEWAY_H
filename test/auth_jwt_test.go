package test


import (
"testing"
"example.com/api-gateway/config"
"example.com/api-gateway/internal/auth"
)


func TestJWTSignParse(t *testing.T) {
cfg := config.JWT{Issuer: "iss", Audience: "aud", Secret: "s", TTLMinutes: 1}
tok, err := auth.Sign(cfg, "u1", "admin")
if err != nil { t.Fatal(err) }
cl, err := auth.Parse(cfg, tok)
if err != nil { t.Fatal(err) }
if cl.Sub != "u1" || cl.Role != "admin" { t.Fatalf("unexpected claims: %+v", cl) }
}
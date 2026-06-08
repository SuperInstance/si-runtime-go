package siruntime

import (
	"math"
	"testing"
)

func TestAgentHomeostasis(t *testing.T) {
	tests := []struct {
		name     string
		initial  float64
		target   float64
		rate     float64
		want     float64
		wantErr  float64
	}{
		{"halfway", 50, 100, 0.1, 55, 45},
		{"no movement", 50, 50, 0.1, 50, 0},
		{"full reach", 50, 100, 1.0, 100, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := NewAgent("test")
			a.SetState("energy", tt.initial)
			a.SetHomeostasis("energy", tt.target)
			a.UpdateHomeostasis(tt.rate)
			v, _ := a.GetState("energy")
			if math.Abs(v-tt.want) > 1e-9 {
				t.Errorf("energy=%.2f want %.2f", v, tt.want)
			}
			if math.Abs(a.HomeostasisError()-tt.wantErr) > 1e-9 {
				t.Errorf("error=%.2f want %.2f", a.HomeostasisError(), tt.wantErr)
			}
		})
	}
}

func TestAgentCapability(t *testing.T) {
	a := NewAgent("a1")
	tests := []struct {
		name    string
		cap     string
		wantErr bool
		add     bool
	}{
		{"add fly", "fly", false, true},
		{"duplicate fly", "fly", true, true},
		{"add swim", "swim", false, true},
		{"empty name", "", true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.add {
				err := a.AddCapability(tt.cap)
				if (err != nil) != tt.wantErr {
					t.Errorf("AddCapability(%q) error=%v wantErr=%v", tt.cap, err, tt.wantErr)
				}
			}
		})
	}
	if len(a.ListCapabilities()) != 2 {
		t.Errorf("expected 2 capabilities, got %d", len(a.ListCapabilities()))
	}
	if !a.RemoveCapability("fly") {
		t.Error("expected fly removal to succeed")
	}
	if a.RemoveCapability("run") {
		t.Error("expected run removal to fail")
	}
}

func TestAgentString(t *testing.T) {
	a := NewAgent("x")
	a.SetState("load", 1.0)
	a.AddCapability("compute")
	s := a.String()
	if s != "Agent[x] caps=1 states=1" {
		t.Errorf("unexpected string: %q", s)
	}
}

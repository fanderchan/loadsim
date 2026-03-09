package stress

import "testing"

func TestParseMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    Mode
		wantErr bool
	}{
		{name: "fixed", input: "fixed", want: ModeFixed},
		{name: "wave", input: "wave", want: ModeWave},
		{name: "invalid", input: "burst", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseMode(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error for mode %q", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Fatalf("got %q, want %q", got, tt.want)
			}
		})
	}
}

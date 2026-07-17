package ws

import "testing"

func TestMessageExceedsLimitCountsUnicodeCharacters(t *testing.T) {
	tests := []struct {
		name       string
		content    string
		maxLength  int
		wantExceed bool
	}{
		{name: "at limit", content: "hello", maxLength: 5, wantExceed: false},
		{name: "over limit", content: "hello!", maxLength: 5, wantExceed: true},
		{name: "unicode at limit", content: "xin chào 👋", maxLength: 10, wantExceed: false},
		{name: "unicode over limit", content: "xin chào 👋!", maxLength: 10, wantExceed: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := messageExceedsLimit(tt.content, tt.maxLength); got != tt.wantExceed {
				t.Errorf("messageExceedsLimit(%q, %d) = %v, want %v", tt.content, tt.maxLength, got, tt.wantExceed)
			}
		})
	}
}

func TestMaxInboundMessageSizeIncludesJSONEncodingOverhead(t *testing.T) {
	if got, minimum := maxInboundMessageSize(1000), int64(6000); got <= minimum {
		t.Errorf("maxInboundMessageSize(1000) = %d, want greater than %d", got, minimum)
	}
}

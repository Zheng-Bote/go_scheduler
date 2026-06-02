/**
 * SPDX-FileComment: HTTP Server Tests
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file server_test.go
 * @brief Unit tests for HTTP server utilities
 * @version 1.0.0
 * @date 2026-06-02
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package http

import (
	"testing"
	"time"
)

func TestParseDateParam(t *testing.T) {
	tests := []struct {
		input    string
		expected *time.Time
		wantErr  bool
	}{
		{
			input:    "",
			expected: nil,
			wantErr:  false,
		},
		{
			input:    "2026-06-02T10:16:42Z",
			wantErr:  false,
		},
		{
			input:    "2026-06-02",
			wantErr:  false,
		},
		{
			input:    "invalid-date",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		got, err := parseDateParam(tt.input)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseDateParam(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			continue
		}
		if tt.input == "" && got != nil {
			t.Errorf("parseDateParam(%q) = %v, want nil", tt.input, got)
		}
		if tt.input == "2026-06-02" && got != nil {
			expectedTime := time.Date(2026, 6, 2, 0, 0, 0, 0, time.UTC)
			if !got.Equal(expectedTime) {
				t.Errorf("parseDateParam(%q) = %v, want %v", tt.input, got, expectedTime)
			}
		}
	}
}

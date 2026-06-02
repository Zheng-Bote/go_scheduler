//go:build !windows

/**
 * SPDX-FileComment: Scheduler Admin
 * SPDX-FileType: SOURCE
 * SPDX-FileContributor: ZHENG Robert
 * SPDX-FileCopyrightText: 2026 ZHENG Robert
 * SPDX-License-Identifier: Apache-2.0
 *
 * @file hello_other.go
 * @brief No-op Windows Hello verification stub for non-Windows builds
 * @version 1.0.0
 * @date 2026-06-02
 *
 * @author ZHENG Robert (robert@hase-zheng.net)
 * @copyright Copyright (c) 2026 ZHENG Robert
 * @LICENSE Apache-2.0
 */

package main

import (
	"fyne.io/fyne/v2"
)

// verifyWindowsHello is a no-op stub for non-Windows platforms. It always
// returns (true, nil) so that the admin GUI can call it unconditionally
// without platform-specific build tags.
func verifyWindowsHello(w fyne.Window) (bool, error) {
	// No-op for non-Windows platforms, just return true
	return true, nil
}

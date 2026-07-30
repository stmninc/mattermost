// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"os"
	"sync"
)

// TownSquareReadOnlyEnvVar is the environment variable that, when set to "true",
// makes the default channel ("town-square") read-only for non system admins in
// this custom Engage Chat build. This is intentionally NOT part of the
// model.Config data structure so it can be toggled per tenant purely via the
// environment, mirroring INTEGRATION_ADMIN_USERNAME used by the official channel
// feature.
const TownSquareReadOnlyEnvVar = "ENGAGECHAT_TOWNSQUARE_READONLY"

var (
	townSquareReadOnly     bool
	townSquareReadOnlyOnce sync.Once
)

// IsTownSquareReadOnlyEnabled reports whether the read-only Town Square feature
// is enabled via the ENGAGECHAT_TOWNSQUARE_READONLY environment variable. The
// value is read once for performance using sync.Once.
func IsTownSquareReadOnlyEnabled() bool {
	townSquareReadOnlyOnce.Do(func() {
		townSquareReadOnly = os.Getenv(TownSquareReadOnlyEnvVar) == "true"
	})
	return townSquareReadOnly
}

// ResetTownSquareReadOnlyCache resets the cached value so tests can toggle the
// environment variable between cases.
func ResetTownSquareReadOnlyCache() {
	townSquareReadOnly = false
	townSquareReadOnlyOnce = sync.Once{}
}

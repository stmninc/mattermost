// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsTownSquareReadOnlyEnabled(t *testing.T) {
	t.Run("unset defaults to false", func(t *testing.T) {
		t.Setenv(TownSquareReadOnlyEnvVar, "")
		ResetTownSquareReadOnlyCache()
		defer ResetTownSquareReadOnlyCache()
		require.False(t, IsTownSquareReadOnlyEnabled())
	})

	t.Run("set to true enables the feature", func(t *testing.T) {
		t.Setenv(TownSquareReadOnlyEnvVar, "true")
		ResetTownSquareReadOnlyCache()
		defer ResetTownSquareReadOnlyCache()
		require.True(t, IsTownSquareReadOnlyEnabled())
	})

	t.Run("any other value is false", func(t *testing.T) {
		t.Setenv(TownSquareReadOnlyEnvVar, "1")
		ResetTownSquareReadOnlyCache()
		defer ResetTownSquareReadOnlyCache()
		require.False(t, IsTownSquareReadOnlyEnabled())
	})

	t.Run("value is cached until reset", func(t *testing.T) {
		t.Setenv(TownSquareReadOnlyEnvVar, "true")
		ResetTownSquareReadOnlyCache()
		defer ResetTownSquareReadOnlyCache()
		require.True(t, IsTownSquareReadOnlyEnabled())

		// Changing the env without resetting keeps the cached value.
		t.Setenv(TownSquareReadOnlyEnvVar, "false")
		require.True(t, IsTownSquareReadOnlyEnabled())

		// After reset the new value takes effect.
		ResetTownSquareReadOnlyCache()
		require.False(t, IsTownSquareReadOnlyEnabled())
	})
}

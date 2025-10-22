# Cherry-Pick Workflow Test

This file is created to test the automated cherry-pick workflow for Issue #571.

## Purpose

Test that the GitHub Actions workflow can automatically create a cherry-pick PR when a merged PR receives a `/cherry-pick <branch-name>` comment.

## Test Details

- **Issue**: https://github.com/stmninc/tunag-chat/issues/571
- **Test Branch**: test/issue_571_cherry_pick_workflow_sample
- **Target Branch**: release-test
- **Test Date**: 2025-10-22

## Expected Behavior

1. Merge this PR to `test-main`
2. Comment `/cherry-pick release-test` on the merged PR
3. Workflow should automatically create a new PR from `test-main` to `release-test`
4. The new PR should contain this file change

## Workflow File

The workflow is defined in `.github/workflows/cherry_pick_on_comment.yaml`

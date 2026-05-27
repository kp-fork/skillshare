# skillshare v0.19.24 Release Notes

Release date: 2026-05-27

## TL;DR

1. **Local git checkpoints** — `skillshare commit` creates a local commit for your source skills without requiring a remote or pushing anywhere.
2. **Dashboard Commit locally action** — the Git Sync page now lets you commit local changes from the browser without sharing them yet.
3. **Update page status stays current** — after dashboard updates finish, successful rows are marked **Up to date** instead of falling back to **Unchecked**.

---

## Local Checkpoints Without Pushing

`skillshare commit` gives you a quick way to save a restore point while editing skills locally. It stages source changes and creates a git commit, but never pushes to a remote. This is useful when you want history for experiments, drafts, or machines that do not have a remote configured yet.

```bash
skillshare commit -m "Update review skill"
skillshare commit --dry-run
```

Use `skillshare push` when you are ready to share those changes with a remote.

## Commit Locally from the Dashboard

The Git Sync page now includes a **Commit locally** action next to the existing push and pull controls. It uses the same commit message and dry-run preview area, but only creates a local commit.

Push remains the action for sending changes to a remote; Commit locally is for saving local history without publishing anything.

## Dashboard Update Status Stays Current

After updating skills from the dashboard, successfully updated or already-current items now keep their latest check status and show **Up to date**. The page no longer resets those rows to **Unchecked** after the update finishes.

// Copyright 2015 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package git

import (
	"context"
	"strings"
)

// MergeBaseOptions contains optional arguments for getting merge base.
//
// Docs: https://git-scm.com/docs/git-merge-base
type MergeBaseOptions struct {
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// MergeBase returns merge base between base and head revisions of the
// repository.
func (r *Repository) MergeBase(ctx context.Context, base, head string, opts ...MergeBaseOptions) (string, error) {
	var opt MergeBaseOptions
	if len(opts) > 0 {
		opt = opts[0]
	}

	args := []string{"merge-base"}
	args = append(args, opt.CommandOptions.Args...)
	args = append(args, "--end-of-options", base, head)

	stdout, err := gitRun(ctx, r.path, args, opt.CommandOptions.Envs)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			return "", ErrNoMergeBase
		}
		return "", err
	}
	return strings.TrimSpace(string(stdout)), nil
}

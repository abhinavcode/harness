// Copyright 2023 Harness, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package githook

import (
	"context"
	"fmt"

	"github.com/harness/gitness/git"
	"github.com/harness/gitness/git/hook"
	"github.com/harness/gitness/types"
)

func (c *Controller) processObjects(
	ctx context.Context,
	rgit RestrictedGIT,
	repo *types.RepositoryCore,
	principal *types.Principal,
	refUpdates changedRefs,
	checks *protectionChecks,
	in types.GithookPreReceiveInput,
	output *hook.Output,
) (git.ProcessPreReceiveObjectsOutput, error) {
	if refUpdates.hasOnlyDeletedBranches() {
		return git.ProcessPreReceiveObjectsOutput{}, nil
	}

	sizeLimit := checks.SettingsFileSizeLimit
	if checks.RulesFileSizeLimit != 0 && checks.RulesFileSizeLimit < sizeLimit {
		sizeLimit = checks.RulesFileSizeLimit
	}

	principalCommitterMatch := checks.SettingsPrincipalCommitterMatch || checks.RulesPrincipalCommitterMatch

	if sizeLimit == 0 && !principalCommitterMatch && !checks.SettingsGitLFSEnabled {
		return git.ProcessPreReceiveObjectsOutput{}, nil
	}

	preReceiveObjsIn := git.ProcessPreReceiveObjectsParams{
		ReadParams: git.ReadParams{
			RepoUID:             repo.GitUID,
			AlternateObjectDirs: in.Environment.AlternateObjectDirs,
		},
	}

	if sizeLimit > 0 {
		preReceiveObjsIn.FindOversizeFilesParams = &git.FindOversizeFilesParams{
			SizeLimit: sizeLimit,
		}
	}

	if principalCommitterMatch && principal != nil && !in.Internal {
		preReceiveObjsIn.FindCommitterMismatchParams = &git.FindCommitterMismatchParams{
			PrincipalEmail: principal.Email,
		}
	}

	if checks.SettingsGitLFSEnabled {
		preReceiveObjsIn.FindLFSPointersParams = &git.FindLFSPointersParams{}
	}

	preReceiveObjsOut, err := rgit.ProcessPreReceiveObjects(
		ctx,
		preReceiveObjsIn,
	)
	if err != nil {
		return git.ProcessPreReceiveObjectsOutput{}, fmt.Errorf("failed to process pre-receive objects: %w", err)
	}

	if preReceiveObjsOut.FindOversizeFilesOutput != nil &&
		len(preReceiveObjsOut.FindOversizeFilesOutput.FileInfos) > 0 {
		printOversizeFiles(
			output,
			preReceiveObjsOut.FindOversizeFilesOutput.FileInfos,
			preReceiveObjsOut.FindOversizeFilesOutput.Total,
			sizeLimit,
		)
	}

	if preReceiveObjsOut.FindCommitterMismatchOutput != nil &&
		len(preReceiveObjsOut.FindCommitterMismatchOutput.CommitInfos) > 0 {
		printCommitterMismatch(
			output,
			preReceiveObjsOut.FindCommitterMismatchOutput.CommitInfos,
			preReceiveObjsIn.FindCommitterMismatchParams.PrincipalEmail,
			preReceiveObjsOut.FindCommitterMismatchOutput.Total,
		)
	}

	if preReceiveObjsOut.FindLFSPointersOutput != nil &&
		len(preReceiveObjsOut.FindLFSPointersOutput.LFSInfos) > 0 {
		objIDs := make([]string, len(preReceiveObjsOut.FindLFSPointersOutput.LFSInfos))
		for i, info := range preReceiveObjsOut.FindLFSPointersOutput.LFSInfos {
			objIDs[i] = info.ObjID
		}

		existingObjs, err := c.lfsStore.FindMany(ctx, in.RepoID, objIDs)
		if err != nil {
			return git.ProcessPreReceiveObjectsOutput{}, fmt.Errorf("failed to find lfs objects: %w", err)
		}

		//nolint:lll
		if len(existingObjs) != len(objIDs) {
			printLFSPointers(
				output,
				preReceiveObjsOut.FindLFSPointersOutput.LFSInfos,
				preReceiveObjsOut.FindLFSPointersOutput.Total,
			)
		}
	}

	return preReceiveObjsOut, nil
}

package handlers

import (
	"hermes-relay/internal/cqrs/commands"
	"hermes-relay/internal/domain/entities/code"
	"hermes-relay/internal/lib/test-helpers/domain-helpers"
	"hermes-relay/internal/lib/utils"
	rh "hermes-relay/internal/lib/test-helpers/router-helpers"
	"testing"
)

var (
	testProjectID        = utils.NewID()
	testCodeID1          = utils.NewID()
	testCodeID2          = utils.NewID()
	testDeleteCodeID     = utils.NewID()
	testNonexistentID    = utils.NewID()
)

var cmds = []*commands.AnyMessage{
	commands.ToAny(commands.NewDomainEvent[any, any](
		"CreatedProject",
		map[string]any{"name": "Test Project"},
		"Project",
		testProjectID,
		domain_helpers.TestActor(),
		nil,
	)),
	commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](
		code.CreatedCode,
		code.CreatedCodePayload{
			ProjectID: testProjectID,
			Slug:      "topic:climate",
			Color:     "blue",
			Definition: "Climate topics",
		},
		code.EntityName,
		testCodeID1,
		domain_helpers.TestActor(),
		nil,
	)),
	commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](
		code.CreatedCode,
		code.CreatedCodePayload{
			ProjectID: testProjectID,
			Slug:      "topic:health",
			Color:     "green",
			Definition: "Health topics",
		},
		code.EntityName,
		testCodeID2,
		domain_helpers.TestActor(),
		nil,
	)),
}

func TestCodeRouter(t *testing.T) {
	tests := []rh.RouterTestCase{
		{
			Name: "CreateCode with valid payload",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: testProjectID,
				Slug:      "topic:economy",
				Color:     "green",
				Definition: "Economic topics",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[code.CreatedCodePayload, any](code.CreatedCode, code.CreatedCodePayload{
				ProjectID: testProjectID,
				Slug:      "topic:economy",
				Color:     "green",
				Definition: "Economic topics",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
		},
		{
			Name: "UpdateCode with both fields for existing code",
			Input: commands.ToAny(commands.NewCommand[code.UpdateCodePayload, any](code.UpdateCode, code.UpdateCodePayload{
				Color:     "teal",
				Definition: "Updated climate coverage",
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[code.UpdatedCodePayload, any](code.UpdatedCode, code.UpdatedCodePayload{
				Color:     "teal",
				Definition: "Updated climate coverage",
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "UpdateCode for non-existent code fails",
			Input: commands.ToAny(commands.NewCommand[code.UpdateCodePayload, any](code.UpdateCode, code.UpdateCodePayload{
				Color: "red",
			}, code.EntityName, "code-nonexistent", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: ProjectID not found",
		},
		{
			Name: "DeleteCode",
			Input: commands.ToAny(commands.NewCommand[code.DeleteCodePayload, any](code.DeleteCode, code.DeleteCodePayload{}, code.EntityName, testDeleteCodeID, domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](code.DeletedCode, nil, code.EntityName, testDeleteCodeID, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "CreateCode with missing Color and Definition",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: testProjectID,
				Slug:      "topic:incomplete",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Color is required, Definition is required",
		},
		{
			Name: "CreateCode with invalid color (not a radix color)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID:  testProjectID,
				Slug:       "topic:invalid-color",
				Color:      "emerald",
				Definition: "Invalid color",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Color must be a valid Radix color",
		},
		{
			Name: "CreateCode with invalid color (has number suffix)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID:  testProjectID,
				Slug:       "topic:invalid-color-suffix",
				Color:      "blue-500",
				Definition: "Invalid color format",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Color must be a valid Radix color",
		},
		{
			Name: "CreateCode with invalid slug (no colon)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: testProjectID,
				Slug:      "topicclimate",
				Color:     "blue",
				Definition: "Invalid format",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			Name: "CreateCode with invalid slug (uppercase)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: testProjectID,
				Slug:      "Topic:climate",
				Color:     "blue",
				Definition: "Invalid format",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			Name: "CreateCode with invalid slug (empty after colon)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: testProjectID,
				Slug:      "topic:",
				Color:     "blue",
				Definition: "Invalid format",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			Name: "CreateCode with invalid slug (starts with hyphen)",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: testProjectID,
				Slug:      "topic:-climate",
				Color:     "blue",
				Definition: "Invalid format",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: Slug must match code slug format (lowercase with colon and optional dashes)",
		},
		{
			Name: "Wrong entity type returns nil",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: testProjectID,
				Slug:      "topic:test",
				Color:     "blue",
				Definition: "Test",
			}, "DifferentEntity", "", domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
		{
			Name: "Wrong action returns nil",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any]("DifferentAction", code.CreateCodePayload{
				ProjectID: testProjectID,
				Slug:      "topic:test",
				Color:     "blue",
				Definition: "Test",
			}, code.EntityName, "test-aggregate-id", domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: nil,
		},
		{
			Name: "CreateCode with duplicate slug fails",
			Input: commands.ToAny(commands.NewCommand[code.CreateCodePayload, any](code.CreateCode, code.CreateCodePayload{
				ProjectID: testProjectID,
				Slug:      "topic:climate",
				Color:     "blue",
				Definition: "Duplicate slug",
			}, code.EntityName, "", domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: slug already in use",
		},
		{
			Name: "MergeCodes with valid source and target",
			Input: commands.ToAny(commands.NewCommand[code.MergeCodesPayload, any](code.MergeCodes, code.MergeCodesPayload{
				SourceID: testCodeID1,
				TargetID: testCodeID2,
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[code.MergedCodesPayload, any](code.MergedCodes, code.MergedCodesPayload{
				SourceID: testCodeID1,
				TargetID: testCodeID2,
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "MergeCodes with non-existent source fails",
			Input: commands.ToAny(commands.NewCommand[code.MergeCodesPayload, any](code.MergeCodes, code.MergeCodesPayload{
				SourceID: testNonexistentID,
				TargetID: testCodeID2,
			}, code.EntityName, testCodeID2, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: source_id not found",
		},
		{
			Name: "MergeCodes with non-existent target fails",
			Input: commands.ToAny(commands.NewCommand[code.MergeCodesPayload, any](code.MergeCodes, code.MergeCodesPayload{
				SourceID: testCodeID1,
				TargetID: testNonexistentID,
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: target_id not found",
		},
		{
			Name: "MergeCodes with same source and target fails",
			Input: commands.ToAny(commands.NewCommand[code.MergeCodesPayload, any](code.MergeCodes, code.MergeCodesPayload{
				SourceID: testCodeID1,
				TargetID: testCodeID1,
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: source_id cannot merge with itself",
		},
		{
			Name:        "ClearCodeApplications",
			Input:       commands.ToAny(commands.NewCommand[code.ClearCodeApplicationsPayload, any](code.ClearCodeApplications, code.ClearCodeApplicationsPayload{}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
			ExpectErr:   "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[any, any](code.ClearedCodeApplications, nil, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "RecodeAll with valid target",
			Input: commands.ToAny(commands.NewCommand[code.RecodeAllPayload, any](code.RecodeAll, code.RecodeAllPayload{
				TargetCodeID: testCodeID2,
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
			ExpectErr: "",
			ExpectEvent: commands.ToAny(commands.NewDomainEvent[code.RecodedAllPayload, any](code.RecodedAll, code.RecodedAllPayload{
				TargetCodeID: testCodeID2,
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
		},
		{
			Name: "RecodeAll with non-existent target fails",
			Input: commands.ToAny(commands.NewCommand[code.RecodeAllPayload, any](code.RecodeAll, code.RecodeAllPayload{
				TargetCodeID: testNonexistentID,
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: target_code_id not found",
		},
		{
			Name: "RecodeAll to itself fails",
			Input: commands.ToAny(commands.NewCommand[code.RecodeAllPayload, any](code.RecodeAll, code.RecodeAllPayload{
				TargetCodeID: testCodeID1,
			}, code.EntityName, testCodeID1, domain_helpers.TestActor(), nil)),
			ExpectErr: "validation failed: target_code_id cannot recode to itself",
		},
	}

	rh.RunRouterTests(t, cmds, tests, NewRouter)
}

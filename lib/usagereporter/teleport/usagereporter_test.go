// Copyright 2023 Gravitational, Inc
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package usagereporter

import (
	"testing"

	"github.com/gravitational/teleport"
	"github.com/gravitational/trace"
	"github.com/stretchr/testify/require"

	usageeventsv1 "github.com/gravitational/teleport/api/gen/proto/go/usageevents/v1"
	prehogv1a "github.com/gravitational/teleport/gen/proto/go/prehog/v1alpha"
	"github.com/gravitational/teleport/lib/utils"
)

func TestConvertUsageEvent(t *testing.T) {
	anonymizer, err := utils.NewHMACAnonymizer("cluster-id")
	require.NoError(t, err)

	expectedAnonymizedUserString := anonymizer.AnonymizeString("myuser")

	for _, tt := range []struct {
		name             string
		event            *usageeventsv1.UsageEventOneOf
		identityUsername string
		isSSOUser        bool
		errCheck         require.ErrorAssertionFunc
		expected         *prehogv1a.SubmitEventRequest
	}{
		{
			name: "discover started event",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiDiscoverStartedEvent{
				UiDiscoverStartedEvent: &usageeventsv1.UIDiscoverStartedEvent{
					Metadata: &usageeventsv1.DiscoverMetadata{Id: "someid"},
					Status:   &usageeventsv1.DiscoverStepStatus{Status: usageeventsv1.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
				},
			}},
			identityUsername: "myuser",
			errCheck:         require.NoError,
			expected: &prehogv1a.SubmitEventRequest{Event: &prehogv1a.SubmitEventRequest_UiDiscoverStartedEvent{
				UiDiscoverStartedEvent: &prehogv1a.UIDiscoverStartedEvent{
					Metadata: &prehogv1a.DiscoverMetadata{
						Id:       "someid",
						UserName: expectedAnonymizedUserString,
						Sso:      false,
					},
					Status: &prehogv1a.DiscoverStepStatus{Status: prehogv1a.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
				},
			}},
		},
		{
			name: "discover resource selection event",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiDiscoverResourceSelectionEvent{
				UiDiscoverResourceSelectionEvent: &usageeventsv1.UIDiscoverResourceSelectionEvent{
					Metadata: &usageeventsv1.DiscoverMetadata{Id: "someid"},
					Resource: &usageeventsv1.DiscoverResourceMetadata{Resource: usageeventsv1.DiscoverResource_DISCOVER_RESOURCE_SERVER},
					Status:   &usageeventsv1.DiscoverStepStatus{Status: usageeventsv1.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
				},
			}},
			identityUsername: "myuser",
			errCheck:         require.NoError,
			expected: &prehogv1a.SubmitEventRequest{Event: &prehogv1a.SubmitEventRequest_UiDiscoverResourceSelectionEvent{
				UiDiscoverResourceSelectionEvent: &prehogv1a.UIDiscoverResourceSelectionEvent{
					Metadata: &prehogv1a.DiscoverMetadata{
						Id:       "someid",
						UserName: expectedAnonymizedUserString,
						Sso:      false,
					},
					Resource: &prehogv1a.DiscoverResourceMetadata{Resource: prehogv1a.DiscoverResource_DISCOVER_RESOURCE_SERVER},
					Status:   &prehogv1a.DiscoverStepStatus{Status: prehogv1a.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
				},
			}},
		},
		{
			name: "error when discover metadata doesn't have id",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiDiscoverResourceSelectionEvent{
				UiDiscoverResourceSelectionEvent: &usageeventsv1.UIDiscoverResourceSelectionEvent{
					Metadata: &usageeventsv1.DiscoverMetadata{Id: ""},
					Resource: &usageeventsv1.DiscoverResourceMetadata{Resource: usageeventsv1.DiscoverResource_DISCOVER_RESOURCE_SERVER},
					Status:   &usageeventsv1.DiscoverStepStatus{Status: usageeventsv1.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
				},
			}},
			identityUsername: "myuser",
			errCheck: func(tt require.TestingT, err error, i ...interface{}) {
				require.True(tt, trace.IsBadParameter(err), "exepcted trace.IsBadParameter error, got: %v", err)
			},
		},
		{
			name: "error when discover metadata resource",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiDiscoverResourceSelectionEvent{
				UiDiscoverResourceSelectionEvent: &usageeventsv1.UIDiscoverResourceSelectionEvent{
					Metadata: &usageeventsv1.DiscoverMetadata{Id: "someid"},
					Resource: &usageeventsv1.DiscoverResourceMetadata{Resource: 0},
					Status:   &usageeventsv1.DiscoverStepStatus{Status: usageeventsv1.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
				},
			}},
			identityUsername: "myuser",
			errCheck: func(tt require.TestingT, err error, i ...interface{}) {
				require.True(tt, trace.IsBadParameter(err), "exepcted trace.IsBadParameter error, got: %v", err)
			},
		},
		{
			name: "error when discover has stepStatus=ERROR but no error message",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiDiscoverResourceSelectionEvent{
				UiDiscoverResourceSelectionEvent: &usageeventsv1.UIDiscoverResourceSelectionEvent{
					Metadata: &usageeventsv1.DiscoverMetadata{Id: "someid"},
					Resource: &usageeventsv1.DiscoverResourceMetadata{Resource: usageeventsv1.DiscoverResource_DISCOVER_RESOURCE_SERVER},
					Status:   &usageeventsv1.DiscoverStepStatus{Status: usageeventsv1.DiscoverStatus_DISCOVER_STATUS_ERROR},
				},
			}},
			identityUsername: "myuser",
			errCheck: func(tt require.TestingT, err error, i ...interface{}) {
				require.True(tt, trace.IsBadParameter(err), "exepcted trace.IsBadParameter error, got: %v", err)
			},
		},
		{
			name: "when discover has resources count and its values is zero: no error",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiDiscoverAutoDiscoveredResourcesEvent{
				UiDiscoverAutoDiscoveredResourcesEvent: &usageeventsv1.UIDiscoverAutoDiscoveredResourcesEvent{
					Metadata:       &usageeventsv1.DiscoverMetadata{Id: "someid"},
					Resource:       &usageeventsv1.DiscoverResourceMetadata{Resource: usageeventsv1.DiscoverResource_DISCOVER_RESOURCE_SERVER},
					Status:         &usageeventsv1.DiscoverStepStatus{Status: usageeventsv1.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
					ResourcesCount: 0,
				},
			}},
			identityUsername: "myuser",
			errCheck:         require.NoError,
			expected: &prehogv1a.SubmitEventRequest{Event: &prehogv1a.SubmitEventRequest_UiDiscoverAutoDiscoveredResourcesEvent{
				UiDiscoverAutoDiscoveredResourcesEvent: &prehogv1a.UIDiscoverAutoDiscoveredResourcesEvent{
					Metadata: &prehogv1a.DiscoverMetadata{
						Id:       "someid",
						UserName: expectedAnonymizedUserString,
						Sso:      false,
					},
					Resource:       &prehogv1a.DiscoverResourceMetadata{Resource: prehogv1a.DiscoverResource_DISCOVER_RESOURCE_SERVER},
					Status:         &prehogv1a.DiscoverStepStatus{Status: prehogv1a.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
					ResourcesCount: 0,
				},
			}},
		},
		{
			name: "when discover has resources count and its values is positive: no error",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiDiscoverAutoDiscoveredResourcesEvent{
				UiDiscoverAutoDiscoveredResourcesEvent: &usageeventsv1.UIDiscoverAutoDiscoveredResourcesEvent{
					Metadata:       &usageeventsv1.DiscoverMetadata{Id: "someid"},
					Resource:       &usageeventsv1.DiscoverResourceMetadata{Resource: usageeventsv1.DiscoverResource_DISCOVER_RESOURCE_SERVER},
					Status:         &usageeventsv1.DiscoverStepStatus{Status: usageeventsv1.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
					ResourcesCount: 2,
				},
			}},
			identityUsername: "myuser",
			errCheck:         require.NoError,
			expected: &prehogv1a.SubmitEventRequest{Event: &prehogv1a.SubmitEventRequest_UiDiscoverAutoDiscoveredResourcesEvent{
				UiDiscoverAutoDiscoveredResourcesEvent: &prehogv1a.UIDiscoverAutoDiscoveredResourcesEvent{
					Metadata: &prehogv1a.DiscoverMetadata{
						Id:       "someid",
						UserName: expectedAnonymizedUserString,
						Sso:      false,
					},
					Resource:       &prehogv1a.DiscoverResourceMetadata{Resource: prehogv1a.DiscoverResource_DISCOVER_RESOURCE_SERVER},
					Status:         &prehogv1a.DiscoverStepStatus{Status: prehogv1a.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
					ResourcesCount: 2,
				},
			}},
		},
		{
			name: "when discover has resources count and its values is negative: bad parameter error",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiDiscoverAutoDiscoveredResourcesEvent{
				UiDiscoverAutoDiscoveredResourcesEvent: &usageeventsv1.UIDiscoverAutoDiscoveredResourcesEvent{
					Metadata:       &usageeventsv1.DiscoverMetadata{Id: "someid"},
					Resource:       &usageeventsv1.DiscoverResourceMetadata{Resource: usageeventsv1.DiscoverResource_DISCOVER_RESOURCE_SERVER},
					Status:         &usageeventsv1.DiscoverStepStatus{Status: usageeventsv1.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
					ResourcesCount: -2,
				},
			}},
			identityUsername: "myuser",
			errCheck: func(tt require.TestingT, err error, i ...interface{}) {
				require.True(tt, trace.IsBadParameter(err), "exepcted trace.IsBadParameter error, got: %v", err)
			},
		},
		{
			name: "discover started event with sso user",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiDiscoverStartedEvent{
				UiDiscoverStartedEvent: &usageeventsv1.UIDiscoverStartedEvent{
					Metadata: &usageeventsv1.DiscoverMetadata{Id: "someid"},
					Status:   &usageeventsv1.DiscoverStepStatus{Status: usageeventsv1.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
				},
			}},
			identityUsername: "myuser",
			isSSOUser:        true,
			errCheck:         require.NoError,
			expected: &prehogv1a.SubmitEventRequest{Event: &prehogv1a.SubmitEventRequest_UiDiscoverStartedEvent{
				UiDiscoverStartedEvent: &prehogv1a.UIDiscoverStartedEvent{
					Metadata: &prehogv1a.DiscoverMetadata{
						Id:       "someid",
						UserName: expectedAnonymizedUserString,
						Sso:      true,
					},
					Status: &prehogv1a.DiscoverStepStatus{Status: prehogv1a.DiscoverStatus_DISCOVER_STATUS_SUCCESS},
				},
			}},
		},
		{
			name: "integration enroll started event",
			event: &usageeventsv1.UsageEventOneOf{Event: &usageeventsv1.UsageEventOneOf_UiIntegrationEnrollStartEvent{
				UiIntegrationEnrollStartEvent: &usageeventsv1.UIIntegrationEnrollStartEvent{
					Metadata: &usageeventsv1.IntegrationEnrollMetadata{Id: "someid", Kind: usageeventsv1.IntegrationEnrollKind_INTEGRATION_ENROLL_KIND_AWS_OIDC},
				},
			}},
			identityUsername: "myuser",
			errCheck:         require.NoError,
			expected: &prehogv1a.SubmitEventRequest{Event: &prehogv1a.SubmitEventRequest_UiIntegrationEnrollStartEvent{
				UiIntegrationEnrollStartEvent: &prehogv1a.UIIntegrationEnrollStartEvent{
					Metadata: &prehogv1a.IntegrationEnrollMetadata{
						Id:       "someid",
						UserName: expectedAnonymizedUserString,
						Kind:     prehogv1a.IntegrationEnrollKind_INTEGRATION_ENROLL_KIND_AWS_OIDC,
					},
				},
			}},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			tt := tt
			t.Parallel()

			userMD := UserMetadata{
				Username: tt.identityUsername,
				IsSSO:    tt.isSSOUser,
			}
			usageEvent, err := ConvertUsageEvent(tt.event, userMD)
			tt.errCheck(t, err)
			if err != nil {
				return
			}

			got := usageEvent.Anonymize(anonymizer)

			require.Equal(t, tt.expected, &got)
		})
	}
}

func TestEmitEditorChangeEvent(t *testing.T) {
	tt := []struct {
		name           string
		username       string
		prevRoles      []string
		newRoles       []string
		expectedStatus prehogv1a.EditorChangeStatus
	}{
		{
			name:           "Role is granted to user",
			username:       "user1",
			prevRoles:      []string{"role1", "role2"},
			newRoles:       []string{"role1", "role2", teleport.PresetEditorRoleName},
			expectedStatus: prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_GRANTED,
		},
		{
			name:           "Role is removed from user",
			username:       "user2",
			prevRoles:      []string{"role1", "role2", teleport.PresetEditorRoleName},
			newRoles:       []string{"role1", "role2"},
			expectedStatus: prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_REMOVED,
		},
		{
			name:      "Role remains the same",
			username:  "user3",
			prevRoles: []string{"role1", "role2", teleport.PresetEditorRoleName},
			newRoles:  []string{"role1", "role2", teleport.PresetEditorRoleName},
		},
		{
			name:      "Role is not granted or removed",
			username:  "user4",
			prevRoles: []string{"role1", "role2"},
			newRoles:  []string{"role1", "role2"},
		},
		{
			name:           "User is granted the editor role but had other roles",
			username:       "user5",
			prevRoles:      []string{"role1", "role2"},
			newRoles:       []string{"role1", "role2", teleport.PresetEditorRoleName},
			expectedStatus: prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_GRANTED,
		},
		{
			name:           "User is removed from the editor role but still has other roles",
			username:       "user6",
			prevRoles:      []string{"role1", "role2", teleport.PresetEditorRoleName},
			newRoles:       []string{"role1", "role2"},
			expectedStatus: prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_REMOVED,
		},
		{
			name:           "No previous roles, editor role granted",
			username:       "user7",
			prevRoles:      nil,
			newRoles:       []string{teleport.PresetEditorRoleName},
			expectedStatus: prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_GRANTED,
		},
		{
			name:           "Only had editor role, role removed",
			username:       "user8",
			prevRoles:      []string{teleport.PresetEditorRoleName},
			newRoles:       nil,
			expectedStatus: prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_REMOVED,
		},
		{
			name:      "Nil roles",
			username:  "user9",
			prevRoles: nil,
			newRoles:  nil,
		},
		{
			name:           "Granted multiple roles, including editor",
			username:       "user10",
			prevRoles:      []string{"role1", "role2"},
			newRoles:       []string{"role1", "role2", "role3", teleport.PresetEditorRoleName},
			expectedStatus: prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_GRANTED,
		},
		{
			name:           "Removed from multiple roles, including editor",
			username:       "user11",
			prevRoles:      []string{"role1", "role2", "role3", teleport.PresetEditorRoleName},
			newRoles:       []string{"role1", "role2"},
			expectedStatus: prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_REMOVED,
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			var submittedEvents []Anonymizable
			mockSubmit := func(a ...Anonymizable) {
				submittedEvents = append(submittedEvents, a...)
			}

			EmitEditorChangeEvent(tc.username, tc.prevRoles, tc.newRoles, mockSubmit)

			if tc.expectedStatus == prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_GRANTED || tc.expectedStatus == prehogv1a.EditorChangeStatus_EDITOR_CHANGE_STATUS_ROLE_REMOVED {
				require.NotEmpty(t, submittedEvents)
				event, ok := submittedEvents[0].(*EditorChangeEvent)
				require.True(t, ok, "event is not of type *EditorChangeEvent")
				require.Equal(t, tc.expectedStatus, event.Status)
				require.Equal(t, tc.username, event.UserName)
			} else {
				require.Empty(t, submittedEvents, "No event should have been submitted")
			}
		})
	}
}

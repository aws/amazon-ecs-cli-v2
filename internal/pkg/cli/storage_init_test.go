// Copyright Amazon, Inc. or its affiliates. All rights reserved.

package cli

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/cli/mocks"
	"github.com/aws/amazon-ecs-cli-v2/internal/pkg/term/color"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestStorageInitOpts_Validate(t *testing.T) {
	testCases := map[string]struct {
		inAppName     string
		inStorageType string
		inSvcName     string
		inStorageName string
		inAttributes  []string
		inPartition   string
		inSort        string
		inLSISorts    []string

		mockWs    func(m *mocks.MockwsAddonManager)
		mockStore func(m *mocks.Mockstore)

		wantedErr error
	}{
		"no app in workspace": {
			mockWs:    func(m *mocks.MockwsAddonManager) {},
			mockStore: func(m *mocks.Mockstore) {},

			wantedErr: errNoAppInWorkspace,
		},
		"svc not in workspace": {
			mockWs: func(m *mocks.MockwsAddonManager) {
				m.EXPECT().ServiceNames().Return([]string{"bad", "workspace"}, nil)
			},
			mockStore: func(m *mocks.Mockstore) {},

			inAppName:     "bowie",
			inStorageType: s3StorageType,
			inSvcName:     "frontend",
			inStorageName: "my-bucket",
			wantedErr:     errors.New("service frontend not found in the workspace"),
		},
		"workspace error": {
			mockWs: func(m *mocks.MockwsAddonManager) {
				m.EXPECT().ServiceNames().Return(nil, errors.New("wanted err"))
			},
			mockStore: func(m *mocks.Mockstore) {},

			inAppName:     "bowie",
			inStorageType: s3StorageType,
			inSvcName:     "frontend",
			inStorageName: "my-bucket",
			wantedErr:     errors.New("retrieve local service names: wanted err"),
		},
		"happy path s3": {
			mockWs: func(m *mocks.MockwsAddonManager) {
				m.EXPECT().ServiceNames().Return([]string{"frontend"}, nil)
			},
			mockStore:     func(m *mocks.Mockstore) {},
			inAppName:     "bowie",
			inStorageType: s3StorageType,
			inSvcName:     "frontend",
			inStorageName: "my-bucket.4",
			wantedErr:     nil,
		},
		"happy path ddb": {
			mockWs: func(m *mocks.MockwsAddonManager) {
				m.EXPECT().ServiceNames().Return([]string{"frontend"}, nil)
			},
			mockStore:     func(m *mocks.Mockstore) {},
			inAppName:     "bowie",
			inStorageType: dynamoDBStorageType,
			inSvcName:     "frontend",
			inStorageName: "my-cool_table.3",
			wantedErr:     nil,
		},
		"default to ddb name validation when storage type unspecified": {
			mockWs: func(m *mocks.MockwsAddonManager) {
				m.EXPECT().ServiceNames().Return([]string{"frontend"}, nil)
			},
			mockStore:     func(m *mocks.Mockstore) {},
			inAppName:     "bowie",
			inStorageType: "",
			inSvcName:     "frontend",
			inStorageName: "my-cool_table.3",
			wantedErr:     nil,
		},
		"s3 bad character": {
			mockWs: func(m *mocks.MockwsAddonManager) {
				m.EXPECT().ServiceNames().Return([]string{"frontend"}, nil)
			},
			mockStore:     func(m *mocks.Mockstore) {},
			inAppName:     "bowie",
			inStorageType: s3StorageType,
			inSvcName:     "frontend",
			inStorageName: "mybadbucket???",
			wantedErr:     errValueBadFormatWithPeriod,
		},
		"ddb bad character": {
			mockWs: func(m *mocks.MockwsAddonManager) {
				m.EXPECT().ServiceNames().Return([]string{"frontend"}, nil)
			},
			mockStore:     func(m *mocks.Mockstore) {},
			inAppName:     "bowie",
			inStorageType: dynamoDBStorageType,
			inSvcName:     "frontend",
			inStorageName: "badTable!!!",
			wantedErr:     errValueBadFormatWithPeriodUnderscore,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			mockWs := mocks.NewMockwsAddonManager(ctrl)
			mockStore := mocks.NewMockstore(ctrl)
			tc.mockWs(mockWs)
			tc.mockStore(mockStore)
			opts := initStorageOpts{
				initStorageVars: initStorageVars{
					GlobalOpts: &GlobalOpts{
						appName: tc.inAppName,
					},
					storageType:  tc.inStorageType,
					storageName:  tc.inStorageName,
					storageSvc:   tc.inSvcName,
					attributes:   tc.inAttributes,
					partitionKey: tc.inPartition,
					sortKey:      tc.inSort,
					lsiSorts:     tc.inLSISorts,
				},
				ws:    mockWs,
				store: mockStore,
			}

			// WHEN
			err := opts.Validate()

			// THEN
			if tc.wantedErr != nil {
				require.EqualError(t, err, tc.wantedErr.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}

func TestStorageInitOpts_Ask(t *testing.T) {
	const (
		wantedAppName      = "ddos"
		wantedSvcName      = "frontend"
		wantedBucketName   = "coolBucket"
		wantedTableName    = "coolTable"
		wantedPartitionKey = "DogName:S"
		wantedSortKey      = "PhotoId:N"
	)
	testCases := map[string]struct {
		inAppName     string
		inStorageType string
		inSvcName     string
		inStorageName string
		inAttributes  []string
		inPartition   string
		inSort        string
		inLSISorts    []string
		inNoLsi       bool
		inNoSort      bool

		mockPrompt func(m *mocks.Mockprompter)
		mockCfg    func(m *mocks.MockconfigSelector)

		wantedErr error
	}{
		"Asks for storage type": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageName: wantedBucketName,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().SelectOne(gomock.Any(), gomock.Any(), gomock.Eq(storageTypes), gomock.Any()).Return(s3StorageType, nil)
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
		"error if storage type not gotten": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageName: wantedBucketName,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().SelectOne(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("select storage type: some error"),
		},
		"asks for storage svc": {
			inAppName:     wantedAppName,
			inStorageName: wantedBucketName,
			inStorageType: s3StorageType,

			mockPrompt: func(m *mocks.Mockprompter) {},
			mockCfg: func(m *mocks.MockconfigSelector) {
				m.EXPECT().Service(gomock.Eq(storageInitSvcPrompt), gomock.Any(), gomock.Eq(wantedAppName)).Return(wantedSvcName, nil)
			},

			wantedErr: nil,
		},
		"error if svc not returned": {
			inAppName:     wantedAppName,
			inStorageName: wantedBucketName,
			inStorageType: s3StorageType,

			mockPrompt: func(m *mocks.Mockprompter) {},
			mockCfg: func(m *mocks.MockconfigSelector) {
				m.EXPECT().Service(gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("some error"))
			},

			wantedErr: fmt.Errorf("retrieve local service names: some error"),
		},
		"asks for storage name": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: s3StorageType,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Get(gomock.Eq(
					fmt.Sprintf(fmtStorageInitNamePrompt,
						color.HighlightUserInput(s3BucketFriendlyText),
					),
				),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(wantedBucketName, nil)
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
		"error if storage name not returned": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: s3StorageType,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Get(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return("", errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("input storage name: some error"),
		},
		"asks for partition key if not specified": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inSort:        wantedSortKey,
			inNoLsi:       true,

			mockPrompt: func(m *mocks.Mockprompter) {
				keyPrompt := fmt.Sprintf(fmtStorageInitDDBKeyPrompt,
					color.HighlightUserInput("partition key"),
					color.HighlightUserInput(dynamoDBStorageType),
				)
				keyTypePrompt := fmt.Sprintf(fmtStorageInitDDBKeyTypePrompt, ddbKeyString)
				m.EXPECT().Get(gomock.Eq(keyPrompt),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(wantedPartitionKey, nil)
				m.EXPECT().SelectOne(gomock.Eq(keyTypePrompt),
					gomock.Any(),
					attributeTypesLong,
					gomock.Any(),
				).Return(ddbStringType, nil)
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
		"error if fail to return partition key": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Get(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return("", errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("get DDB partition key: some error"),
		},
		"error if fail to return partition key type": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inSort:        wantedSortKey,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Get(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(wantedPartitionKey, nil)
				m.EXPECT().SelectOne(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return("", errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("get DDB partition key datatype: some error"),
		},
		"ask for sort key if not specified": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inNoLsi:       true,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBSortKeyConfirm),
					gomock.Any(),
					gomock.Any(),
				).Return(true, nil)
				keyPrompt := fmt.Sprintf(fmtStorageInitDDBKeyPrompt,
					color.HighlightUserInput("sort key"),
					color.HighlightUserInput(dynamoDBStorageType),
				)
				keyTypePrompt := fmt.Sprintf(fmtStorageInitDDBKeyTypePrompt, ddbKeyString)
				m.EXPECT().Get(gomock.Eq(keyPrompt),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(wantedPartitionKey, nil)
				m.EXPECT().SelectOne(gomock.Eq(keyTypePrompt),
					gomock.Any(),
					attributeTypesLong,
					gomock.Any(),
				).Return(ddbStringType, nil)
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
		"error if fail to confirm add sort key": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBSortKeyConfirm),
					gomock.Any(),
					gomock.Any(),
				).Return(false, errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("confirm DDB sort key: some error"),
		},
		"error if fail to return sort key": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBSortKeyConfirm),
					gomock.Any(),
					gomock.Any(),
				).Return(true, nil)
				m.EXPECT().Get(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return("", errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("get DDB sort key: some error"),
		},
		"error if fail to return sort key type": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBSortKeyConfirm),
					gomock.Any(),
					gomock.Any(),
				).Return(true, nil)
				m.EXPECT().Get(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(wantedPartitionKey, nil)
				m.EXPECT().SelectOne(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return("", errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("get DDB sort key datatype: some error"),
		},
		"don't ask for sort key if no-sort specified": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inNoSort:      true,
			inNoLsi:       true,

			mockPrompt: func(m *mocks.Mockprompter) {},
			mockCfg:    func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
		"ok if --no-lsi and --sort-key are both specified": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,
			inNoLsi:       true,

			mockPrompt: func(m *mocks.Mockprompter) {},
			mockCfg:    func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
		"don't ask about LSI if no-sort is specified": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inNoSort:      true,

			mockPrompt: func(m *mocks.Mockprompter) {},
			mockCfg:    func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
		"asks for attributes": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBMoreAttributesPrompt),
					gomock.Any(),
					gomock.Any(),
				).Return(true, nil)
				attributeTypePrompt := fmt.Sprintf(fmtStorageInitDDBKeyTypePrompt, color.Emphasize(ddbAttributeString))
				m.EXPECT().Get(gomock.Eq(storageInitDDBAttributePrompt),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return("points", nil)
				m.EXPECT().SelectOne(gomock.Eq(attributeTypePrompt),
					gomock.Any(),
					attributeTypesLong,
					gomock.Any(),
				).Return(ddbIntType, nil)
				// Stop adding attributes now
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBMoreAttributesPrompt),
					gomock.Any(),
					gomock.Any(),
				).Return(false, nil)
				// Don't add an LSI.
				m.EXPECT().Confirm(gomock.Any(), gomock.Any(), gomock.Any()).Return(false, nil)
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
		"error if attributes misspecified": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBMoreAttributesPrompt),
					gomock.Any(),
					gomock.Any(),
				).Return(true, nil)
				m.EXPECT().Get(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return("", errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("get DDB attribute name: some error"),
		},
		"errors if fail to confirm attribute type": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(false, errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("confirm add more attributes: some error"),
		},
		"error if attribute type misspecified": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return(true, nil)
				m.EXPECT().Get(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return("cool", nil)
				m.EXPECT().SelectOne(gomock.Any(),
					gomock.Any(),
					gomock.Any(),
					gomock.Any(),
				).Return("", errors.New("some error"))

			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("get DDB attribute type: some error"),
		},
		"asks for LSI configuration": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,
			inAttributes:  []string{"email:S", "points:N"},

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBLSIPrompt),
					gomock.Any(),
					gomock.Any(),
				).Return(true, nil)
				m.EXPECT().MultiSelect(
					gomock.Eq(storageInitDDBLSINamePrompt),
					gomock.Any(),
					[]string{"email", "points"},
					gomock.Any(),
				).Return([]string{"email"}, nil)
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
		"error if more than 5 lsis specified": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,
			inAttributes:  []string{"email:S", "points:N", "awesomeness:N", "badness:N", "heart:N", "justice:N"},

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBLSIPrompt),
					gomock.Any(),
					gomock.Any(),
				).Return(true, nil)
				m.EXPECT().MultiSelect(
					gomock.Eq(storageInitDDBLSINamePrompt),
					gomock.Any(),
					[]string{"email", "points", "awesomeness", "badness", "heart", "justice"},
					gomock.Any(),
				).Return([]string{"email", "points", "awesomeness", "badness", "heart", "justice"}, nil)
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("cannot specify more than 5 alternate sort keys"),
		},
		"error if LSIs specified incorrectly": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,
			inAttributes:  []string{"email:S", "points:N", "awesomeness:N", "badness:N", "heart:N", "justice:N"},

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBLSIPrompt),
					gomock.Any(),
					gomock.Any(),
				).Return(true, nil)
				m.EXPECT().MultiSelect(
					gomock.Eq(storageInitDDBLSINamePrompt),
					gomock.Any(),
					[]string{"email", "points", "awesomeness", "badness", "heart", "justice"},
					gomock.Any(),
				).Return([]string{}, errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("get LSI sort keys: some error"),
		},
		"error if LSIs not confirmed": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,
			inAttributes:  []string{"email:S", "points:N", "awesomeness:N", "badness:N", "heart:N", "justice:N"},

			mockPrompt: func(m *mocks.Mockprompter) {
				m.EXPECT().Confirm(
					gomock.Eq(storageInitDDBLSIPrompt),
					gomock.Any(),
					gomock.Any(),
				).Return(false, errors.New("some error"))
			},
			mockCfg: func(m *mocks.MockconfigSelector) {},

			wantedErr: fmt.Errorf("confirm add LSI to table: some error"),
		},
		"no error or asks when fully specified": {
			inAppName:     wantedAppName,
			inSvcName:     wantedSvcName,
			inStorageType: dynamoDBStorageType,
			inStorageName: wantedTableName,
			inPartition:   wantedPartitionKey,
			inSort:        wantedSortKey,
			inAttributes:  []string{"email:S", "points:N"},
			inLSISorts:    []string{"email"},

			mockPrompt: func(m *mocks.Mockprompter) {},
			mockCfg:    func(m *mocks.MockconfigSelector) {},

			wantedErr: nil,
		},
	}
	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockPrompt := mocks.NewMockprompter(ctrl)
			mockConfig := mocks.NewMockconfigSelector(ctrl)
			opts := initStorageOpts{
				initStorageVars: initStorageVars{
					GlobalOpts: &GlobalOpts{
						appName: tc.inAppName,
						prompt:  mockPrompt,
					},
					storageType:  tc.inStorageType,
					storageName:  tc.inStorageName,
					storageSvc:   tc.inSvcName,
					attributes:   tc.inAttributes,
					partitionKey: tc.inPartition,
					sortKey:      tc.inSort,
					lsiSorts:     tc.inLSISorts,
					noLsi:        tc.inNoLsi,
					noSort:       tc.inNoSort,
				},
				sel: mockConfig,
			}
			tc.mockPrompt(mockPrompt)
			tc.mockCfg(mockConfig)
			// WHEN
			err := opts.Ask()

			// THEN
			if tc.wantedErr != nil {
				require.EqualError(t, err, tc.wantedErr.Error())
			} else {
				require.Nil(t, err)
			}
		})
	}
}

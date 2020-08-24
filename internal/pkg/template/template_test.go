// Copyright Amazon.com, Inc. or its affiliates. All Rights Reserved.
// SPDX-License-Identifier: Apache-2.0

package template

import (
	"errors"
	"fmt"
	"testing"

	"github.com/gobuffalo/packd"
	"github.com/stretchr/testify/require"
)

func TestTemplate_Read(t *testing.T) {
	testCases := map[string]struct {
		inPath           string
		mockDependencies func(t *Template)

		wantedContent string
		wantedErr     error
	}{
		"template does not exist": {
			inPath: "/fake/manifest.yml",
			mockDependencies: func(t *Template) {
				mockBox := packd.NewMemoryBox()
				t.box = mockBox
			},

			wantedErr: errors.New("read template /fake/manifest.yml"),
		},
		"returns content": {
			inPath: "/fake/manifest.yml",
			mockDependencies: func(t *Template) {
				mockBox := packd.NewMemoryBox()
				mockBox.AddString("/fake/manifest.yml", "hello")
				t.box = mockBox
			},

			wantedContent: "hello",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			tpl := &Template{}
			tc.mockDependencies(tpl)

			// WHEN
			c, err := tpl.Read(tc.inPath)

			if tc.wantedErr != nil {
				require.Contains(t, err.Error(), tc.wantedErr.Error())
			} else {
				require.Equal(t, tc.wantedContent, c.String())
			}
		})
	}
}

func TestTemplate_Parse(t *testing.T) {
	testCases := map[string]struct {
		inPath           string
		inData           interface{}
		mockDependencies func(t *Template)

		wantedContent string
		wantedErr     error
	}{
		"template does not exist": {
			inPath: "/fake/manifest.yml",
			mockDependencies: func(t *Template) {
				mockBox := packd.NewMemoryBox()
				t.box = mockBox
			},

			wantedErr: errors.New("read template /fake/manifest.yml"),
		},
		"template cannot be parsed": {
			inPath: "/fake/manifest.yml",
			mockDependencies: func(t *Template) {
				mockBox := packd.NewMemoryBox()
				mockBox.AddString("/fake/manifest.yml", `{{}}`)
				t.box = mockBox
			},

			wantedErr: errors.New("parse template /fake/manifest.yml"),
		},
		"template cannot be executed": {
			inPath: "/fake/manifest.yml",
			inData: struct{}{},
			mockDependencies: func(t *Template) {
				mockBox := packd.NewMemoryBox()
				mockBox.AddString("/fake/manifest.yml", `{{.Name}}`)
				t.box = mockBox
			},

			wantedErr: fmt.Errorf("execute template %s with data %v", "/fake/manifest.yml", struct{}{}),
		},
		"valid template": {
			inPath: "/fake/manifest.yml",
			inData: struct {
				Name string
			}{
				Name: "webhook",
			},
			mockDependencies: func(t *Template) {
				mockBox := packd.NewMemoryBox()
				mockBox.AddString("/fake/manifest.yml", `{{.Name}}`)
				t.box = mockBox
			},

			wantedContent: "webhook",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			// GIVEN
			tpl := &Template{}
			tc.mockDependencies(tpl)

			// WHEN
			c, err := tpl.Parse(tc.inPath, tc.inData)

			if tc.wantedErr != nil {
				require.Contains(t, err.Error(), tc.wantedErr.Error())
			} else {
				require.Equal(t, tc.wantedContent, c.String())
			}
		})
	}
}

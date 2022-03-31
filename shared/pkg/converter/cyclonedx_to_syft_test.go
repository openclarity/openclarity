// Copyright © 2022 Cisco Systems, Inc. and its affiliates.
// All rights reserved.
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

package converter

import (
	"reflect"
	"testing"

	cdx "github.com/CycloneDX/cyclonedx-go"
	"github.com/anchore/syft/syft/linux"
	syft_pkg "github.com/anchore/syft/syft/pkg"
	syft_cpe "github.com/anchore/syft/syft/pkg/cataloger/common/cpe"
	syft_sbom "github.com/anchore/syft/syft/sbom"
	syft_source "github.com/anchore/syft/syft/source"
	purl "github.com/package-url/packageurl-go"
)

func TestConvertCycloneDXtoSyft(t *testing.T) {
	bom := &cdx.BOM{
		Metadata: &cdx.Metadata{
			Component: &cdx.Component{
				Type: cdx.ComponentTypeContainer,
				Name: "testCDX",
			},
		},
		Components: &[]cdx.Component{
			{
				Type: cdx.ComponentTypeOS,
				Properties: &[]cdx.Property{
					{
						Name:  "id",
						Value: "alpine",
					},
					{
						Name:  "versionID",
						Value: "1.13",
					},
				},
			},
			{
				Name:       "test",
				Version:    "1.0.0",
				Type:       cdx.ComponentTypeLibrary,
				PackageURL: "pkg:golang/test.org/test@v1.0.0",
				Properties: &[]cdx.Property{
					{
						Name:  "path",
						Value: "/test",
					},
					{
						Name:  "layerID",
						Value: "1111",
					},
				},
			},
			{
				Name:       "test-2",
				Version:    "1.1.0",
				Type:       cdx.ComponentTypeLibrary,
				PackageURL: "pkg:golang/test.org/test-2@v1.0.0",
				Properties: &[]cdx.Property{
					{
						Name:  "path",
						Value: "/test-2",
					},
					{
						Name:  "layerID",
						Value: "2222",
					},
				},
			},
		},
	}

	syftSbom := syft_sbom.SBOM{
		Artifacts: syft_sbom.Artifacts{
			PackageCatalog: syft_pkg.NewCatalog(),
			LinuxDistribution: &linux.Release{
				ID:        "alpine",
				VersionID: "1.13",
			},
		},
		Source: syft_source.Metadata{
			Scheme:        syft_source.ImageScheme,
			ImageMetadata: syft_source.ImageMetadata{},
			Path:          "testCDX",
		},
	}
	pkg1 := syft_pkg.Package{
		Name:         "test",
		Version:      "1.0.0",
		Type:         syft_pkg.GoModulePkg,
		PURL:         "pkg:golang/test.org/test@v1.0.0",
		MetadataType: syft_pkg.GolangBinMetadataType,
		Metadata:     emptyMetadata{},
		Locations: []syft_source.Location{
			{
				Coordinates: syft_source.Coordinates{
					RealPath:     "/test",
					FileSystemID: "1111",
				},
			},
		},
		Language: syft_pkg.Go,
	}
	pkg1.CPEs = syft_cpe.Generate(pkg1)
	pkg2 := syft_pkg.Package{
		Name:         "test-2",
		Version:      "1.1.0",
		Type:         syft_pkg.GoModulePkg,
		PURL:         "pkg:golang/test.org/test-2@v1.0.0",
		MetadataType: syft_pkg.GolangBinMetadataType,
		Metadata:     emptyMetadata{},
		Locations: []syft_source.Location{
			{
				Coordinates: syft_source.Coordinates{
					RealPath:     "/test-2",
					FileSystemID: "2222",
				},
			},
		},
		Language: syft_pkg.Go,
	}

	pkg2.CPEs = syft_cpe.Generate(pkg2)
	syftSbom.Artifacts.PackageCatalog.Add(pkg1)
	syftSbom.Artifacts.PackageCatalog.Add(pkg2)

	type args struct {
		bom *cdx.BOM
	}
	tests := []struct {
		name    string
		args    args
		want    syft_sbom.SBOM
		wantErr bool
	}{
		{
			name: "input is nil",
			args: args{
				bom: nil,
			},
			want:    syft_sbom.SBOM{},
			wantErr: true,
		},
		{
			name: "metadata is nil",
			args: args{
				bom: &cdx.BOM{},
			},
			want:    syft_sbom.SBOM{},
			wantErr: true,
		},
		{
			name: "metadata component is nil",
			args: args{
				bom: &cdx.BOM{
					Metadata: &cdx.Metadata{},
				},
			},
			want:    syft_sbom.SBOM{},
			wantErr: true,
		},
		{
			name: "successful conversion",
			args: args{
				bom: bom,
			},
			want:    syftSbom,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := convertCycloneDXtoSyft(tt.args.bom)
			if (err != nil) != tt.wantErr {
				t.Errorf("convertCycloneDXtoSyft() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				for _, p := range got.Artifacts.PackageCatalog.Sorted() {
					t.Logf("Packages got: %v", p)
				}
				for _, p := range tt.want.Artifacts.PackageCatalog.Sorted() {
					t.Logf("Packages want: %v", p)
				}
				t.Errorf("convertCycloneDXtoSyft() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_determineSourceMeta(t *testing.T) {
	type args struct {
		packageURL string
	}
	tests := []struct {
		name  string
		args  args
		want  syft_pkg.MetadataType
		want1 syft_pkg.Type
		want2 syft_pkg.Language
	}{
		{
			name: "get packageURL from string failed",
			args: args{
				packageURL: "very-bad-packageurl",
			},
			want:  "",
			want1: "",
			want2: "",
		},
		{
			name: "golang source",
			args: args{
				packageURL: createFakePackageURL("golang"),
			},
			want:  syft_pkg.GolangBinMetadataType,
			want1: syft_pkg.GoModulePkg,
			want2: syft_pkg.Go,
		},
		{
			name: "pypi source",
			args: args{
				packageURL: createFakePackageURL("pypi"),
			},
			want:  syft_pkg.PythonPackageMetadataType,
			want1: syft_pkg.PythonPkg,
			want2: syft_pkg.Python,
		},
		{
			name: "npm source",
			args: args{
				packageURL: createFakePackageURL("npm"),
			},
			want:  syft_pkg.NpmPackageJSONMetadataType,
			want1: syft_pkg.NpmPkg,
			want2: syft_pkg.JavaScript,
		},
		{
			name: "gem source",
			args: args{
				packageURL: createFakePackageURL("gem"),
			},
			want:  syft_pkg.GemMetadataType,
			want1: syft_pkg.GemPkg,
			want2: syft_pkg.Ruby,
		},
		{
			name: "cargo source",
			args: args{
				packageURL: createFakePackageURL("cargo"),
			},
			want:  syft_pkg.RustCargoPackageMetadataType,
			want1: syft_pkg.RustPkg,
			want2: syft_pkg.Rust,
		},
		{
			name: "maven source",
			args: args{
				packageURL: createFakePackageURL("maven"),
			},
			want:  syft_pkg.JavaMetadataType,
			want1: syft_pkg.JavaPkg,
			want2: syft_pkg.Java,
		},
		{
			name: "alpine source",
			args: args{
				packageURL: createFakePackageURL("alpine"),
			},
			want:  syft_pkg.ApkMetadataType,
			want1: syft_pkg.ApkPkg,
			want2: "",
		},
		{
			name: "rpm source",
			args: args{
				packageURL: createFakePackageURL("rpm"),
			},
			want:  syft_pkg.RpmdbMetadataType,
			want1: syft_pkg.RpmPkg,
			want2: "",
		},
		{
			name: "deb source",
			args: args{
				packageURL: createFakePackageURL("deb"),
			},
			want:  syft_pkg.DpkgMetadataType,
			want1: syft_pkg.DebPkg,
			want2: "",
		},
		{
			name: "unknown source",
			args: args{
				packageURL: createFakePackageURL("fakeLang"),
			},
			want:  syft_pkg.UnknownMetadataType,
			want1: syft_pkg.UnknownPkg,
			want2: syft_pkg.UnknownLanguage,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, got1, got2 := determineSourceMeta(tt.args.packageURL)
			if got != tt.want {
				t.Errorf("determineSourceMeta() got = %v, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("determineSourceMeta() got1 = %v, want %v", got1, tt.want1)
			}
			if got2 != tt.want2 {
				t.Errorf("determineSourceMeta() got2 = %v, want %v", got2, tt.want2)
			}
		})
	}
}

// nolint:cyclop
func createFakePackageURL(language string) string {
	switch language {
	case "golang":
		return purl.NewPackageURL(
			"golang",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	case "pypi":
		return purl.NewPackageURL(
			"pypi",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	case "npm":
		return purl.NewPackageURL(
			"npm",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	case "gem":
		return purl.NewPackageURL(
			"gem",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	case "cargo":
		return purl.NewPackageURL(
			"cargo",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	case "maven":
		return purl.NewPackageURL(
			"maven",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	case "alpine":
		return purl.NewPackageURL(
			"alpine",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	case "deb":
		return purl.NewPackageURL(
			"deb",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	case "rpm":
		return purl.NewPackageURL(
			"rpm",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	default:
		return purl.NewPackageURL(
			"unknown",
			"github.com/test",
			"test",
			"1.0.0",
			purl.Qualifiers{},
			"").ToString()
	}
}

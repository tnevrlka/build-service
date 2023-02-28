/*
Copyright 2021-2022 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"reflect"
	"testing"

	"github.com/redhat-appstudio/application-service/gitops"
	corev1 "k8s.io/api/core/v1"

	tektonapi "github.com/tektoncd/pipeline/pkg/apis/pipeline/v1beta1"
)

func TestGetContainerImageRepository(t *testing.T) {
	tests := []struct {
		name  string
		image string
		want  string
	}{
		{
			name:  "should not change image",
			image: "image-name",
			want:  "image-name",
		},
		{
			name:  "should not change /user/image",
			image: "user/image",
			want:  "user/image",
		},
		{
			name:  "should not change repository.io/user/image",
			image: "repository.io/user/image",
			want:  "repository.io/user/image",
		},
		{
			name:  "should delete tag",
			image: "repository.io/user/image:tag",
			want:  "repository.io/user/image",
		},
		{
			name:  "should delete sha",
			image: "repository.io/user/image@sha256:586ab46b9d6d906b2df3dad12751e807bd0f0632d5a2ab3991bdac78bdccd59a",
			want:  "repository.io/user/image",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getContainerImageRepository(tt.image)
			if got != tt.want {
				t.Errorf("getContainerImageRepository(): got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMergeAndSortTektonParams(t *testing.T) {
	tests := []struct {
		name       string
		existing   []tektonapi.Param
		additional []tektonapi.Param
		want       []tektonapi.Param
	}{
		{
			name: "should merge two different parameters lists",
			existing: []tektonapi.Param{
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "main"}},
			},
			additional: []tektonapi.Param{
				{Name: "dockerfile", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "docker/Dockerfile"}},
				{Name: "rebuild", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "true"}},
			},
			want: []tektonapi.Param{
				{Name: "dockerfile", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "docker/Dockerfile"}},
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
				{Name: "rebuild", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "true"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "main"}},
			},
		},
		{
			name: "should append empty parameters list",
			existing: []tektonapi.Param{
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "main"}},
			},
			additional: []tektonapi.Param{},
			want: []tektonapi.Param{
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "main"}},
			},
		},
		{
			name: "should sort parameters list",
			existing: []tektonapi.Param{
				{Name: "rebuild", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "true"}},
				{Name: "dockerfile", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "docker/Dockerfile"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "main"}},
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
			},
			additional: []tektonapi.Param{},
			want: []tektonapi.Param{
				{Name: "dockerfile", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "docker/Dockerfile"}},
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
				{Name: "rebuild", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "true"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "main"}},
			},
		},
		{
			name: "should override existing parameters",
			existing: []tektonapi.Param{
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "main"}},
				{Name: "rebuild", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "false"}},
			},
			additional: []tektonapi.Param{
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "2378a064bf6b66a8ffc650ad88d404cca24ade29"}},
				{Name: "rebuild", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "true"}},
			},
			want: []tektonapi.Param{
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
				{Name: "rebuild", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "true"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "2378a064bf6b66a8ffc650ad88d404cca24ade29"}},
			},
		},
		{
			name: "should append and override parameters",
			existing: []tektonapi.Param{
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "main"}},
			},
			additional: []tektonapi.Param{
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "2378a064bf6b66a8ffc650ad88d404cca24ade29"}},
				{Name: "rebuild", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "true"}},
			},
			want: []tektonapi.Param{
				{Name: "git-url", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "https://githost.com/user/repo.git"}},
				{Name: "rebuild", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "true"}},
				{Name: "revision", Value: tektonapi.ArrayOrString{Type: "string", StringVal: "2378a064bf6b66a8ffc650ad88d404cca24ade29"}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := mergeAndSortTektonParams(tt.existing, tt.additional)
			if len(got) != len(tt.want) {
				t.Errorf("mergeAndSortTektonParams(): got %v, want %v", got, tt.want)
			}
			for i := range got {
				if !reflecfunc TestGetGitProvider(t *testing.T) {
					type args struct {
						ctx    context.Context
						gitURL string
					}
					tests := []struct {
						name       string
						args       args
						wantErr    bool
						wantString string
					}{
						{
							name: "github ssh",
							args: args{
								ctx:    context.Background(),
								gitURL: "git@github.com:redhat-appstudio/application-service.git",
							},
							wantErr:    false,
							wantString: "https://github.com",
						},
						{
							name: "gitlab ssh",
							args: args{
								ctx:    context.Background(),
								gitURL: "git@gitlab.com:namespace/project-name.git",
							},
							wantErr:    false,
							wantString: "https://gitlab.com",
						},
						{
							name: "bitbucket ssh",
							args: args{
								ctx:    context.Background(),
								gitURL: "git@bitbucket.org:organization/project-name.git",
							},
							wantErr:    false,
							wantString: "https://bitbucket.org",
						},
						{
							name: "github https",
							args: args{
								ctx:    context.Background(),
								gitURL: "https://github.com/redhat-appstudio/application-service.git",
							},
							wantErr:    false,
							wantString: "https://github.com",
						},
						{
							name: "bitbucket https",
							args: args{
								ctx:    context.Background(),
								gitURL: "https://sbose78@bitbucket.org/sbose78/appstudio.git",
							},
							wantErr:    false,
							wantString: "https://bitbucket.org",
						},
						{
							name: "no scheme",
							args: args{
								ctx:    context.Background(),
								gitURL: "github.com/redhat-appstudio/application-service.git",
							},
							wantErr:    true, //fully qualified URL is a must
							wantString: "",
						},
						{
							name: "invalid url",
							args: args{
								ctx:    context.Background(),
								gitURL: "not-even-a-url",
							},
							wantErr:    true, //fully qualified URL is a must
							wantString: "",
						},
					}
					for _, tt := range tests {
						t.Run(tt.name, func(t *testing.T) {
							if got, err := getGitProviderUrl(tt.args.gitURL); (got != tt.wantString) ||
								(tt.wantErr == true && err == nil) ||
								(tt.wantErr == false && err != nil) {
								t.Errorf("UpdateServiceAccountIfSecretNotLinked() Got Error: = %v, want %v ; Got String:  = %v , want %v", err, tt.wantErr, got, tt.wantString)
							}
						})
					}
				}
				
				func TestUpdateServiceAccountIfSecretNotLinked(t *testing.T) {
					type args struct {
						gitSecretName  string
						serviceAccount *corev1.ServiceAccount
					}
					tests := []struct {
						name string
						args args
						want bool
					}{
						{
							name: "present",
							args: args{
								gitSecretName: "present",
								serviceAccount: &corev1.ServiceAccount{
									Secrets: []corev1.ObjectReference{
										{
											Name: "present",
										},
									},
								},
							},
							want: false, // since it was present, this implies the SA wasn't updated.
						},
						{
							name: "not present",
							args: args{
								gitSecretName: "not-present",
								serviceAccount: &corev1.ServiceAccount{
									Secrets: []corev1.ObjectReference{
										{
											Name: "something-else",
										},
									},
								},
							},
							want: true, // since it wasn't present, this implies the SA was updated.
						},
						{
							name: "secretname is empty string",
							args: args{
								gitSecretName: "",
								serviceAccount: &corev1.ServiceAccount{
									Secrets: []corev1.ObjectReference{
										{
											Name: "present",
										},
									},
								},
							},
							want: false, // since it wasn't present, this implies the SA was updated.
						},
					}
					for _, tt := range tests {
						t.Run(tt.name, func(t *testing.T) {
							if got := updateServiceAccountIfSecretNotLinked(tt.args.gitSecretName, tt.args.serviceAccount); got != tt.want {
								t.Errorf("UpdateServiceAccountIfSecretNotLinked() = %v, want %v", got, tt.want)
							}
						})
					}
				}
				
				func TestValidatePaCConfiguration(t *testing.T) {
					const ghAppPrivateKeyStub = "-----BEGIN RSA PRIVATE KEY-----_key-content_-----END RSA PRIVATE KEY-----"
				
					tests := []struct {
						name        string
						gitProvider string
						config      map[string][]byte
						expectError bool
					}{
						{
							name:        "should accept GitHub application configuration",
							gitProvider: "github",
							config: map[string][]byte{
								gitops.PipelinesAsCode_githubAppIdKey:   []byte("12345"),
								gitops.PipelinesAsCode_githubPrivateKey: []byte(ghAppPrivateKeyStub),
							},
							expectError: false,
						},
						{
							name:        "should accept GitHub application configuration with end line",
							gitProvider: "github",
							config: map[string][]byte{
								gitops.PipelinesAsCode_githubAppIdKey:   []byte("12345"),
								gitops.PipelinesAsCode_githubPrivateKey: []byte(ghAppPrivateKeyStub + "\n"),
							},
							expectError: false,
						},
						{
							name:        "should accept GitHub webhook configuration",
							gitProvider: "github",
							config: map[string][]byte{
								"github.token": []byte("ghp_token"),
							},
							expectError: false,
						},
						{
							name:        "should reject empty GitHub webhook token",
							gitProvider: "github",
							config: map[string][]byte{
								"github.token": []byte(""),
							},
							expectError: true,
						},
						{
							name:        "should accept GitHub application configuration if both GitHub application and webhook configured",
							gitProvider: "github",
							config: map[string][]byte{
								gitops.PipelinesAsCode_githubAppIdKey:   []byte("12345"),
								gitops.PipelinesAsCode_githubPrivateKey: []byte(ghAppPrivateKeyStub),
								"github.token":                          []byte("ghp_token"),
								"gitlab.token":                          []byte("token"),
							},
							expectError: false,
						},
						{
							name:        "should reject GitHub application configuration if GitHub id is missing",
							gitProvider: "github",
							config: map[string][]byte{
								gitops.PipelinesAsCode_githubPrivateKey: []byte(ghAppPrivateKeyStub),
								"github.token":                          []byte("ghp_token"),
							},
							expectError: true,
						},
						{
							name:        "should reject GitHub application configuration if GitHub application private key is missing",
							gitProvider: "github",
							config: map[string][]byte{
								gitops.PipelinesAsCode_githubAppIdKey: []byte("12345"),
								"github.token":                        []byte("ghp_token"),
							},
							expectError: true,
						},
						{
							name:        "should reject GitHub application configuration if GitHub application id is invalid",
							gitProvider: "github",
							config: map[string][]byte{
								gitops.PipelinesAsCode_githubAppIdKey:   []byte("12ab"),
								gitops.PipelinesAsCode_githubPrivateKey: []byte(ghAppPrivateKeyStub),
							},
							expectError: true,
						},
						{
							name:        "should reject GitHub application configuration if GitHub application application private key is invalid",
							gitProvider: "github",
							config: map[string][]byte{
								gitops.PipelinesAsCode_githubAppIdKey:   []byte("12345"),
								gitops.PipelinesAsCode_githubPrivateKey: []byte("private-key"),
							},
							expectError: true,
						},
						{
							name:        "should accept GitLab webhook configuration",
							gitProvider: "gitlab",
							config: map[string][]byte{
								"gitlab.token": []byte("token"),
							},
							expectError: false,
						},
						{
							name:        "should accept GitLab webhook configuration even if other providers configured",
							gitProvider: "gitlab",
							config: map[string][]byte{
								gitops.PipelinesAsCode_githubAppIdKey:   []byte("12345"),
								gitops.PipelinesAsCode_githubPrivateKey: []byte(ghAppPrivateKeyStub),
								"github.token":                          []byte("ghp_token"),
								"gitlab.token":                          []byte("token"),
								"bitbucket.token":                       []byte("token2"),
							},
							expectError: false,
						},
						{
							name:        "should reject empty GitLab webhook token",
							gitProvider: "gitlab",
							config: map[string][]byte{
								"gitlab.token": []byte(""),
							},
							expectError: true,
						},
						{
							name:        "should accept Bitbucket webhook configuration even if other providers configured",
							gitProvider: "bitbucket",
							config: map[string][]byte{
								gitops.PipelinesAsCode_githubAppIdKey:   []byte("12345"),
								gitops.PipelinesAsCode_githubPrivateKey: []byte(ghAppPrivateKeyStub),
								"github.token":                          []byte("ghp_token"),
								"gitlab.token":                          []byte("token"),
								"bitbucket.token":                       []byte("token2"),
								"username":                              []byte("user"),
							},
							expectError: false,
						},
						{
							name:        "should reject empty Bitbucket webhook token",
							gitProvider: "gitlab",
							config: map[string][]byte{
								"bitbucket.token": []byte(""),
								"username":        []byte("user"),
							},
							expectError: true,
						},
						{
							name:        "should reject Bitbucket webhook configuration with empty username",
							gitProvider: "gitlab",
							config: map[string][]byte{
								"bitbucket.token": []byte("token"),
							},
							expectError: true,
						},
					}
				
					for _, tt := range tests {
						t.Run(tt.name, func(t *testing.T) {
							err := validatePaCConfiguration(tt.gitProvider, tt.config)
							if err != nil {
								if !tt.expectError {
									t.Errorf("Expected that the configuration %#v from provider %s should be valid", tt.config, tt.gitProvider)
								}
							} else {
								if tt.expectError {
									t.Errorf("Expected that the configuration %#v from provider %s fails", tt.config, tt.gitProvider)
								}
							}
						})
					}
				}
				
				func TestGeneratePaCWebhookSecretString(t *testing.T) {
					expectedSecretStringLength := 20 * 2 // each byte is represented by 2 hex chars
				
					t.Run("should be able to generate webhook secret string", func(t *testing.T) {
						secret := generatePaCWebhookSecretString()
						if len(secret) != expectedSecretStringLength {
							t.Errorf("Expected that webhook secret string has length %d, but got %d", expectedSecretStringLength, len(secret))
						}
					})
				
					t.Run("should generate different webhook secret strings", func(t *testing.T) {
						n := 100
						secrets := make([]string, n)
						for i := 0; i < n; i++ {
							secrets[i] = generatePaCWebhookSecretString()
						}
				
						secret := secrets[0]
						for i := 1; i < n; i++ {
							if secret == secrets[i] {
								t.Errorf("All webhook secrets strings must be different")
							}
						}
					})
				}
				
				func TestGetPathContext(t *testing.T) {
					tests := []struct {
						name              string
						gitContext        string
						dockerfileContext string
						want              string
					}{
						{
							name:              "should return empty context if both contexts empty",
							gitContext:        "",
							dockerfileContext: "",
							want:              "",
						},
						{
							name:              "should use current directory from git context",
							gitContext:        ".",
							dockerfileContext: "",
							want:              ".",
						},
						{
							name:              "should use current directory from dockerfile context",
							gitContext:        "",
							dockerfileContext: ".",
							want:              ".",
						},
						{
							name:              "should use current directory if both contexts are current directory",
							gitContext:        ".",
							dockerfileContext: ".",
							want:              ".",
						},
						{
							name:              "should use git context if dockerfile context if not set",
							gitContext:        "dir",
							dockerfileContext: "",
							want:              "dir",
						},
						{
							name:              "should use dockerfile context if git context if not set",
							gitContext:        "",
							dockerfileContext: "dir",
							want:              "dir",
						},
						{
							name:              "should use git context if dockerfile context is current directory",
							gitContext:        "dir",
							dockerfileContext: ".",
							want:              "dir",
						},
						{
							name:              "should use dockerfile context if git context is current directory",
							gitContext:        ".",
							dockerfileContext: "dir",
							want:              "dir",
						},
						{
							name:              "should respect both git and dockerfile contexts",
							gitContext:        "component-dir",
							dockerfileContext: "dockerfile-dir",
							want:              "component-dir/dockerfile-dir",
						},
						{
							name:              "should respect both git and dockerfile contexts in subfolders",
							gitContext:        "path/to/component",
							dockerfileContext: "path/to/dockercontext/",
							want:              "path/to/component/path/to/dockercontext",
						},
						{
							name:              "should remove slash at the end",
							gitContext:        "path/to/dir/",
							dockerfileContext: "",
							want:              "path/to/dir",
						},
						{
							name:              "should remove slash at the end",
							gitContext:        "",
							dockerfileContext: "path/to/dir/",
							want:              "path/to/dir",
						},
						{
							name:              "should not allow absolute path",
							gitContext:        "/path/to/dir/",
							dockerfileContext: "",
							want:              "path/to/dir",
						},
						{
							name:              "should not allow absolute path",
							gitContext:        "",
							dockerfileContext: "/path/to/dir/",
							want:              "path/to/dir",
						},
						{
							name:              "should not allow absolute path",
							gitContext:        "/path/to/dir1/",
							dockerfileContext: "/path/to/dir2/",
							want:              "path/to/dir1/path/to/dir2",
						},
					}
					for _, tt := range tests {
						t.Run(tt.name, func(t *testing.T) {
							got := getPathContext(tt.gitContext, tt.dockerfileContext)
							if got != tt.want {
								t.Errorf("Expected \"%s\", but got \"%s\"", tt.want, got)
							}
						})
					}
				}
				
				func TestCreateWorkspaceBinding(t *testing.T) {
					tests := []struct {
						name                      string
						pipelineWorkspaces        []tektonapi.PipelineWorkspaceDeclaration
						expectedWorkspaceBindings []tektonapi.WorkspaceBinding
					}{
						{
							name: "should not bind unknown workspaces",
							pipelineWorkspaces: []tektonapi.PipelineWorkspaceDeclaration{
								{
									Name: "unknown1",
								},
								{
									Name: "unknown2",
								},
							},
							expectedWorkspaceBindings: []tektonapi.WorkspaceBinding{},
						},
						{
							name: "should bind git-auth",
							pipelineWorkspaces: []tektonapi.PipelineWorkspaceDeclaration{
								{
									Name: "git-auth",
								},
							},
							expectedWorkspaceBindings: []tektonapi.WorkspaceBinding{
								{
									Name:   "git-auth",
									Secret: &corev1.SecretVolumeSource{SecretName: "{{ git_auth_secret }}"},
								},
							},
						},
						{
							name: "should bind git-auth and workspace, should not bind unknown",
							pipelineWorkspaces: []tektonapi.PipelineWorkspaceDeclaration{
								{
									Name: "git-auth",
								},
								{
									Name: "unknown",
								},
								{
									Name: "workspace",
								},
							},
							expectedWorkspaceBindings: []tektonapi.WorkspaceBinding{
								{
									Name:   "git-auth",
									Secret: &corev1.SecretVolumeSource{SecretName: "{{ git_auth_secret }}"},
								},
								{
									Name:                "workspace",
									VolumeClaimTemplate: generateVolumeClaimTemplate(),
								},
							},
						},
					}
				
					for _, tt := range tests {
						t.Run(tt.name, func(t *testing.T) {
							got := createWorkspaceBinding(tt.pipelineWorkspaces)
							if !reflect.DeepEqual(got, tt.expectedWorkspaceBindings) {
								t.Errorf("Expected %#v, but received %#v", tt.expectedWorkspaceBindings, got)
							}
						})
					}
				}
				
				func TestGetRandomString(t *testing.T) {
					tests := []struct {
						name   string
						length int
					}{
						{
							name:   "should be able to generate one symbol rangom string",
							length: 1,
						},
						{
							name:   "should be able to generate rangom string",
							length: 5,
						},
						{
							name:   "should be able to generate long rangom string",
							length: 100,
						},
					}
					for _, tt := range tests {
						t.Run(tt.name, func(t *testing.T) {
							got := getRandomString(tt.length)
							if len(got) != tt.length {
								t.Errorf("Got string %s has lenght %d but expected length is %d", got, len(got), tt.length)
							}
						})
					}
				}
				t.DeepEqual(got[i], tt.want[i]) {
					t.Errorf("mergeAndSortTektonParams(): got %v, want %v", got, tt.want)
				}
			}
		})
	}
}
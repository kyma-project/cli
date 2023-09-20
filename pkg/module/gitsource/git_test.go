package gitsource

import (
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
)

func Test_fetchRepoURLFromRemotes(t *testing.T) {
	type args struct {
		gitRemotes []*git.Remote
		remoteName string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{name: "Should return correct GitHub URL - Git default format",
			args: args{
				gitRemotes: createTestRemotes(),
				remoteName: "upstream-git-default",
			},
			want:    "https://github.com/kyma-test/test.git",
			wantErr: false,
		},
		{name: "Should return correct GitHub URL -  https format",
			args: args{
				gitRemotes: createTestRemotes(),
				remoteName: "upstream-https",
			},
			want:    "https://github.com/kyma-test/test.git",
			wantErr: false,
		},
		{name: "Should return return error due remote not existing",
			args: args{
				gitRemotes: createTestRemotes(),
				remoteName: "non-existent",
			},
			want:    "",
			wantErr: true,
		},
		{name: "Should return unchanged string -  no scheme format",
			args: args{
				gitRemotes: createTestRemotes(),
				remoteName: "origin",
			},
			want:    "github.com/user-test/test",
			wantErr: false,
		},
		{name: "Should return return error due invalid URL",
			args: args{
				gitRemotes: createTestRemotes(),
				remoteName: "invalidURL",
			},
			want:    "",
			wantErr: true,
		},
		{name: "Should return return error due to empty remotes",
			args: args{
				gitRemotes: []*git.Remote{},
				remoteName: "upstream",
			},
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := fetchRepoURLFromRemotes(tt.args.gitRemotes, tt.args.remoteName)
			if (err != nil) != tt.wantErr {
				t.Errorf("fetchRepoURLFromRemotes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("fetchRepoURLFromRemotes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func createTestRemotes() []*git.Remote {
	return []*git.Remote{
		git.NewRemote(nil, &config.RemoteConfig{
			Name: "upstream-git-default",
			URLs: []string{"git@github.com:kyma-test/test.git"},
		}),
		git.NewRemote(nil, &config.RemoteConfig{
			Name: "upstream-https",
			URLs: []string{"https://github.com/kyma-test/test.git"},
		}),
		git.NewRemote(nil, &config.RemoteConfig{
			Name: "origin",
			URLs: []string{"github.com/user-test/test"},
		}),
		git.NewRemote(nil, &config.RemoteConfig{
			Name: "invalidURL",
			URLs: []string{"\t"},
		}),
	}

}

package buildinfo

import "testing"

func TestNormalizeBuildStatus(t *testing.T) {
	for _, testCase := range []struct {
		name string
		info Info
		want string
	}{
		{name: "development", info: Info{}, want: "development"},
		{name: "release", info: Info{Version: "v1.0.0", Commit: "abc", Status: "release"}, want: "release"},
		{name: "unknown packaged", info: Info{Version: "v1.0.0", Commit: "abc"}, want: "unknown"},
		{name: "dirty overrides", info: Info{Version: "v1.0.0", Commit: "abc", Status: "release", Dirty: true}, want: "dirty"},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got := normalize(testCase.info)
			if got.Status != testCase.want || got.Version == "" || got.Commit == "" {
				t.Fatalf("normalize(%#v)=%#v", testCase.info, got)
			}
		})
	}
}

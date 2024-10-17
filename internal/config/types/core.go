// Package types maintains configuration definitions
package types

type (
	// GitTaggedMode means a repository+download is required to manage
	GitTaggedMode struct {
		Download string   `yaml:"download"`
		Filters  []string `yaml:"filters"`
	}
	// GitHubReleaseMode are github modes operating on releases
	GitHubReleaseMode struct {
		Asset string `yaml:"asset"`
	}
	// GitHubBranchMode will enable a repository+branch to pull a tarball
	GitHubBranchMode struct {
		Name string `yaml:"name"`
	}
	// GitMode enables git-based fetches
	GitMode struct {
		Repository string         `yaml:"repository"`
		Tagged     *GitTaggedMode `yaml:"tagged"`
	}
	// GitHubMode indicates processing of a github project for upstreams
	GitHubMode struct {
		Project string             `yaml:"project"`
		Release *GitHubReleaseMode `yaml:"release"`
		Branch  *GitHubBranchMode  `yaml:"branch"`
	}
	// Application defines how an application is downloaded, unpacked, and deployed
	Application struct {
		Priority   int         `yaml:"priority"`
		Disable    bool        `yaml:"disable"`
		GitHub     *GitHubMode `yaml:"github"`
		Git        *GitMode    `yaml:"git"`
		Extract    Extraction  `yaml:"extract"`
		BuildSteps []Step      `yaml:"build"`
		Deploy     []Artifact  `yaml:"deploy"`
	}
	// Artifact are definitions of what to deploy
	Artifact struct {
		Files       []string `yaml:"files"`
		Destination string   `yaml:"destination"`
	}
	// Step is a build process step
	Step struct {
		Directory string   `yaml:"directory"`
		Command   []string `yaml:"command"`
	}
	// Extraction handles asset extraction
	Extraction struct {
		NoDepth bool     `yaml:"nodepth"`
		Command []string `yaml:"command"`
	}
)

package git

// Status represents the status of a git repository
type Status struct {
	Branch         string   `json:"branch"`
	Staged         []string `json:"staged"`
	Modified       []string `json:"modified"`
	Untracked      []string `json:"untracked"`
	CommitsAhead   int      `json:"commits_ahead"`
	CommitsBehind  int      `json:"commits_behind"`
	Clean          bool     `json:"clean"`
}

// IsClean returns true if the working tree is clean
func (s *Status) IsClean() bool {
	return len(s.Staged) == 0 && len(s.Modified) == 0 && len(s.Untracked) == 0
}

// HasChanges returns true if there are any changes (staged or unstaged)
func (s *Status) HasChanges() bool {
	return len(s.Staged) > 0 || len(s.Modified) > 0
}

// FileChange represents a file change for batch operations
type FileChange struct {
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Mode    string `json:"mode"` // "create", "modify", "delete"
}

// CloneOptions contains options for cloning repositories
type CloneOptions struct {
	Branch string
	Depth  int
	Quiet  bool
}

type CloneOption func(*CloneOptions)

func WithBranch(branch string) CloneOption {
	return func(opts *CloneOptions) {
		opts.Branch = branch
	}
}

func WithDepth(depth int) CloneOption {
	return func(opts *CloneOptions) {
		opts.Depth = depth
	}
}

func WithQuiet(quiet bool) CloneOption {
	return func(opts *CloneOptions) {
		opts.Quiet = quiet
	}
}

// CommitOptions contains options for committing
type CommitOptions struct {
	Author     string
	AllowEmpty bool
	Amend      bool
}

type CommitOption func(*CommitOptions)

func WithAuthor(author string) CommitOption {
	return func(opts *CommitOptions) {
		opts.Author = author
	}
}

func WithAllowEmpty(allowEmpty bool) CommitOption {
	return func(opts *CommitOptions) {
		opts.AllowEmpty = allowEmpty
	}
}

func WithAmend(amend bool) CommitOption {
	return func(opts *CommitOptions) {
		opts.Amend = amend
	}
}

// PushOptions contains options for pushing
type PushOptions struct {
	Remote         string
	Branch         string
	SetUpstream    bool
	Force          bool
	ForceWithLease bool
	DryRun         bool
	Tags           bool
	Delete         bool
}

type PushOption func(*PushOptions)

func WithRemote(remote string) PushOption {
	return func(opts *PushOptions) {
		opts.Remote = remote
	}
}

func WithBranchToPush(branch string) PushOption {
	return func(opts *PushOptions) {
		opts.Branch = branch
	}
}

func WithSetUpstream(setUpstream bool) PushOption {
	return func(opts *PushOptions) {
		opts.SetUpstream = setUpstream
	}
}

func WithForce(force bool) PushOption {
	return func(opts *PushOptions) {
		opts.Force = force
	}
}

func WithForceWithLease(forceWithLease bool) PushOption {
	return func(opts *PushOptions) {
		opts.ForceWithLease = forceWithLease
	}
}

func WithDryRun(dryRun bool) PushOption {
	return func(opts *PushOptions) {
		opts.DryRun = dryRun
	}
}

func WithTags(tags bool) PushOption {
	return func(opts *PushOptions) {
		opts.Tags = tags
	}
}

func WithDelete(delete bool) PushOption {
	return func(opts *PushOptions) {
		opts.Delete = delete
	}
}

// PullOptions contains options for pulling
type PullOptions struct {
	Remote string
	Branch string
	Rebase bool
	FFOnly bool
}

type PullOption func(*PullOptions)

func WithRemoteToPull(remote string) PullOption {
	return func(opts *PullOptions) {
		opts.Remote = remote
	}
}

func WithBranchToPull(branch string) PullOption {
	return func(opts *PullOptions) {
		opts.Branch = branch
	}
}

func WithRebase(rebase bool) PullOption {
	return func(opts *PullOptions) {
		opts.Rebase = rebase
	}
}

func WithFFOnly(ffOnly bool) PullOption {
	return func(opts *PullOptions) {
		opts.FFOnly = ffOnly
	}
}

// BranchOptions contains options for creating branches
type BranchOptions struct {
	StartPoint string
}

type BranchOption func(*BranchOptions)

func WithStartPoint(startPoint string) BranchOption {
	return func(opts *BranchOptions) {
		opts.StartPoint = startPoint
	}
}

// CheckoutOptions contains options for checkout
type CheckoutOptions struct {
	CreateBranch bool
	Force        bool
}

type CheckoutOption func(*CheckoutOptions)

func WithCreateBranch(createBranch bool) CheckoutOption {
	return func(opts *CheckoutOptions) {
		opts.CreateBranch = createBranch
	}
}

func WithForceCheckout(force bool) CheckoutOption {
	return func(opts *CheckoutOptions) {
		opts.Force = force
	}
}

// WorkflowResult represents the result of a high-level git workflow operation
type WorkflowResult struct {
	Success        bool              `json:"success"`
	Message        string            `json:"message"`
	Branch         string            `json:"branch,omitempty"`
	CommitHash     string            `json:"commit_hash,omitempty"`
	FilesChanged   int               `json:"files_changed,omitempty"`
	PRURL          string            `json:"pr_url,omitempty"`
	PRNumber       int               `json:"pr_number,omitempty"`
	Errors         []string          `json:"errors,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}
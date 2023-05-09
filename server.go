package git

import "time"

// UpdateServerInfoOptions represents the available UpdateServerInfo() options.
type UpdateServerInfoOptions struct {
	// Force indicates to overwrite the existing server info.
	Force bool
	// Timeout represents the maximum time in duration that the command is allowed to run.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// UpdateServerInfo updates the server info in the repository.
func UpdateServerInfo(path string, opts ...UpdateServerInfoOptions) ([]byte, error) {
	var opt UpdateServerInfoOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	cmd := NewCommand("update-server-info").AddOptions(opt.CommandOptions)
	if opt.Force {
		cmd.AddArgs("--force")
	}
	cmd.AddArgs(".")
	return cmd.RunInDirWithTimeout(opt.Timeout, path)
}

// ReceivePackOptions represents the available options for ReceivePack().
type ReceivePackOptions struct {
	// Quiet is true for not printing anything to stdout.
	Quiet bool
	// Timeout represents the maximum time in duration that the command is allowed to run.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// HttpBackendInfoRefs indicates generating the info/refs in the
	// http-backend. This is used for smart http.
	HttpBackendInfoRefs bool
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// RecevePack receives the packfile from the client.
func ReceivePack(path string, opts ...ReceivePackOptions) ([]byte, error) {
	var opt ReceivePackOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	cmd := NewCommand("receive-pack").AddOptions(opt.CommandOptions)
	if opt.Quiet {
		cmd.AddArgs("--quiet")
	}
	if opt.HttpBackendInfoRefs {
		cmd.AddArgs("--http-backend-info-refs")
	}
	cmd.AddArgs(".")
	return cmd.RunInDirWithTimeout(opt.Timeout, path)
}

// UploadPackOptions represents the available options for UploadPack().
type UploadPackOptions struct {
	// Quit after a single request/response exchange.
	StatelessRPC bool
	// Do not try <directory>/.git/ if <directory> is no Git directory.
	Strict bool
	// Interrupt transfer after <n> seconds of inactivity.
	// Note: this is different from CommandOptions.Timeout which is the maximum
	// time in duration that the command is allowed to run.
	Timeout time.Duration
	// HttpBackendInfoRefs indicates generating the info/refs in the
	// http-backend. This is used for smart http.
	HttpBackendInfoRefs bool
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// UploadPack uploads the packfile to the client.
func UploadPack(path string, opts ...UploadPackOptions) ([]byte, error) {
	var opt UploadPackOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	cmd := NewCommand("upload-pack").AddOptions(opt.CommandOptions)
	if opt.StatelessRPC {
		cmd.AddArgs("--stateless-rpc")
	}
	if opt.Strict {
		cmd.AddArgs("--strict")
	}
	if opt.Timeout > 0 {
		cmd.AddArgs("--timeout", opt.Timeout.String())
	}
	if opt.HttpBackendInfoRefs {
		cmd.AddArgs("--http-backend-info-refs")
	}
	cmd.AddArgs(".")
	return cmd.RunInDirWithTimeout(opt.Timeout, path)
}

// UploadArchiveOptions represents the available options for UploadArchive().
type UploadArchiveOptions struct {
	// Timeout represents the maximum time in duration that the command is allowed to run.
	//
	// Deprecated: Use CommandOptions.Timeout instead.
	Timeout time.Duration
	// The additional options to be passed to the underlying git.
	CommandOptions
}

// UploadArchive uploads the archive to the client.
func UploadArchive(path string, opts ...UploadArchiveOptions) ([]byte, error) {
	var opt UploadArchiveOptions
	if len(opts) > 0 {
		opt = opts[0]
	}
	cmd := NewCommand("upload-archive").AddOptions(opt.CommandOptions)
	cmd.AddArgs(".")
	return cmd.RunInDirWithTimeout(opt.Timeout, path)
}

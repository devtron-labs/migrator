package main

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"strings"
)

type GitCliUtil struct {
	logger *zap.SugaredLogger
}

func NewGitCliUtilImpl(logger *zap.SugaredLogger) *GitCliUtil {
	return &GitCliUtil{
		logger: logger,
	}
}

const GIT_ASK_PASS = "/git-ask-pass.sh"

func (impl GitCliUtil) Fetch(rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Infow("git fetch ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "fetch", "origin", "--tags", "--force")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Infow("fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl GitCliUtil) Checkout(rootDir string, username string, password string, branchName string) (response, errMsg string, err error) {
	impl.logger.Infow("git checkout ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "checkout", branchName)
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Infow("checkout output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl GitCliUtil) runCommandWithCred(cmd *exec.Cmd, userName, password string) (response, errMsg string, err error) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", GIT_ASK_PASS),
		fmt.Sprintf("GIT_USERNAME=%s", userName),
		fmt.Sprintf("GIT_PASSWORD=%s", password),
	)
	return impl.runCommand(cmd)
}

func (impl GitCliUtil) runCommand(cmd *exec.Cmd) (response, errMsg string, err error) {
	cmd.Env = append(cmd.Env, "HOME=/dev/null")
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		impl.logger.Errorw("error in git cli operation", "msg", string(outBytes), "err", err)
		exErr, ok := err.(*exec.ExitError)
		if !ok {
			return "", string(outBytes), err
		}
		errOutput := string(exErr.Stderr)
		return "", errOutput, err
	}
	output := string(outBytes)
	output = strings.TrimSpace(output)
	return output, "", nil
}

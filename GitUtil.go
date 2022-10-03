package main

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4/config"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

const GIT_ASK_PASS = "/git-ask-pass.sh"

type GitConfig struct {
	Token          string `env:"GIT_AUTH_TOKEN" secretData:"-"`      //not null
	UserName       string `env:"GIT_USER"`                           //not null
	GitWorkingDir  string `env:"GIT_WORKING_DIR" envDefault:"/tmp/"` //working directory for git. might use pvc
	GitRepoUrl     string `env:"GIT_REPO_URL" `
	GitTag         string `env:"GIT_TAG"`
	GitHash        string `env:"GIT_HASH"`
	GitBranch      string `env:"GIT_BRANCH" envDefault:"main"`
	ScriptLocation string `env:"SCRIPT_LOCATION"` //FIXME add usages
}

func (cfg GitConfig) valid() bool {
	//checkNonEmpty(cfg.Token, "GIT_AUTH_TOKEN")
	//checkNonEmpty(cfg.UserName, "GIT_USER")
	checkNonEmpty(cfg.GitWorkingDir, "GIT_WORKING_DIR")
	checkNonEmpty(cfg.GitRepoUrl, "GIT_REPO_URL")
	if cfg.GitHash == "" && cfg.Token == "" {
		log.Panic(fmt.Errorf("hash and and token both are empty"))
	}
	if cfg.GitHash != "" && cfg.GitTag != "" {
		log.Panic(fmt.Errorf("hash and and token both are present"))
	}
	return true
}

func checkNonEmpty(val, key string) {

	if val == "" {
		log.Panic(fmt.Errorf("%s is empty", key))
	}
}
func GetGitConfig() (*GitConfig, error) {
	cfg := &GitConfig{}
	err := env.Parse(cfg)
	return cfg, err
}

type GitService interface {
	CloneAndCheckout2(targetDir string) (clonedDir string, err error)
	CloneAndCheckout(targetDir string) (clonedDir string, err error)
	BuildScriptSource(clonedDir string) string
}
type GitServiceImpl struct {
	Auth   transport.AuthMethod
	config *GitConfig
	logger *zap.SugaredLogger
}

func NewGitServiceImpl(config *GitConfig, logger *zap.SugaredLogger) *GitServiceImpl {
	auth := &http.BasicAuth{Password: config.Token, Username: config.UserName}
	return &GitServiceImpl{
		Auth:   auth,
		logger: logger,
		config: config,
	}
}

func (impl GitServiceImpl) BuildScriptSource(clonedDir string) string {
	return filepath.Join(clonedDir, impl.config.ScriptLocation)
}

func (impl GitServiceImpl) CloneAndCheckout(targetDir string) (clonedDir string, err error) {
	branch := impl.config.GitBranch
	if branch == "" {
		branch = "master"
	}
	workTree, cloneDir, err := impl.Clone(targetDir, branch)
	if err != nil {
		return "", err
	}

	fmt.Println("sleeping after clone")
	time.Sleep(100 * time.Second)

	if impl.config.GitHash != "" {
		err = impl.CheckoutHash(workTree, impl.config.GitHash)
	} else if impl.config.GitTag != "" {
		err = impl.CheckoutTag(workTree, impl.config.GitTag)
	} else {
		return "", fmt.Errorf("neither tag nor hash provided")
	}

	return cloneDir, err
}

func (impl GitServiceImpl) CloneAndCheckout2(targetDir string) (clonedDir string, err error) {
	branch := impl.config.GitBranch
	if branch == "" {
		branch = "master"
	}
	cloneDir, err := impl.Clone2(targetDir, branch)
	if err != nil {
		return "", err
	}

	return cloneDir, err
}

func (impl GitServiceImpl) Clone(targetDir, branch string) (workTree *git.Worktree, clonedDir string, err error) {
	impl.logger.Infow("git checkout ", "url", impl.config.GitRepoUrl, "dir", targetDir, "branch", branch)
	clonedDir = filepath.Join(impl.config.GitWorkingDir, targetDir)
	repo, err := git.PlainClone(clonedDir, false, &git.CloneOptions{
		URL:           impl.config.GitRepoUrl,
		Auth:          impl.Auth,
		ReferenceName: plumbing.ReferenceName(fmt.Sprintf("refs/heads/%s", branch)),
	})
	if err != nil {
		impl.logger.Errorw("error in git checkout ", "url", impl.config.GitRepoUrl, "targetDir", targetDir, "err", err)
		return nil, "", err
	}

	impl.logger.Infow("cloned ", "dir", clonedDir, "source", impl.config.GitRepoUrl)

	fmt.Println("sleeping before Worktree")
	time.Sleep(100 * time.Second)

	w, err := repo.Worktree()
	if err != nil {
		impl.logger.Errorw("error in work tree resolution", "err", err)
		return nil, "", err
	}

	return w, clonedDir, nil
}

func (impl GitServiceImpl) Clone2(targetDir, branch string) (clonedDir string, err error) {
	impl.logger.Infow("git checkout ", "url", impl.config.GitRepoUrl, "dir", targetDir, "branch", branch)
	clonedDir = filepath.Join(impl.config.GitWorkingDir, targetDir)

	fmt.Println("sleeping before init")
	time.Sleep(50 * time.Second)

	err = impl.Init(clonedDir, impl.config.GitRepoUrl, false)
	if err != nil {
		impl.logger.Errorw("error in git init ", "url", impl.config.GitRepoUrl, "targetDir", targetDir, "err", err)
		return "", err
	}

	fmt.Println("sleeping before fetch")
	time.Sleep(50 * time.Second)

	_, _, err = impl.Fetch(clonedDir, "", "")
	if err != nil {
		impl.logger.Errorw("error in git fetch ", "url", impl.config.GitRepoUrl, "targetDir", targetDir, "err", err)
		return "", err
	}

	fmt.Println("sleeping before checkout")
	time.Sleep(50 * time.Second)

	_, _, err = impl.Checkout(clonedDir, "", "", branch)
	if err != nil {
		impl.logger.Errorw("error in git checkout ", "url", impl.config.GitRepoUrl, "targetDir", targetDir, "err", err)
		return "", err
	}

	return clonedDir, nil
}

func (impl GitServiceImpl) CheckoutHash(workTree *git.Worktree, hash string) error {
	impl.logger.Infow("checking out hash ", "hash", hash)
	err := workTree.Checkout(&git.CheckoutOptions{
		Hash:  plumbing.NewHash(hash),
		Force: true,
	})
	return err
}

func (impl GitServiceImpl) CheckoutTag(workTree *git.Worktree, tag string) error {
	err := workTree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewTagReferenceName(tag),
	})
	return err
}

func (impl GitServiceImpl) CommitAndPushAllChanges(repoRoot, commitMsg string) (commitHash string, err error) {
	repo, workTree, err := impl.getRepoAndWorktree(repoRoot)
	if err != nil {
		return "", err
	}
	err = workTree.AddGlob("")
	if err != nil {
		return "", err
	}
	//--  commit
	commit, err := workTree.Commit(commitMsg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Devtron Boat",
			Email: "manifest-boat@devtron.ai",
			When:  time.Now(),
		},
	})
	if err != nil {
		return "", err
	}
	impl.logger.Infow("git hash", "repo", repoRoot, "hash", commit.String())
	//-----------push
	err = repo.Push(&git.PushOptions{
		Auth: impl.Auth,
	})

	return commit.String(), err
}

func (impl GitServiceImpl) getRepoAndWorktree(repoRoot string) (*git.Repository, *git.Worktree, error) {
	r, err := git.PlainOpen(repoRoot)
	if err != nil {
		return nil, nil, err
	}
	w, err := r.Worktree()
	return r, w, err
}

func (impl GitServiceImpl) ForceResetHead(repoRoot string) (err error) {
	_, workTree, err := impl.getRepoAndWorktree(repoRoot)
	if err != nil {
		return err
	}
	err = workTree.Reset(&git.ResetOptions{Mode: git.HardReset})
	if err != nil {
		return err
	}
	err = workTree.Pull(&git.PullOptions{
		Auth:         impl.Auth,
		Force:        true,
		SingleBranch: true,
	})
	return err
}

func (impl GitServiceImpl) Init(rootDir string, remoteUrl string, isBare bool) error {
	//-----------------

	err := os.MkdirAll(rootDir, 0755)
	if err != nil {
		return err
	}
	repo, err := git.PlainInit(rootDir, isBare)
	if err != nil {
		return err
	}
	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: git.DefaultRemoteName,
		URLs: []string{remoteUrl},
	})
	return err
}

func (impl GitServiceImpl) Fetch(rootDir string, username string, password string) (response, errMsg string, err error) {
	impl.logger.Debugw("git fetch ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "fetch", "origin", "--tags", "--force")
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl GitServiceImpl) Checkout(rootDir string, username string, password string, branchName string) (response, errMsg string, err error) {
	impl.logger.Debugw("git checkout ", "location", rootDir)
	cmd := exec.Command("git", "-C", rootDir, "checkout", branchName)
	output, errMsg, err := impl.runCommandWithCred(cmd, username, password)
	impl.logger.Debugw("fetch output", "root", rootDir, "opt", output, "errMsg", errMsg, "error", err)
	return output, errMsg, err
}

func (impl GitServiceImpl) runCommandWithCred(cmd *exec.Cmd, userName, password string) (response, errMsg string, err error) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("GIT_ASKPASS=%s", GIT_ASK_PASS),
		fmt.Sprintf("GIT_USERNAME=%s", userName),
		fmt.Sprintf("GIT_PASSWORD=%s", password),
	)
	return impl.runCommand(cmd)
}

func (impl GitServiceImpl) runCommand(cmd *exec.Cmd) (response, errMsg string, err error) {
	cmd.Env = append(cmd.Env, "HOME=/dev/null")
	outBytes, err := cmd.CombinedOutput()
	if err != nil {
		impl.logger.Error("error in git cli operation", "msg", string(outBytes), "err", err)
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

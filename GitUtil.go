package main

import (
	"fmt"
	"gopkg.in/src-d/go-git.v4/config"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/caarlos0/env"
	"go.uber.org/zap"
	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	"gopkg.in/src-d/go-git.v4/plumbing/transport/http"
)

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
	CloneAndCheckout(targetDir string) (string, error)
	BuildScriptSource(clonedDir string) string
}

type GitServiceImpl struct {
	Auth       transport.AuthMethod
	config     *GitConfig
	logger     *zap.SugaredLogger
	gitCliUtil *GitCliUtil
}

func NewGitServiceImpl(config *GitConfig, logger *zap.SugaredLogger, gitCliUtil *GitCliUtil) *GitServiceImpl {
	auth := &http.BasicAuth{Password: config.Token, Username: config.UserName}
	return &GitServiceImpl{
		Auth:       auth,
		logger:     logger,
		config:     config,
		gitCliUtil: gitCliUtil,
	}
}

func (impl GitServiceImpl) BuildScriptSource(clonedDir string) string {
	return filepath.Join(clonedDir, impl.config.ScriptLocation)
}

func (impl GitServiceImpl) CloneAndCheckout(targetDir string) (string, error) {
	var checkout string
	if impl.config.GitHash != "" {
		checkout = impl.config.GitHash
	} else if impl.config.GitTag != "" {
		checkout = impl.config.GitTag
	} else {
		return "", fmt.Errorf("neither tag nor hash provided")
	}

	clonedDir := filepath.Join(impl.config.GitWorkingDir, targetDir)

	err := impl.Init(clonedDir, impl.config.GitRepoUrl, false)
	if err != nil {
		impl.logger.Errorw("error in git init ", "url", impl.config.GitRepoUrl, "targetDir", targetDir, "err", err)
		return clonedDir, err
	}

	_, _, err = impl.gitCliUtil.Fetch(clonedDir, impl.config.UserName, impl.config.Token)
	if err != nil {
		impl.logger.Errorw("error in git fetch ", "url", impl.config.GitRepoUrl, "targetDir", targetDir, "err", err)
		return clonedDir, err
	}

	_, _, err = impl.gitCliUtil.Checkout(clonedDir, impl.config.UserName, impl.config.Token, checkout)
	if err != nil {
		impl.logger.Errorw("error in git checkout ", "url", impl.config.GitRepoUrl, "targetDir", targetDir, "err", err)
		return clonedDir, err
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

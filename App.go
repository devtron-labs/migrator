package main

import (
	"fmt"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"log"
)

func main() {
	fmt.Println("starting migration")
	app := NewApp()
	fmt.Println("app initialised ")
	cloneDir, err := app.gitService.CloneAndCheckout("app")
	fmt.Println("checkout " + cloneDir)
	checkErr(err)
	scriptSource := app.gitService.BuildScriptSource(cloneDir)
	v, err := app.migrateUtil.Migrate(scriptSource)
	checkErr(err)
	fmt.Printf("migrated to %d", v)

	/*m, err := migrate.New(
		"file:///Users/nishant/go/src/devtron.ai/orchestrator/scripts/sql",
		"postgres://postgres:devtronpg@localhost:5432/migrate_test?sslmode=disable")
	fmt.Println(err)
	v, _, _ := m.Version()
	fmt.Println(v)*/

}

type App struct {
	gitService  GitService
	migrateUtil *MigrateUtil
}

func NewApp() *App {
	gitcfg, err := GetGitConfig()
	checkErr(err)
	valid := gitcfg.valid()
	if !valid {
		log.Fatal("not valid git config")
	}
	fmt.Println("valid git config found")
	fmt.Println(gitcfg)
	logger, err := zap.NewProduction()
	checkErr(err)
	gitService := NewGitServiceImpl(gitcfg, logger.Sugar())
	migrateConfig, err := GetMigrateConfig()
	checkErr(err)
	valid = migrateConfig.Valid()
	if !valid {
		log.Fatal("not valid migrate config")
	}
	fmt.Println("valid migrate config found")
	fmt.Println(migrateConfig)
	migrateUtil := NewMigrateUtil(migrateConfig, logger.Sugar())
	fmt.Println("migrate util created")
	return &App{migrateUtil: migrateUtil, gitService: gitService}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"log"
	"reflect"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func main() {
	fmt.Println("starting migration")
	app := NewApp()
	fmt.Println("app initialised ")
	sourceDir := "/tmp/app/"
	var err error
	if !app.migrateUtil.config.IsScriptsMounted {
		sourceDir, err = app.gitService.CloneAndCheckout("app")
		fmt.Println("checkout " + sourceDir)
		checkErr(err)
	}
	scriptSource := app.gitService.BuildScriptSource(sourceDir)
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

	migrateConfig, err := GetMigrateConfig()
	checkErr(err)
	valid := migrateConfig.Valid()
	if !valid {
		log.Fatal("not valid migrate config")
	}
	logger, err := zap.NewProduction()
	checkErr(err)
	fmt.Printf("valid migrate config found: %v\n", obfuscateSecretTags(migrateConfig))
	migrateUtil := NewMigrateUtil(migrateConfig, logger.Sugar())
	gitcfg := &GitConfig{}
	fmt.Println("migrate util created")
	if !migrateConfig.IsScriptsMounted {
		gitcfg, err = GetGitConfig()
		checkErr(err)
		valid = gitcfg.valid()
		if !valid {
			log.Fatal("not valid git config")
		}
		fmt.Printf("valid git config found %v\n", obfuscateSecretTags(gitcfg))
	}
	gitCliUtil := NewGitCliUtilImpl(logger.Sugar())
	gitService := NewGitServiceImpl(gitcfg, logger.Sugar(), gitCliUtil)
	return &App{migrateUtil: migrateUtil, gitService: gitService}
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func obfuscateSecretTags(cfg interface{}) interface{} {

	cfgDpl := reflect.New(reflect.ValueOf(cfg).Elem().Type()).Interface()
	cfgDplElm := reflect.ValueOf(cfgDpl).Elem()
	t := cfgDplElm.Type()
	for i := 0; i < t.NumField(); i++ {
		if _, ok := t.Field(i).Tag.Lookup("secretData"); ok {
			cfgDplElm.Field(i).SetString("********")
		} else {
			cfgDplElm.Field(i).Set(reflect.ValueOf(cfg).Elem().Field(i))
		}
	}
	return cfgDpl
}

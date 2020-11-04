### MigrateConfig

key|type|description|defaultValue
---|----|-------------|-------------
DB_URL|string|url with auth of db|postgres://user:password@127.0.0.1:5432/migrate_test?sslmode=disable"
MIGRATE_TO_VERSION|int|target version. if zero will point to latest version|0
GIT_AUTH_TOKEN|string|git auth token
GIT_USER|string| username 
GIT_WORKING_DIR|string||/devtron-data/
GIT_REPO_URL|string|https://gitlab.com/devtron-gitops/my-project2.git|""
GIT_TAG|string|either GIT_TAG or GIT_HASH must be present|
GIT_HASH|string|either GIT_TAG or GIT_HASH must be present|
GIT_BRANCH|string|default master|




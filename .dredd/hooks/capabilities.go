package main

import (
	"encoding/json"
	"github.com/ansible-semaphore/semaphore/db"
	trans "github.com/snikch/goodman/transaction"
	"strconv"
	"strings"
)

// STATE
// Runtime created objects we needs to reference in test setups
var testRunnerUser *db.User
var userPathTestUser *db.User
var userProject *db.Project
var userKey *db.AccessKey
var task *db.Task

// Runtime created simple ID values for some items we need to reference in other objects
var repoID int64
var inventoryID int64
var environmentID int64
var templateID int64

var capabilities = map[string][]string{
	"user":        {},
	"project":     {"user"},
	"access_key":  {"project"},
	"repository":  {"access_key"},
	"inventory":   {"repository"},
	"environment": {"repository"},
	"template":    {"repository", "inventory", "environment"},
	"task":		   {"template"},
}

func capabilityWrapper(cap string) func(t *trans.Transaction) {
	return func(t *trans.Transaction) {
		addCapabilities([]string{cap})
	}
}

func addCapabilities(caps []string) {
	dbConnect()
	defer db.Mysql.Db.Close()
	resolved := make([]string, 0)
	uid := getUUID()
	resolveCapability(caps, resolved, uid)
}

func resolveCapability(caps []string, resolved []string, uid string) {
	for _, v := range caps {

		//if cap has deps resolve them
		if val, ok := capabilities[v]; ok {
			resolveCapability(val, resolved, uid)
		}

		//skip if already resolved
		if _, exists := stringInSlice(v, resolved); exists {
			continue
		}

		//Add dep specific stuff
		switch v {
		case "user":
			userPathTestUser = addUser()
		case "project":
			userProject = addProject()
			//allow the admin user (test executor) to manipulate the project
			addUserProjectRelation(userProject.ID, testRunnerUser.ID)
			addUserProjectRelation(userProject.ID, userPathTestUser.ID)
		case "access_key":
			userKey = addAccessKey(&userProject.ID)
		case "repository":
			pRepo, err := db.Mysql.Exec("insert into project__repository set project_id=?, git_url=?, ssh_key_id=?, name=?", userProject.ID, "git@github.com/ansible,semaphore/semaphore", userKey.ID, "ITR-"+uid)
			printError(err)
			repoID, _ = pRepo.LastInsertId()
		case "inventory":
			res, err := db.Mysql.Exec("insert into project__inventory set project_id=?, name=?, type=?, key_id=?, ssh_key_id=?, inventory=?", userProject.ID, "ITI-"+uid, "static", userKey.ID, userKey.ID, "Test Inventory")
			printError(err)
			inventoryID, _ = res.LastInsertId()
		case "environment":
			res, err := db.Mysql.Exec("insert into project__environment set project_id=?, name=?, json=?, password=?", userProject.ID, "ITI-"+uid, "{}", "test-pass")
			printError(err)
			environmentID, _ = res.LastInsertId()
		case "template":
			res, err := db.Mysql.Exec("insert into project__template set ssh_key_id=?, project_id=?, inventory_id=?, repository_id=?, environment_id=?, alias=?, playbook=?, arguments=?, override_args=?", userKey.ID, userProject.ID, inventoryID, repoID, environmentID, "Test-"+uid, "test-playbook.yml", "", false)
			printError(err)
			templateID, _ = res.LastInsertId()
		case "task":
			task = addTask()
		}
		resolved = append(resolved, v)
	}
}

// HOOKS
var skipTest = func(t *trans.Transaction) {
	t.Skip = true
}

// Contains all the substitutions for paths under test
// The parameter example value in the api-doc should respond to the index+1 of the function in this slice
// ie the project id, with example value 1, will be replaced by the return value of pathSubPatterns[0]
var pathSubPatterns = []func() string{
	func() string { return strconv.Itoa(userProject.ID) },
	func() string { return strconv.Itoa(userPathTestUser.ID) },
	func() string { return strconv.Itoa(userKey.ID) },
	func() string { return strconv.Itoa(int(repoID)) },
	func() string { return strconv.Itoa(int(inventoryID)) },
	func() string { return strconv.Itoa(int(environmentID)) },
	func() string { return strconv.Itoa(int(templateID)) },
	func() string { return strconv.Itoa(task.ID) },
}

// alterRequestPath with the above slice of functions
func alterRequestPath(t *trans.Transaction) {
	pathArgs := strings.Split(t.FullPath, "/")
	exploded := make([]string, len(pathArgs))
	copy(exploded, pathArgs)
	for k, v := range pathSubPatterns {
		pos, exists := stringInSlice(strconv.Itoa(k+1), exploded)
		if exists {
			pathArgs[pos] = v()
		}
	}
	t.FullPath = strings.Join(pathArgs, "/")
	t.Request.URI = t.FullPath
}

func alterRequestBody(t *trans.Transaction) {
	var request map[string]interface{}
	json.Unmarshal([]byte(t.Request.Body), &request)

	if userProject != nil {
		bodyFieldProcessor("project_id", userProject.ID, &request)
	}
	bodyFieldProcessor("json", "{}", &request)
	if userKey != nil {
		bodyFieldProcessor("ssh_key_id", userKey.ID, &request)
		bodyFieldProcessor("key_id", userKey.ID, &request)
	}
	bodyFieldProcessor("environment_id", environmentID, &request)
	bodyFieldProcessor("inventory_id", inventoryID, &request)
	bodyFieldProcessor("repository_id", repoID, &request)
	bodyFieldProcessor("template_id", templateID, &request)
	if task != nil {
		bodyFieldProcessor("task_id", task.ID, &request)
	}

	out, _ := json.Marshal(request)
	t.Request.Body = string(out)
}

func bodyFieldProcessor(id string, sub interface{}, request *map[string]interface{}) {
	if _, ok := (*request)[id]; ok {
		(*request)[id] = sub
	}
}

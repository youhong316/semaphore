package db

// Template is a user defined model that is used to run a task
type Template struct {
	ID int `db:"id" json:"id"`

	SSHKeyID      int  `db:"ssh_key_id" json:"ssh_key_id"`
	ProjectID     int  `db:"project_id" json:"project_id"`
	InventoryID   int  `db:"inventory_id" json:"inventory_id"`
	RepositoryID  int  `db:"repository_id" json:"repository_id"`
	EnvironmentID *int `db:"environment_id" json:"environment_id"`

	// Alias as described in https://github.com/ansible-semaphore/semaphore/issues/188
	Alias string `db:"alias" json:"alias"`
	// playbook name in the form of "some_play.yml"
	Playbook string `db:"playbook" json:"playbook"`
	// to fit into []string
	Arguments *string `db:"arguments" json:"arguments"`
	// if true, semaphore will not prepend any arguments to `arguments` like inventory, etc
	OverrideArguments bool `db:"override_args" json:"override_args"`
}

package qb

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"testing"
)

type TableTestSuite struct {
	suite.Suite
}

func (suite *TableTestSuite) TestTableSimpleCreateDrop() {
	dialect := NewDialect("mysql")
	usersTable := Table("users", Column("id", Varchar().Size(40)))

	ddl := usersTable.Create(dialect)
	fmt.Println(ddl, "\n")
	assert.Equal(suite.T(), ddl, "CREATE TABLE users (\n\tid VARCHAR(40)\n);")

	statement := usersTable.Build(dialect)
	assert.Equal(suite.T(), statement.SQL(), "CREATE TABLE users (\n\tid VARCHAR(40)\n);")
	assert.Equal(suite.T(), statement.Bindings(), []interface{}{})

	assert.Equal(suite.T(), usersTable.Drop(dialect), "DROP TABLE users;")
}

func (suite *TableTestSuite) TestTablePrimaryForeignKey() {
	usersTable := Table(
		"users",
		Column("id", Varchar().Size(40)),
		Column("session_id", Varchar().Size(40)),
		Column("auth_token", Varchar().Size(40)),
		Column("role_id", Varchar().Size(40)),
		PrimaryKey("id"),
		ForeignKey().
			Ref("session_id", "sessions", "id").
			Ref("auth_token", "sessions", "auth_token").
			Ref("role_id", "roles", "id"),
	)

	ddl := usersTable.Create(NewDialect("mysql"))
	fmt.Println(ddl, "\n")
	assert.Contains(suite.T(), ddl, "CREATE TABLE users (")
	assert.Contains(suite.T(), ddl, "auth_token VARCHAR(40)")
	assert.Contains(suite.T(), ddl, "role_id VARCHAR(40)")
	assert.Contains(suite.T(), ddl, "id VARCHAR(40)")
	assert.Contains(suite.T(), ddl, "session_id VARCHAR(40)")
	assert.Contains(suite.T(), ddl, "PRIMARY KEY(id)")
	assert.Contains(suite.T(), ddl, "FOREIGN KEY(session_id, auth_token) REFERENCES sessions(id, auth_token)")
	assert.Contains(suite.T(), ddl, "FOREIGN KEY(role_id) REFERENCES roles(id)")
	assert.Contains(suite.T(), ddl, ");")
}

func (suite *TableTestSuite) TestTableUniqueCompositeUnique() {
	usersTable := Table(
		"users",
		Column("id", Varchar().Size(40)),
		Column("email", Varchar().Size(40).Unique()),
		Column("device_id", Varchar().Size(255).Unique()),
		UniqueKey("email", "device_id"),
	)

	ddl := usersTable.Create(NewDialect("mysql"))
	fmt.Println(ddl, "\n")
	assert.Contains(suite.T(), ddl, "CREATE TABLE users (")
	assert.Contains(suite.T(), ddl, "id VARCHAR(40)")
	assert.Contains(suite.T(), ddl, "email VARCHAR(40) UNIQUE")
	assert.Contains(suite.T(), ddl, "device_id VARCHAR(255) UNIQUE")
	assert.Contains(suite.T(), ddl, "CONSTRAINT u_email_device_id UNIQUE(email, device_id)")
	assert.Contains(suite.T(), ddl, ");")
}

func (suite *TableTestSuite) TestTableIndex() {
	usersTable := Table(
		"users",
		Column("id", Varchar().Size(40)),
		Column("email", Varchar().Size(40).Unique()),
		Index("users", "id"),
		Index("users", "email"),
		Index("users", "id", "email"),
	)
	ddl := usersTable.Create(NewDialect("postgres"))
	fmt.Println(ddl, "\n")
	assert.Contains(suite.T(), ddl, "CREATE TABLE users (")
	assert.Contains(suite.T(), ddl, "id VARCHAR(40)")
	assert.Contains(suite.T(), ddl, "email VARCHAR(40) UNIQUE")
	assert.Contains(suite.T(), ddl, ")")
	assert.Contains(suite.T(), ddl, "CREATE INDEX i_id ON users(id)")
	assert.Contains(suite.T(), ddl, "CREATE INDEX i_email ON users(email)")
	assert.Contains(suite.T(), ddl, "CREATE INDEX i_id_email ON users(id, email);")

	assert.Equal(suite.T(), usersTable.C("id"), ColumnElem{Name: "id", Type: Varchar().Size(40), Table: "users"})
	assert.Zero(suite.T(), usersTable.C("nonExisting"))
}

func (suite *TableTestSuite) TestTableIndexChain() {
	usersTable := Table("users", Column("id", Varchar().Size(40))).Index("id")
	ddl := usersTable.Create(NewDialect("mysql"))
	fmt.Println(ddl, "\n")
	assert.Equal(suite.T(), ddl, "CREATE TABLE users (\n\tid VARCHAR(40)\n);\nCREATE INDEX i_id ON users(id);")
}

func (suite *TableTestSuite) TestTableStarters() {
	users := Table(
		"users",
		Column("id", Varchar().Size(40)),
		Column("email", Varchar().Size(40).Unique()),
		PrimaryKey("id"),
	)

	sqlite := NewDialect("sqlite3")

	ins := users.
		Insert().
		Values(map[string]interface{}{
			"id":    "5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55",
			"email": "al@pacino.com",
		}).
		Build(sqlite)

	assert.Contains(suite.T(), ins.SQL(), "INSERT INTO users")
	assert.Contains(suite.T(), ins.SQL(), "id")
	assert.Contains(suite.T(), ins.SQL(), "email")
	assert.Contains(suite.T(), ins.SQL(), "VALUES(?, ?)")
	assert.Contains(suite.T(), ins.Bindings(), "5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55")
	assert.Contains(suite.T(), ins.Bindings(), "al@pacino.com")

	ups := users.Upsert().
		Values(map[string]interface{}{
			"id":    "5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55",
			"email": "al@pacino.com",
		}).
		Build(sqlite)

	assert.Contains(suite.T(), ups.SQL(), "REPLACE INTO users")
	assert.Contains(suite.T(), ups.SQL(), "id")
	assert.Contains(suite.T(), ups.SQL(), "email")
	assert.Contains(suite.T(), ups.SQL(), "VALUES(?, ?)")
	assert.Contains(suite.T(), ups.Bindings(), "5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55")
	assert.Contains(suite.T(), ups.Bindings(), "al@pacino.com")

	upd := users.
		Update().
		Values(map[string]interface{}{
			"email": "al@pacino.com",
		}).
		Where(users.C("id").Eq("5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55")).
		Build(sqlite)

	assert.Equal(suite.T(), upd.SQL(), "UPDATE users\nSET email = ?\nWHERE users.id = ?;")
	assert.Equal(suite.T(), upd.Bindings(), []interface{}{"al@pacino.com", "5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55"})

	del := users.
		Delete().
		Where(users.C("id").Eq("5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55")).
		Build(sqlite)

	assert.Equal(suite.T(), del.SQL(), "DELETE FROM users\nWHERE users.id = ?;")
	assert.Equal(suite.T(), del.Bindings(), []interface{}{"5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55"})

	sel := users.
		Select(users.C("id"), users.C("email")).
		Where(users.C("id").Eq("5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55")).
		Build(sqlite)

	assert.Equal(suite.T(), sel.SQL(), "SELECT id, email\nFROM users\nWHERE users.id = ?;")
	assert.Equal(suite.T(), sel.Bindings(), []interface{}{"5a73ef89-cf0a-4c51-ab8c-cc273ebb3a55"})
}

func TestTableTestSuite(t *testing.T) {
	suite.Run(t, new(TableTestSuite))
}

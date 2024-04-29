package model

import (
	"os"
	"testing"

	"github.com/gagasdiv/cdule/pkg"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/require"
)

func Test_ConnectDatabase(t *testing.T) {
	config := pkg.NewDefaultConfig()
	ConnectDataBase(config)
	require.NotEqual(t, pkg.EMPTYSTRING, config.Dburl)
}

func Test_ConnectDatabaseFailedToReadConfig(t *testing.T) {
	recovered := false
	defer func() {
		if r := recover(); r != nil {
			log.Warning("Recovered in Test_ConnectPostgresDBPanic ", r)
			recovered = true
		}
	}()
	config := pkg.NewDefaultConfig()
	ConnectDataBase(config)
	require.EqualValues(t, true, recovered)
}

func Test_ConnectPostgresDB(t *testing.T) {
	db := postgresConn("postgres://cduleuser:cdulepassword@localhost:5432/cdule?sslmode=disable")
	require.NotNil(t, db)
}

func Test_ConnectPostgresDBPanic(t *testing.T) {
	recovered := false
	defer func() {
		if r := recover(); r != nil {
			log.Warning("Recovered in Test_ConnectPostgresDBPanic ", r)
			recovered = true
		}
	}()
	db := postgresConn("postgres://abc:abc@localhost:5432/cdule?sslmode=disable")
	require.Nil(t, db)
	require.EqualValues(t, true, recovered)
}

func Test_ConnectSqlite(t *testing.T) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	_ = os.Remove(dirname + "/sqlite.db")

	db := sqliteConn(dirname + "/sqlite.db")
	require.NotNil(t, db)
}

func Test_ConnectSqliteDBPanic(t *testing.T) {
	recovered := false
	defer func() {
		if r := recover(); r != nil {
			log.Warning("Recovered in Test_ConnectSqliteDBPanic ", r)
			recovered = true
		}
	}()
	dirname, err := os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
	_ = os.Remove(dirname + "/sqlite.db")

	db := sqliteConn("///")
	require.Nil(t, db)
	require.EqualValues(t, true, recovered)
}

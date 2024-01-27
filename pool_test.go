package cpool

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var dbName = "mysql"

func TestDefault(t *testing.T) {

	con1, errcon1 := Pool.GetConnection(dbName)
	assert.NoError(t, errcon1)

	assert.Equal(t, 1, len(Pool.InnerPool)) // Pool should contain 1 connection

	con2, errcon2 := Pool.GetConnection(dbName)
	assert.NoError(t, errcon2)

	assert.Equal(t, 2, len(Pool.InnerPool)) // Pool should now contain 2 connections for the same database

	// Second connection should return the same as first connection
	assert.Equal(t, con1, con2)

	_, errquery := con2.Query("select 1")
	assert.NoError(t, errquery)
}

func TestCustom(t *testing.T) {
	Pool.Init(10, 2, 300)

	// Create four connections, this should fill the pool up to maxConnectionsPerDatabase
	Pool.GetConnection(dbName)
	Pool.GetConnection(dbName)
	Pool.GetConnection(dbName)
	Pool.GetConnection(dbName)

	assert.Equal(t, 2, len(Pool.InnerPool)) // Pool should only contain 2 connections for the same database
}

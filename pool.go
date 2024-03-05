package cpool

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/rs/zerolog/log"
)

type PoolItem struct {
	Connection *sql.DB
	Time       time.Time
	Name       string
}

type ConnectionPool struct {
	sync.Mutex
	InnerPool                 []PoolItem
	MaxSize                   int
	MaxConnectionsPerDatabase int
	TTL                       time.Duration
}

var (
	Pool = &ConnectionPool{
		InnerPool:                 []PoolItem{},
		MaxSize:                   100,
		MaxConnectionsPerDatabase: 5,
		TTL:                       300 * time.Second,
	}
)

func (c *ConnectionPool) Init(maxSize int, maxConnectionsPerDatabase int, ttl int) {
	Pool.MaxSize = maxSize
	Pool.MaxConnectionsPerDatabase = maxConnectionsPerDatabase
	Pool.TTL = time.Duration(ttl) * time.Second
}

func (c *ConnectionPool) GetConnection(dbName string) (*sql.DB, error) {
	c.Lock()
	defer c.Unlock()
	defer c.cleanupExpiredConnections()

	indexes := []int{}
	for i, c := range c.InnerPool {
		if c.Name == dbName {
			indexes = append(indexes, i)
		}
	}

	if len(indexes) > 0 {
		if len(indexes) < c.MaxConnectionsPerDatabase {
			log.Trace().Msg("Adding new connection for existing database")
			defer c.newConnection(dbName)
		}

		oldest := &c.InnerPool[indexes[0]]
		for i := range indexes {
			if c.InnerPool[indexes[i]].Time.Before(oldest.Time) {
				oldest = &c.InnerPool[indexes[i]]
			}
		}
		log.Trace().Msg("Returning existing connection")
		oldest.Time = time.Now()
		return oldest.Connection, nil
	}

	log.Trace().Msg("Creating new connection")
	conn, err := c.newConnection(dbName)
	if err != nil {
		log.Error().Err(err).Msg("Error creating new connection")
	}
	return conn.Connection, nil

}

func (c *ConnectionPool) cleanupExpiredConnections() {
	for i := 0; i < len(c.InnerPool); i++ {
		if time.Since(c.InnerPool[i].Time) > c.TTL {
			log.Trace().Msg("Removing expired connection from pool")
			c.InnerPool[i].Connection.Close()
			c.InnerPool = append(c.InnerPool[:i], c.InnerPool[i+1:]...)
			i--
		}
	}

	if len(c.InnerPool) > c.MaxSize {
		log.Trace().Msg("Reducing pool size to max")
		c.InnerPool = c.InnerPool[(len(c.InnerPool) - c.MaxSize):]
	}

	for i := 0; i < len(c.InnerPool); i++ {
		c.InnerPool[i].Connection.Ping()
	}
}

func (c *ConnectionPool) newConnection(dbName string) (*PoolItem, error) {
	host := getEnv("MYSQL_HOST", "localhost")
	port := getEnv("MYSQL_PORT", "3306")
	username := getEnv("MYSQL_USERNAME", "user")
	password := getEnv("MYSQL_PASSWORD", "password")

	conn, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&tls=skip-verify", username, password, host, port, dbName))
	if err != nil {
		return nil, err
	}
	pi := PoolItem{
		Connection: conn,
		Time:       time.Now(),
		Name:       dbName,
	}
	c.InnerPool = append(c.InnerPool, pi)
	return &pi, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

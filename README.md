# A multi tenant connection pool for your lambdas

A simple connection pool for multi tenant lambas using the silo or bridged model. Used correctly this pool will persist between invokations of lambas to reuse the database connections.

The intented use is for those that are stuck with RDS and not able to use RDS Proxy for various reasons.

## Usage

```go
import cpool "github.com/alexanderwink/lambda-multi-tenant-connection-pool"

// --- //

db, err := cpool.Pool.GetConnection(databaseName)
```

### Defaults

|                           |     |                                                                            |
| ------------------------- | --- | -------------------------------------------------------------------------- |
| MaxSize                   | 100 | Max total number of connections for the pool                               |
| MaxConnectionsPerDatabase | 5   | Max connections in pool for each database                                  |
| TTL                       | 300 | Connections will be removed from the pool when their TTL has been exceeded |

## Custom configuration

```go
cpool.Pool.Init(100, 5, 300)
```

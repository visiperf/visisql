# PostgreSQL service for Golang #


## Install ##

Use `go get` to install this package.

    go get github.com/visiperf/visisql


## Usage ##

The following examples will be based on this PostgreSQL table named `company` :

|  id  |  name  |
|------|--------|
| 1 | Company 1 |
| 2 | Company 2 |

and this structure :

```go
type Company struct {
    Id   int64  `sql:"id"`
    Name string `sql:"name"`
}
```

### Select one row ###

Here is an example to demonstrate how to select the company with id = 1 :

```go
db, err := sql.Open(...)

var fields = []string{"c.id", "c.name"}

var from = "company c"

var joins = []*visisql.Join{
    visisql.NewJoin("user u", "u.company_id = c.id")
}

var where = [][]*visisql.Predicate{{
    visisql.NewPredicate("c.id", visisql.OperatorEqual, []interface{}{1})
}}

var groupBy = []string{"c.id"}

/*

SQL equivalent :

select c.id, c.name
from company c
    inner join user u on u.company_id = c.id
where c.id = 1
group by c.id

*/

var company Company

err := visisql.NewSelectService(db).Get(fields, from, joins, where, groupBy, &company)

// company.Id -> 1
// company.Name -> "Company 1"
```

### Select multiple rows ###

### Insert ###

### Update ###

### Delete ###


## References ###

* SQL builder for Go : [github.com/huandu/go-sqlbuilder](https://github.com/huandu/go-sqlbuilder)
* mapstructure : [github.com/mitchellh/mapstructure](https://github.com/mitchellh/mapstructure)

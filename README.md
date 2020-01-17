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

## FAQ

- Why `predicates` params is always typed as `[][]*visisql.Predicate` ?

> `predicates` params is two dimentional slice to be able to make request with AND / OR operators.
>
> The first dimension will join predicates with `AND` operator, while second dimension will join predicates with `OR` operator.
>
>  
>
> example with `AND` :
>
> ```go
> [][]*visisql.Predicate{{
>   visisql.NewPredicate("c.name", visisql.OperatorLike, []interface{}{"%@visiperf.io"})
> }, {
>   visisql.NewPredicate("u.id", visisql.OperatorEqual, []interface{}{1})
> }}
> ```
>
> SQL equivalent :
>
> ```sql
> where c.name like '%@visiperf.io' and u.id = 1
> ```
>
>  
>
> example with `OR` :
>
> ```go
> [][]*visisql.Predicate{{
>   visisql.NewPredicate("c.id", visisql.OperatorEqual, []interface{}{1}),
>   visisql.NewPredicate("c.name", visisql.OperatorEqual, []interface{}{"Visiperf"}),
> }}
> ```
>
> SQL equivalent :
>
> ```sql
> where c.id = 1 or c.name = 'Visiperf'
> ```
>
>  
>
> Of course, you can mix `AND` and `OR` operators in same request.
>
> example :
>
> ```go
> [][]*visisql.Predicate{{
>   visisql.NewPredicate("c.id", visisql.OperatorEqual, []interface{}{1}),
>   visisql.NewPredicate("c.name", visisql.OperatorEqual, []interface{}{"Visiperf"}),
> }, {
>   visisql.NewPredicate("u.id", visisql.OperatorEqual, []interface{}{1})
> }}
> ```
>
> SQL equivalent :
>
> ```sql
> where (c.id = 1 or c.name = 'Visiperf') and u.id = 1
> ```

## References ###

* SQL builder for Go : [github.com/huandu/go-sqlbuilder](https://github.com/huandu/go-sqlbuilder)
* mapstructure : [github.com/mitchellh/mapstructure](https://github.com/mitchellh/mapstructure)

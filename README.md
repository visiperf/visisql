# PostgreSQL service for Golang #

Table of contents
=================

  * [Install](#install)
  * [Usage](#usage)
  * [FAQ](#faq)
  * [References](#references)

## Install ##

Use `go get` to install this package.

    go get github.com/visiperf/visisql


## Usage ##

The following examples will be based on this PostgreSQL table named `company` :

|  id  |  name  |
|------|--------|
| 1 | Company 1 |
| 2 | Company 2 |
| 3 | Company 3 |

and this structure :

```go
type Company struct {
    Id   int64  `sql:"id"`
    Name string `sql:"name"`
}
```

### Select one row ###

Here is an example to demonstrate how to select the company with `id = 1` :

```go
db, err := sql.Open(...)

var fields = []string{"c.id", "c.name"}

var from = "company c"

var joins = []*visisql.Join{
    visisql.NewJoin("user u", "u.company_id = c.id"),
}

var where = [][]*visisql.Predicate{{
    visisql.NewPredicate("c.id", visisql.OperatorEqual, []interface{}{1}),
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

Here is an example to demonstrate how to select the 2 first companies starting with `Company` :

```go
db, err := sql.Open(...)

var fields = []string{"c.id", "c.name"}

var from = "company c"

var joins = []*visisql.Join{
    visisql.NewJoin("user u", "u.company_id = c.id"),
}

var where = [][]*visisql.Predicate{{
    visisql.NewPredicate("c.name", visisql.OperatorLike, []interface{}{"Company%"}),
}}

var groupBy = []string{"c.id"}

var orderBy = []*visisql.OrderBy{
  	visisql.NewOrderBy("c.id", visisql.OrderAsc),
}

var pagination = visisql.NewPagination(0, 2)

/*

SQL equivalent :

select c.id, c.name
from company c
    inner join user u on u.company_id = c.id
where c.name like 'Company%'
group by c.id
order by c.id ASC
limit 2

*/

var companies []*Company

c, tc, pc, err := visisql.NewSelectService(db).List(fields, from, joins, where, groupBy, orderBy, pagination, &companies)

/*

companies -> [{
	company.Id -> 1
	company.Name -> "Company 1"
}, {
	company.Id -> 2
	company.Name -> "Company 2"
}]

c (count) -> 2 (number of rows including limit)
tc (total count) -> 3 (number of rows if limit was 0)
pc (page count) -> 2 (number of pages, with c elements per page)

*/
```

### Insert ###

Here is an example to demonstrate how to insert a company in database :

```go
db, err := sql.Open(...)

ts, err := visisql.NewTransactionService(db)

cId, err := ts.Insert("company", map[string]interface{}{"name": "Company 4"}, "id")
// company is not saved in database yet (see sql transaction for more informations)
// if an error is occured, rollback is automatically applied to transaction

/*

SQL equivalent :

insert into company (name) 
values ('Company 4')
returning id

*/

// cId -> 4 because `returning` params is set to `id`. You can set what you want, including `nil` if you don't need returned value.

err = ts.Commit() // all requests made with ts are executed now, Company 4 is now in database
```

### Insert Multiple

Here is an example to demonstrate how to insert multiple companies in database :

```go
db, err := sql.Open(...)

var into = "company"

var fields = []string{"name"}

// values must be in same order as fields
var values = [][]interface{}{{"Company 5"}, {"Company 6"}}

var returning = "id"

/*

SQL equivalent (using prepared statement) :

insert into company (name) 
values ('Company 5')
returning id

insert into company (name) 
values ('Company 6')
returning id

*/

ts, err := visisql.NewTransactionService(db)

ids, err := ts.InsertMultiple(into, fields, values, returning)
// companies are not saved in database yet (see sql transaction for more informations)
// if an error is occured, rollback is automatically applied to transaction

err = ts.Commit()

// ids -> [5, 6] because `returning` params is set to `id`. You can set what you want, including `nil` if you don't need returned value.
```



### Update ###

### Delete ###

## FAQ

- Why `predicates` params is always typed as `[][]*visisql.Predicate` ?

> `predicates` params is two dimentional slice to be able to make request with AND / OR operators.
>
> The first dimension will join predicates with `AND` operator, while second dimension will join predicates with `OR` operator.
>
> <br/>
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
> <br/>
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
> <br/>
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

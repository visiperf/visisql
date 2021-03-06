# PostgreSQL services for Golang #

Package `visisql` provide two services `Select` and `Transaction` to help making PostgreSQL requests. `Select` service will assist you to build and scan queries, while `Transaction` will help you to insert, update and delete rows in your databases with SQL transactions (rollback, commit ...).

Visisql use `github.com/huandu/go-sqlbuilder` to build SQL queries and use `github.com/jmoiron/sqlx` to map result with structs. For more informations, see documentations of each package.

Table of contents
=================

  * [Install](#install)
  * [Usage](#usage)
    * [Select one row](#select-one-row)
    * [Select multiple rows](#select-multiple-rows)
    * [Insert](#insert)
    * [Insert multiple](#insert-multiple)
    * [Update](#update)
    * [Delete](#delete)
  * [FAQ](#faq)
  * [References](#references)

## Install ##

Use `go get` to install this package.

    go get github.com/visiperf/visisql/v3


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
    Id   int64  `db:"id"`
    Name string `db:"name"`
}
```

### Select one row ###

Here is an example to demonstrate how to select the company with `id = 1` :

```go
db, err := sqlx.Connect(...)

var fields = []string{"c.id", "c.name"}

var from = "company c"

var joins = []*visisql.Join{
    visisql.NewJoin(visisql.InnerJoin, "user u", "u.company_id = c.id"),
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
db, err := sqlx.Connect(...)

var fields = []string{"c.id", "c.name"}

var from = "company c"

var joins = []*visisql.Join{
    visisql.NewJoin(visisql.InnerJoin, "user u", "u.company_id = c.id"),
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

c, tc, pc, err := visisql.NewSelectService(db).Search(fields, from, joins, where, groupBy, orderBy, pagination, &companies)

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
db, err := sqlx.Connect(...)

var into = "company"

var values = map[string]interface{}{"name": "Company 4"}

var returning = "id"

/*

SQL equivalent :

insert into company (name) 
values ('Company 4')
returning id

*/

ts, err := visisql.NewTransactionService(db)

cId, err := ts.Insert(into, values, returning)
// company is not saved in database yet (see sql transaction for more informations)
// if an error is occured, rollback is automatically applied to transaction

err = ts.Commit() // all requests made with ts are executed now, Company 4 is now in database

// cId -> 4 because `returning` params is set to `id`. You can set what you want, including `nil` if you don't need returned value.
```

### Insert multiple

Here is an example to demonstrate how to insert multiple companies in database :

```go
db, err := sqlx.Connect(...)

var into = "company"

var fields = []string{"name"}

// values must be in same order as fields
var values = [][]interface{}{{"Company 4"}, {"Company 5"}}

var returning = "id"

/*

SQL equivalent (using prepared statement) :

insert into company (name) 
values ('Company 4')
returning id

insert into company (name) 
values ('Company 5')
returning id

*/

ts, err := visisql.NewTransactionService(db)

ids, err := ts.InsertMultiple(into, fields, values, returning)
// companies are not saved in database yet (see sql transaction for more informations)
// if an error is occured, rollback is automatically applied to transaction

err = ts.Commit() // all requests made with ts are executed now, Companies 4 & 5 are now in database

// ids -> [4, 5] because `returning` params is set to `id`. You can set what you want, including `nil` if you don't need returned value.
```

### Update ###

Here is an example to demonstrate how to update the company with `id = 3` :

```go
db, err := sqlx.Connect(...)

var table = "company"

var set = map[string]interface{}{"name": "Company 4"}

var where = [][]*visisql.Predicate{{
    visisql.NewPredicate("id", visisql.OperatorEqual, []interface{}{3}),
}}

/*

SQL equivalent :

update company
set name = 'Company 4'
where id = 3

*/

ts, err := visisql.NewTransactionService(db)

err = ts.Update(table, set, where)
// company is not updated in database yet (see sql transaction for more informations)
// if an error is occured, rollback is automatically applied to transaction

err = ts.Commit() // all requests made with ts are executed now, Company 3 is now updated to Company 4
```

### Delete ###

Here is an example to demonstrate how to delete the company with `id = 3` :

```go
db, err := sqlx.Connect(...)

var from = "company"

var where = [][]*visisql.Predicate{{
    visisql.NewPredicate("id", visisql.OperatorEqual, []interface{}{3}),
}}

/*

SQL equivalent :

delete from company
where id = 3

*/

ts, err := visisql.NewTransactionService(db)

err = ts.Delete(from, where)
// company is not deleted in database yet (see sql transaction for more informations)
// if an error is occured, rollback is automatically applied to transaction

err = ts.Commit() // all requests made with ts are executed now, Company 3 is now deleted
```

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
* SQLX library : [github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx)


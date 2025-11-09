# gadmin
`gadmin` solves some boring and trivial tasks. Write a short piece of code to generate a standard, user-friendly admin interface for controlling your database.

Write some code like:
```go
mv := gadmin.NewModelView(User{}).
    SetColumnList("type", "first_name", "last_name", "email", "timezone", "phone_number").
    SetColumnEditableList("first_name", "type", "timezone").
    SetColumnDescriptions(map[string]string{"type": "editor will cause ..."}).
    SetCanSetPageSize(true).
    SetPageSize(5).
    SetTablePrefixHtml(`<h4>Caution ...</h4>`)
admin.AddView(mv)
```

![Demo](screenshot.png)

Generate code for all tables
git clone ....
cd 
go run ./cmd
Open http://127.0.0.1:3333/admin/generate input database url like:
- sqlite:path/to/sqlite.db
- postgresql://user:password@localhost:5432/database
- mysql://user:password@localhost:3366/database?charset=utf8mb4&parseTime=True&loc=Local

Then click 'Generate'



## How does it work?
The basic concept behind `gadmin`, is that it lets you build complicated interfaces by grouping individual views together in classes: Each web page you see on the frontend, represents a method on a class that has explicitly been added to the interface.

Inspired by `Flask-Admin`. Copy the template/javascript files from flask-admin(bootstrap4 only).

We do more then flask-admin, with `gadmin` we can auto generate models and add them as plugin

## Features

## Why not generate code?
Less code means less work.
It's easy to generate, but the code will be your responsibility forever.

All table relations and field types are standard.
Why don't you use `gadmin`?


TODO:
https://github.com/xo/usql
https://github.com/smallnest/gen
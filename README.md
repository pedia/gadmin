# gadmin
`gadmin` solves some boring and trivial tasks. Write a short piece of code to generate a standard, user-friendly admin interface for controlling your database.

Write some code like:
```go
mv := gadmin.NewModelView(User{}).
    SetColumnList("type", "first_name", "last_name", "email", "ip_address", "currency", "timezone", "phone_number").
    SetColumnEditableList("first_name", "type", "currency", "timezone").
    SetColumnDescriptions(map[string]string{"first_name": "First Name"}).
    SetCanSetPageSize(true).
    SetPageSize(5).
    SetTablePrefixHtml(`<h4>Some Caution</h4>`)
admin.AddView(mv)
```

![Demo](screenshot.png)

## How does it work?
The basic concept behind `gadmin`, is that it lets you build complicated interfaces by grouping individual views together in classes: Each web page you see on the frontend, represents a method on a class that has explicitly been added to the interface.

Inspired by `Flask-Admin`. Copy the template/javascript files from flask-admin(bootstrap4 only).

## Features

## Why don't generate code?
Less code means less work.
It's easy to generate, but the code will be your responsibility forever.
All table relations and field types are standard.
Why don't you use `gadmin`?